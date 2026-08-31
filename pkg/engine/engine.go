package engine

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"apimock/pkg/auth"
	"apimock/pkg/models"
	"apimock/pkg/openapi"
	"apimock/pkg/store"
	"apimock/pkg/template"
	"apimock/pkg/traffic"
)

// Server is the central MockForge mock engine and HTTP server
type Server struct {
	mu             sync.RWMutex
	endpoints      map[string]*models.Endpoint
	templateEngine *template.Engine
	store          *store.CollectionStore
	trafficLogger  *traffic.TrafficLogger
	corsEnabled    bool
	proxyTarget    string
}

// NewServer initializes the MockForge server
func NewServer(storePath string) *Server {
	srv := &Server{
		endpoints:      make(map[string]*models.Endpoint),
		templateEngine: template.NewEngine(),
		store:          store.NewCollectionStore(storePath),
		trafficLogger:  traffic.NewTrafficLogger(300),
		corsEnabled:    true,
	}

	return srv
}

// GetStore returns the collection store
func (s *Server) GetStore() *store.CollectionStore {
	return s.store
}

// GetTrafficLogger returns the traffic logger
func (s *Server) GetTrafficLogger() *traffic.TrafficLogger {
	return s.trafficLogger
}

// Endpoints CRUD
func (s *Server) AddEndpoint(ep models.Endpoint) (*models.Endpoint, error) {
	if ep.ID == "" {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(900000))
		ep.ID = fmt.Sprintf("ep_%d", 100000+nBig.Int64())
	}
	if err := ep.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	if ep.CreatedAt.IsZero() {
		ep.CreatedAt = now
	}
	ep.UpdatedAt = now

	s.mu.Lock()
	defer s.mu.Unlock()

	s.endpoints[ep.ID] = &ep
	return &ep, nil
}

func (s *Server) UpdateEndpoint(id string, ep models.Endpoint) (*models.Endpoint, error) {
	if err := ep.Validate(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.endpoints[id]
	if !exists {
		return nil, errors.New("endpoint not found")
	}

	ep.ID = id
	ep.CreatedAt = existing.CreatedAt
	ep.UpdatedAt = time.Now()

	s.endpoints[id] = &ep
	return &ep, nil
}

func (s *Server) DeleteEndpoint(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.endpoints[id]; !exists {
		return errors.New("endpoint not found")
	}
	delete(s.endpoints, id)
	return nil
}

func (s *Server) GetEndpoint(id string) (*models.Endpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ep, exists := s.endpoints[id]
	return ep, exists
}

func (s *Server) ListEndpoints() []models.Endpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]models.Endpoint, 0, len(s.endpoints))
	for _, ep := range s.endpoints {
		res = append(res, *ep)
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Path == res[j].Path {
			return res[i].Method < res[j].Method
		}
		return res[i].Path < res[j].Path
	})

	return res
}

// ServeHTTP handles all incoming requests (Admin APIs, Mock Endpoints, and Collections)
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Apply CORS
	if s.corsEnabled {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// 1. Admin & System Management API routes
	if strings.HasPrefix(r.URL.Path, "/api/admin/") {
		s.handleAdminAPI(w, r)
		return
	}

	// Read Request Body (and restore for handler)
	var reqBodyBytes []byte
	if r.Body != nil {
		reqBodyBytes, _ = io.ReadAll(r.Body)
	}
	reqBodyStr := string(reqBodyBytes)

	// 2. Mock Endpoint Dispatch
	matchedEp, params := s.matchEndpoint(r.Method, r.URL.Path)
	if matchedEp != nil && matchedEp.Enabled {
		s.dispatchMockEndpoint(w, r, matchedEp, params, reqBodyStr, startTime)
		return
	}

	// 3. Stateful Collection Auto-CRUD (/api/resources/:col or /resources/:col)
	if strings.HasPrefix(r.URL.Path, "/api/resources/") || strings.HasPrefix(r.URL.Path, "/resources/") {
		if s.handleCollectionAutoCRUD(w, r, reqBodyStr, startTime) {
			return
		}
	}

	// 4. Upstream Proxy Fallback (if configured)
	if s.proxyTarget != "" {
		s.proxyPass(w, r, reqBodyBytes, startTime)
		return
	}

	// 5. 404 Not Found
	duration := time.Since(startTime).Milliseconds()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	respMsg := fmt.Sprintf(`{"error":"No mock endpoint configured for %s %s","hint":"Create this route in MockForge Studio or check path parameters"}`, r.Method, r.URL.Path)
	_, _ = w.Write([]byte(respMsg))

	s.trafficLogger.Log(models.TrafficEntry{
		ClientIP:        r.RemoteAddr,
		Method:          r.Method,
		Path:            r.URL.Path,
		QueryParams:     r.URL.Query(),
		RequestHeaders:  r.Header,
		RequestBody:     reqBodyStr,
		ResponseStatus:  http.StatusNotFound,
		ResponseHeaders: w.Header(),
		ResponseBody:    respMsg,
		DurationMs:      duration,
		Error:           "404 Not Found",
	})
}

// dispatchMockEndpoint handles authentication, chaos errors, latency, and template rendering
func (s *Server) dispatchMockEndpoint(w http.ResponseWriter, r *http.Request, ep *models.Endpoint, params map[string]string, reqBody string, startTime time.Time) {
	// A. Authentication Guard Check
	authRes := auth.ValidateRequest(r, ep.Auth)
	if !authRes.Authenticated {
		duration := time.Since(startTime).Milliseconds()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(authRes.StatusCode)
		errBody := fmt.Sprintf(`{"error":"Unauthorized","message":"%s"}`, authRes.ErrorMessage)
		_, _ = w.Write([]byte(errBody))

		s.trafficLogger.Log(models.TrafficEntry{
			ClientIP:        r.RemoteAddr,
			Method:          r.Method,
			Path:            r.URL.Path,
			QueryParams:     r.URL.Query(),
			RequestHeaders:  r.Header,
			RequestBody:     reqBody,
			ResponseStatus:  authRes.StatusCode,
			ResponseHeaders: w.Header(),
			ResponseBody:    errBody,
			DurationMs:      duration,
			MatchedEndpoint: fmt.Sprintf("[%s] %s", ep.Method, ep.Path),
			Error:           authRes.ErrorMessage,
		})
		return
	}

	// B. Latency Simulation Calculation
	var simulatedDelayMs int64
	if ep.Latency.Enabled {
		if ep.Latency.Mode == models.LatencyModeFixed && ep.Latency.FixedMs > 0 {
			simulatedDelayMs = int64(ep.Latency.FixedMs)
		} else if ep.Latency.Mode == models.LatencyModeRandom && ep.Latency.MaxMs >= ep.Latency.MinMs {
			nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(ep.Latency.MaxMs-ep.Latency.MinMs+1)))
			simulatedDelayMs = int64(ep.Latency.MinMs) + nBig.Int64()
		}
		if simulatedDelayMs > 0 {
			time.Sleep(time.Duration(simulatedDelayMs) * time.Millisecond)
		}
	}

	// C. Chaos Failure Injection
	if ep.Chaos.Enabled && ep.Chaos.Rate > 0 {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(1000))
		roll := float64(nBig.Int64()) / 1000.0
		if roll < ep.Chaos.Rate {
			duration := time.Since(startTime).Milliseconds()
			chaosStatus := ep.Chaos.StatusCode
			if chaosStatus == 0 {
				chaosStatus = http.StatusInternalServerError
			}

			chaosBody := ep.Chaos.ResponseBody
			if strings.TrimSpace(chaosBody) == "" {
				chaosBody = fmt.Sprintf(`{"error":"Simulated Chaos Failure","status":%d,"timestamp":"%s"}`, chaosStatus, time.Now().Format(time.RFC3339))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(chaosStatus)
			_, _ = w.Write([]byte(chaosBody))

			s.trafficLogger.Log(models.TrafficEntry{
				ClientIP:         r.RemoteAddr,
				Method:           r.Method,
				Path:             r.URL.Path,
				QueryParams:      r.URL.Query(),
				RequestHeaders:   r.Header,
				RequestBody:      reqBody,
				ResponseStatus:   chaosStatus,
				ResponseHeaders:  w.Header(),
				ResponseBody:     chaosBody,
				DurationMs:       duration,
				SimulatedDelayMs: simulatedDelayMs,
				MatchedEndpoint:  fmt.Sprintf("[%s] %s", ep.Method, ep.Path),
				Error:            "Simulated Chaos Error",
			})
			return
		}
	}

	// D. Build Template Context and Evaluate
	reqCtx := &template.RequestContext{
		Params:    params,
		Query:     r.URL.Query(),
		Headers:   r.Header,
		Body:      reqBody,
		Method:    r.Method,
		Path:      r.URL.Path,
		ClientIP:  r.RemoteAddr,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	evaluatedBody := s.templateEngine.Evaluate(ep.Response.Body, reqCtx)

	// E. Set Custom Headers
	for hK, hV := range ep.Response.Headers {
		evalHeaderVal := s.templateEngine.Evaluate(hV, reqCtx)
		w.Header().Set(hK, evalHeaderVal)
	}

	contentType := ep.Response.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)

	statusCode := ep.Response.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(evaluatedBody))

	duration := time.Since(startTime).Milliseconds()

	s.trafficLogger.Log(models.TrafficEntry{
		ClientIP:         r.RemoteAddr,
		Method:           r.Method,
		Path:             r.URL.Path,
		QueryParams:      r.URL.Query(),
		RequestHeaders:   r.Header,
		RequestBody:      reqBody,
		ResponseStatus:   statusCode,
		ResponseHeaders:  w.Header(),
		ResponseBody:     evaluatedBody,
		DurationMs:       duration,
		SimulatedDelayMs: simulatedDelayMs,
		MatchedEndpoint:  fmt.Sprintf("[%s] %s", ep.Method, ep.Path),
	})
}

// handleCollectionAutoCRUD processes automatic REST CRUD on stateful collections
func (s *Server) handleCollectionAutoCRUD(w http.ResponseWriter, r *http.Request, reqBody string, startTime time.Time) bool {
	cleanPath := strings.TrimPrefix(r.URL.Path, "/api/resources/")
	cleanPath = strings.TrimPrefix(cleanPath, "/resources/")
	parts := strings.Split(cleanPath, "/")

	if len(parts) == 0 || parts[0] == "" {
		return false
	}

	colName := parts[0]
	var itemID string
	if len(parts) > 1 {
		itemID = parts[1]
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		if itemID == "" {
			// List collection items
			items, total, err := s.store.ListItems(colName, r.URL.Query())
			if err != nil {
				writeJSONError(w, http.StatusNotFound, err.Error())
				return true
			}
			w.Header().Set("X-Total-Count", strconv.Itoa(total))
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(items)
		} else {
			// Get item by ID
			item, found, err := s.store.GetItem(colName, itemID)
			if err != nil || !found {
				writeJSONError(w, http.StatusNotFound, fmt.Sprintf("Item '%s' not found in collection '%s'", itemID, colName))
				return true
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(item)
		}

	case "POST":
		var bodyMap map[string]interface{}
		if err := json.Unmarshal([]byte(reqBody), &bodyMap); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload in request body")
			return true
		}
		newItem, err := s.store.CreateItem(colName, bodyMap)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return true
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(newItem)

	case "PUT":
		if itemID == "" {
			writeJSONError(w, http.StatusBadRequest, "PUT requires item ID in path")
			return true
		}
		var bodyMap map[string]interface{}
		if err := json.Unmarshal([]byte(reqBody), &bodyMap); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload in request body")
			return true
		}
		updated, found, err := s.store.UpdateItem(colName, itemID, bodyMap, false)
		if err != nil || !found {
			writeJSONError(w, http.StatusNotFound, fmt.Sprintf("Item '%s' not found", itemID))
			return true
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(updated)

	case "PATCH":
		if itemID == "" {
			writeJSONError(w, http.StatusBadRequest, "PATCH requires item ID in path")
			return true
		}
		var bodyMap map[string]interface{}
		if err := json.Unmarshal([]byte(reqBody), &bodyMap); err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload in request body")
			return true
		}
		updated, found, err := s.store.UpdateItem(colName, itemID, bodyMap, true)
		if err != nil || !found {
			writeJSONError(w, http.StatusNotFound, fmt.Sprintf("Item '%s' not found", itemID))
			return true
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(updated)

	case "DELETE":
		if itemID == "" {
			writeJSONError(w, http.StatusBadRequest, "DELETE requires item ID in path")
			return true
		}
		deleted, err := s.store.DeleteItem(colName, itemID)
		if err != nil || !deleted {
			writeJSONError(w, http.StatusNotFound, fmt.Sprintf("Item '%s' not found", itemID))
			return true
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"Item deleted"}`))

	default:
		return false
	}

	duration := time.Since(startTime).Milliseconds()
	s.trafficLogger.Log(models.TrafficEntry{
		ClientIP:          r.RemoteAddr,
		Method:            r.Method,
		Path:              r.URL.Path,
		QueryParams:       r.URL.Query(),
		RequestHeaders:    r.Header,
		RequestBody:       reqBody,
		ResponseStatus:    http.StatusOK,
		ResponseHeaders:   w.Header(),
		DurationMs:        duration,
		MatchedCollection: colName,
	})

	return true
}

// matchEndpoint searches for configured mock route
func (s *Server) matchEndpoint(method, reqPath string) (*models.Endpoint, map[string]string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reqParts := splitPath(reqPath)

	for _, ep := range s.endpoints {
		if !strings.EqualFold(ep.Method, method) {
			continue
		}

		epParts := splitPath(ep.Path)
		matched, params := matchPathParts(epParts, reqParts)
		if matched {
			return ep, params
		}
	}

	return nil, nil
}

func matchPathParts(patternParts, reqParts []string) (bool, map[string]string) {
	if len(patternParts) != len(reqParts) {
		// Check for wildcard /*path at end
		if len(patternParts) > 0 && strings.HasPrefix(patternParts[len(patternParts)-1], "*") {
			if len(reqParts) >= len(patternParts)-1 {
				params := make(map[string]string)
				for i := 0; i < len(patternParts)-1; i++ {
					if strings.HasPrefix(patternParts[i], ":") || (strings.HasPrefix(patternParts[i], "{") && strings.HasSuffix(patternParts[i], "}")) {
						pName := strings.TrimPrefix(patternParts[i], ":")
						pName = strings.Trim(pName, "{}")
						params[pName] = reqParts[i]
					} else if patternParts[i] != reqParts[i] {
						return false, nil
					}
				}
				wildcardName := strings.TrimPrefix(patternParts[len(patternParts)-1], "*")
				params[wildcardName] = strings.Join(reqParts[len(patternParts)-1:], "/")
				return true, params
			}
		}
		return false, nil
	}

	params := make(map[string]string)
	for i := range patternParts {
		pPart := patternParts[i]
		rPart := reqParts[i]

		if strings.HasPrefix(pPart, ":") {
			params[pPart[1:]] = rPart
		} else if strings.HasPrefix(pPart, "{") && strings.HasSuffix(pPart, "}") {
			params[pPart[1:len(pPart)-1]] = rPart
		} else if pPart != rPart {
			return false, nil
		}
	}

	return true, params
}

func splitPath(p string) []string {
	clean := strings.Trim(p, "/")
	if clean == "" {
		return []string{}
	}
	return strings.Split(clean, "/")
}

// handleAdminAPI routes admin actions
func (s *Server) handleAdminAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/")
	w.Header().Set("Content-Type", "application/json")

	switch {
	case path == "metrics" && r.Method == "GET":
		s.handleGetMetrics(w, r)
	case path == "endpoints" && r.Method == "GET":
		_ = json.NewEncoder(w).Encode(s.ListEndpoints())
	case path == "endpoints" && r.Method == "POST":
		s.handleCreateEndpoint(w, r)
	case strings.HasPrefix(path, "endpoints/") && r.Method == "PUT":
		id := strings.TrimPrefix(path, "endpoints/")
		s.handleUpdateEndpoint(w, r, id)
	case strings.HasPrefix(path, "endpoints/") && r.Method == "DELETE":
		id := strings.TrimPrefix(path, "endpoints/")
		if err := s.DeleteEndpoint(id); err != nil {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})

	case path == "collections" && r.Method == "GET":
		_ = json.NewEncoder(w).Encode(s.store.ListCollections())
	case path == "collections" && r.Method == "POST":
		s.handleCreateCollection(w, r)
	case strings.HasPrefix(path, "collections/") && r.Method == "PUT":
		id := strings.TrimPrefix(path, "collections/")
		s.handleUpdateCollection(w, r, id)
	case strings.HasPrefix(path, "collections/") && r.Method == "DELETE":
		id := strings.TrimPrefix(path, "collections/")
		if err := s.store.DeleteCollection(id); err != nil {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})

	case path == "traffic" && r.Method == "GET":
		limit := 100
		if lStr := r.URL.Query().Get("limit"); lStr != "" {
			if l, err := strconv.Atoi(lStr); err == nil {
				limit = l
			}
		}
		_ = json.NewEncoder(w).Encode(s.trafficLogger.GetHistory(limit))
	case path == "traffic" && r.Method == "DELETE":
		s.trafficLogger.Clear()
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})

	case path == "traffic/stream" && r.Method == "GET":
		s.handleTrafficSSE(w, r)

	case path == "auth/generate-token" && r.Method == "POST":
		s.handleGenerateAuthToken(w, r)

	case path == "openapi/import" && r.Method == "POST":
		s.handleOpenAPIImport(w, r)
	case path == "openapi/export" && r.Method == "GET":
		spec, err := openapi.ExportOpenAPI(s.ListEndpoints(), "MockForge API", "1.0.0")
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		_ = json.NewEncoder(w).Encode(spec)

	case path == "workspace/export" && r.Method == "GET":
		ws := openapi.ExportWorkspace("MockForge Workspace", s.ListEndpoints(), s.store.ListCollections())
		_ = json.NewEncoder(w).Encode(ws)
	case path == "workspace/import" && r.Method == "POST":
		s.handleWorkspaceImport(w, r)

	case path == "reset-demo" && r.Method == "POST":
		s.SeedDemoData()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"endpoints":   len(s.ListEndpoints()),
			"collections": len(s.store.ListCollections()),
		})

	default:
		writeJSONError(w, http.StatusNotFound, fmt.Sprintf("Admin endpoint '%s' not found", path))
	}
}

func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	stats := s.trafficLogger.Stats()
	metrics := map[string]interface{}{
		"activeEndpoints":   len(s.ListEndpoints()),
		"activeCollections": len(s.store.ListCollections()),
		"trafficStats":      stats,
	}
	_ = json.NewEncoder(w).Encode(metrics)
}

func (s *Server) handleCreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var ep models.Endpoint
	if err := json.NewDecoder(r.Body).Decode(&ep); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	created, err := s.AddEndpoint(ep)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (s *Server) handleUpdateEndpoint(w http.ResponseWriter, r *http.Request, id string) {
	var ep models.Endpoint
	if err := json.NewDecoder(r.Body).Decode(&ep); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	updated, err := s.UpdateEndpoint(id, ep)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = json.NewEncoder(w).Encode(updated)
}

func (s *Server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	var col models.Collection
	if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	if col.ID == "" {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(900000))
		col.ID = fmt.Sprintf("col_%d", 100000+nBig.Int64())
	}
	created, err := s.store.CreateCollection(col)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (s *Server) handleUpdateCollection(w http.ResponseWriter, r *http.Request, id string) {
	var col models.Collection
	if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	updated, err := s.store.UpdateCollection(id, col)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = json.NewEncoder(w).Encode(updated)
}

// handleTrafficSSE streams live traffic to browser via Server-Sent Events
func (s *Server) handleTrafficSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch, unsub := s.trafficLogger.Subscribe()
	defer unsub()

	// Send initial connected ping
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-ch:
			if !ok {
				return
			}
			bytes, err := json.Marshal(entry)
			if err == nil {
				_, _ = fmt.Fprintf(w, "event: traffic\ndata: %s\n\n", string(bytes))
				flusher.Flush()
			}
		}
	}
}

func (s *Server) handleGenerateAuthToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Secret   string                 `json:"secret"`
		Duration int                    `json:"durationMinutes"`
		Claims   map[string]interface{} `json:"claims"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Secret == "" {
		req.Secret = "mockforge-jwt-secret"
	}
	dur := time.Duration(req.Duration) * time.Minute
	if dur == 0 {
		dur = 24 * time.Hour
	}
	token, err := auth.GenerateMockJWT(req.Secret, req.Claims, dur)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token":     token,
		"expiresIn": int64(dur.Seconds()),
		"claims":    req.Claims,
	})
}

func (s *Server) handleOpenAPIImport(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	endpoints, err := openapi.ImportOpenAPI(bytes)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	count := 0
	for _, ep := range endpoints {
		_, err := s.AddEndpoint(ep)
		if err == nil {
			count++
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"importedCount":  count,
		"totalEndpoints": len(s.ListEndpoints()),
	})
}

func (s *Server) handleWorkspaceImport(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	ws, err := openapi.ImportWorkspace(bytes)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	s.endpoints = make(map[string]*models.Endpoint)
	s.mu.Unlock()

	for _, ep := range ws.Endpoints {
		_, _ = s.AddEndpoint(ep)
	}

	for _, col := range ws.Collections {
		_, _ = s.store.CreateCollection(col)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"endpoints":   len(ws.Endpoints),
		"collections": len(ws.Collections),
	})
}

func (s *Server) proxyPass(w http.ResponseWriter, r *http.Request, reqBody []byte, startTime time.Time) {
	targetURL, err := url.Parse(s.proxyTarget)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "Invalid proxy target")
		return
	}

	proxyReqURL := *targetURL
	proxyReqURL.Path = r.URL.Path
	proxyReqURL.RawQuery = r.URL.RawQuery

	proxyReq, err := http.NewRequest(r.Method, proxyReqURL.String(), strings.NewReader(string(reqBody)))
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "Failed to create upstream request")
		return
	}

	for k, vv := range r.Header {
		for _, v := range vv {
			proxyReq.Header.Add(k, v)
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Sprintf("Proxy upstream error: %v", err))
		return
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBytes)

	duration := time.Since(startTime).Milliseconds()
	s.trafficLogger.Log(models.TrafficEntry{
		ClientIP:        r.RemoteAddr,
		Method:          r.Method,
		Path:            r.URL.Path,
		QueryParams:     r.URL.Query(),
		RequestHeaders:  r.Header,
		RequestBody:     string(reqBody),
		ResponseStatus:  resp.StatusCode,
		ResponseHeaders: w.Header(),
		ResponseBody:    string(respBytes),
		DurationMs:      duration,
		MatchedEndpoint: "Proxy Upstream",
	})
}

// SeedDemoData pre-populates rich realistic mock endpoints and stateful collections
func (s *Server) SeedDemoData() {
	s.mu.Lock()
	s.endpoints = make(map[string]*models.Endpoint)
	s.mu.Unlock()

	now := time.Now()

	// 1. Dynamic User List with Faker
	_, _ = s.AddEndpoint(models.Endpoint{
		ID:          "ep-users-list",
		Name:        "List Dynamic Users",
		Description: "Generates paginated realistic user profiles with dynamic avatars, emails, and job titles",
		Method:      "GET",
		Path:        "/api/v1/users",
		Enabled:     true,
		Tags:        []string{"users", "crm"},
		Response: models.ResponseMock{
			StatusCode:  200,
			ContentType: "application/json",
			Headers:     map[string]string{"X-Powered-By": "MockForge", "Cache-Control": "no-cache"},
			Body: `[
  {{#repeat 6}}
  {
    "id": "{{faker.uuid}}",
    "name": "{{faker.name}}",
    "email": "{{faker.email}}",
    "role": "{{faker.role}}",
    "jobTitle": "{{faker.jobTitle}}",
    "avatar": "{{faker.avatar}}",
    "company": "{{faker.company}}",
    "city": "{{faker.city}}",
    "country": "{{faker.country}}",
    "status": "{{faker.status}}"
  }
  {{/repeat}}
]`,
		},
		Latency: models.LatencyConfig{
			Enabled: true,
			Mode:    models.LatencyModeRandom,
			MinMs:   60,
			MaxMs:   220,
		},
		Chaos: models.ChaosConfig{
			Enabled:    false,
			Rate:       0.05,
			StatusCode: 500,
		},
		Auth:      models.AuthConfig{Type: models.AuthTypeNone},
		CreatedAt: now,
		UpdatedAt: now,
	})

	// 2. User Detail by ID with Parameter Echo
	_, _ = s.AddEndpoint(models.Endpoint{
		ID:          "ep-users-detail",
		Name:        "Get User Profile",
		Description: "Fetches user by URL parameter :id with dynamic profile details",
		Method:      "GET",
		Path:        "/api/v1/users/:id",
		Enabled:     true,
		Tags:        []string{"users"},
		Response: models.ResponseMock{
			StatusCode:  200,
			ContentType: "application/json",
			Headers:     map[string]string{"X-Route-Matched": "users-detail"},
			Body: `{
  "id": "{{req.params.id}}",
  "name": "{{faker.name}}",
  "email": "{{faker.email}}",
  "role": "{{faker.role}}",
  "jobTitle": "{{faker.jobTitle}}",
  "company": "{{faker.company}}",
  "phone": "{{faker.phone}}",
  "address": {
    "street": "{{faker.street}}",
    "city": "{{faker.city}}",
    "country": "{{faker.country}}",
    "zipCode": "{{faker.zipCode}}"
  },
  "status": "active",
  "createdAt": "{{faker.pastDate(90)}}"
}`,
		},
		Latency: models.LatencyConfig{
			Enabled: true,
			Mode:    models.LatencyModeFixed,
			FixedMs: 120,
		},
		Auth:      models.AuthConfig{Type: models.AuthTypeNone},
		CreatedAt: now,
		UpdatedAt: now,
	})

	// 3. Mock Login Endpoint generating JWT
	_, _ = s.AddEndpoint(models.Endpoint{
		ID:          "ep-auth-login",
		Name:        "User Authentication (Login)",
		Description: "Accepts email/password and returns a bearer JWT simulation token",
		Method:      "POST",
		Path:        "/api/v1/auth/login",
		Enabled:     true,
		Tags:        []string{"auth"},
		Response: models.ResponseMock{
			StatusCode:  200,
			ContentType: "application/json",
			Body: `{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTYiLCJuYW1lIjoiQWxpY2UgRGV2ZWxvcGVyIiwiZW1haWwiOiJhbGljZUBsaW5lYXIuYXBwIiwicm9sZSI6ImFkbWluIn0.mocked_signature_hash_token_abc123",
  "tokenType": "Bearer",
  "expiresIn": 86400,
  "user": {
    "id": "usr_9981",
    "name": "{{faker.name}}",
    "email": "{{req.body.email}}",
    "role": "admin"
  }
}`,
		},
		Latency: models.LatencyConfig{
			Enabled: true,
			Mode:    models.LatencyModeFixed,
			FixedMs: 250,
		},
		Auth:      models.AuthConfig{Type: models.AuthTypeNone},
		CreatedAt: now,
		UpdatedAt: now,
	})

	// 4. Protected Secure Analytics Endpoint (Bearer Protected + Jitter)
	_, _ = s.AddEndpoint(models.Endpoint{
		ID:          "ep-analytics-secure",
		Name:        "Protected Analytics Metrics",
		Description: "Requires Bearer token and simulates high-throughput telemetry numbers",
		Method:      "GET",
		Path:        "/api/v1/analytics/overview",
		Enabled:     true,
		Tags:        []string{"analytics", "protected"},
		Response: models.ResponseMock{
			StatusCode:  200,
			ContentType: "application/json",
			Body: `{
  "activeSessions": {{faker.number(1200, 8500)}},
  "mrrUSD": {{faker.price(45000, 98000)}},
  "requestsPerSec": {{faker.number(250, 1800)}},
  "systemLatencyP99Ms": {{faker.number(25, 95)}},
  "errorRatePercent": 0.02,
  "topRegions": [
    {"region": "us-east-1", "load": "42%"},
    {"region": "eu-central-1", "load": "38%"},
    {"region": "ap-southeast-1", "load": "20%"}
  ]
}`,
		},
		Latency: models.LatencyConfig{
			Enabled: true,
			Mode:    models.LatencyModeRandom,
			MinMs:   100,
			MaxMs:   350,
		},
		Auth: models.AuthConfig{
			Type:  models.AuthTypeBearer,
			Token: "mockforge-super-token-2026",
		},
		CreatedAt: now,
		UpdatedAt: now,
	})

	// 5. Stateful Products Collection Seed
	_, _ = s.store.CreateCollection(models.Collection{
		ID:          "col-products",
		Name:        "products",
		Description: "Stateful eCommerce catalog with full REST CRUD, search, and filter",
		Items: []map[string]interface{}{
			{
				"id":          1,
				"title":       "Ergonomic Split Mechanical Keyboard",
				"category":    "hardware",
				"price":       199.99,
				"rating":      4.9,
				"inStock":     true,
				"description": "Gateron Oil King linear switches with aluminum CNC case",
				"createdAt":   now.AddDate(0, 0, -10).Format(time.RFC3339),
			},
			{
				"id":          2,
				"title":       "4K OLED Ultra-Wide Studio Monitor 34\"",
				"category":    "hardware",
				"price":       899.00,
				"rating":      4.8,
				"inStock":     true,
				"description": "175Hz refresh rate with True Black HDR and Thunderbolt 4 hub",
				"createdAt":   now.AddDate(0, 0, -8).Format(time.RFC3339),
			},
			{
				"id":          3,
				"title":       "Noise Cancelling Wireless ANC Headphones",
				"category":    "audio",
				"price":       349.50,
				"rating":      4.7,
				"inStock":     true,
				"description": "Spatial audio with 40-hour battery life and custom planar drivers",
				"createdAt":   now.AddDate(0, 0, -5).Format(time.RFC3339),
			},
			{
				"id":          4,
				"title":       "Minimalist Merino Wool Desk Pad",
				"category":    "accessories",
				"price":       55.00,
				"rating":      4.6,
				"inStock":     false,
				"description": "Water-resistant organic wool felt with non-slip cork base",
				"createdAt":   now.AddDate(0, 0, -2).Format(time.RFC3339),
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  http.StatusText(status),
		"status": status,
		"detail": msg,
	})
}
