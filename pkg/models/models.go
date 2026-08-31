package models

import (
	"errors"
	"strings"
	"time"
)

// HTTPMethod constants
const (
	MethodGET     = "GET"
	MethodPOST    = "POST"
	MethodPUT     = "PUT"
	MethodPATCH   = "PATCH"
	MethodDELETE  = "DELETE"
	MethodOPTIONS = "OPTIONS"
	MethodHEAD    = "HEAD"
)

// AuthType constants
const (
	AuthTypeNone   = "none"
	AuthTypeBearer = "bearer"
	AuthTypeAPIKey = "apikey"
	AuthTypeBasic  = "basic"
	AuthTypeJWT    = "jwt"
)

// IsAuthRequired checks if authentication is active
func IsAuthRequired(authType string) bool {
	return authType != "" && authType != AuthTypeNone
}


// LatencyMode constants
const (
	LatencyModeFixed  = "fixed"
	LatencyModeRandom = "random"
)

// ResponseMock represents the mock response specification
type ResponseMock struct {
	StatusCode  int               `json:"statusCode"`
	Headers     map[string]string `json:"headers"`
	ContentType string            `json:"contentType"`
	Body        string            `json:"body"`
}

// LatencyConfig defines network simulation rules
type LatencyConfig struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"` // "fixed" or "random"
	FixedMs int    `json:"fixedMs"`
	MinMs   int    `json:"minMs"`
	MaxMs   int    `json:"maxMs"`
}

// ChaosConfig defines failure rate and error injection
type ChaosConfig struct {
	Enabled      bool    `json:"enabled"`
	Rate         float64 `json:"rate"` // 0.0 to 1.0 (e.g., 0.2 = 20%)
	StatusCode   int     `json:"statusCode"`
	ResponseBody string  `json:"responseBody"`
}

// AuthConfig defines authentication guard requirements
type AuthConfig struct {
	Type           string                 `json:"type"` // "none", "bearer", "apikey", "basic", "jwt"
	Token          string                 `json:"token,omitempty"`
	HeaderName     string                 `json:"headerName,omitempty"`     // for apikey (default: "X-API-Key")
	QueryParam     string                 `json:"queryParam,omitempty"`     // for apikey (default: "api_key")
	BasicUsername  string                 `json:"basicUsername,omitempty"`  // for basic auth
	BasicPassword  string                 `json:"basicPassword,omitempty"`  // for basic auth
	JWTSecret      string                 `json:"jwtSecret,omitempty"`      // for jwt signature verification
	RequiredClaims map[string]interface{} `json:"requiredClaims,omitempty"` // required JWT claims
}

// Endpoint represents a configurable mock HTTP route
type Endpoint struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Method      string        `json:"method"`
	Path        string        `json:"path"`
	Enabled     bool          `json:"enabled"`
	Tags        []string      `json:"tags,omitempty"`
	Response    ResponseMock  `json:"response"`
	Latency     LatencyConfig `json:"latency"`
	Chaos       ChaosConfig   `json:"chaos"`
	Auth        AuthConfig    `json:"auth"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

// Validate checks if the endpoint definition is valid
func (e *Endpoint) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return errors.New("endpoint ID cannot be empty")
	}
	if strings.TrimSpace(e.Path) == "" {
		return errors.New("endpoint Path cannot be empty")
	}
	if !strings.HasPrefix(e.Path, "/") {
		e.Path = "/" + e.Path
	}
	if strings.TrimSpace(e.Method) == "" {
		e.Method = MethodGET
	} else {
		e.Method = strings.ToUpper(strings.TrimSpace(e.Method))
	}
	if e.Response.StatusCode == 0 {
		e.Response.StatusCode = 200
	}
	if e.Response.ContentType == "" {
		e.Response.ContentType = "application/json"
	}
	if e.Latency.Enabled {
		if e.Latency.Mode == "" {
			e.Latency.Mode = LatencyModeFixed
		}
		if e.Latency.Mode == LatencyModeRandom && e.Latency.MaxMs < e.Latency.MinMs {
			e.Latency.MaxMs = e.Latency.MinMs
		}
	}
	if e.Chaos.Enabled {
		if e.Chaos.StatusCode == 0 {
			e.Chaos.StatusCode = 500
		}
		if e.Chaos.Rate < 0 {
			e.Chaos.Rate = 0
		} else if e.Chaos.Rate > 1.0 {
			e.Chaos.Rate = 1.0
		}
	}
	if e.Auth.Type == "" {
		e.Auth.Type = AuthTypeNone
	}
	return nil
}

// Collection represents a stateful resource table with auto-CRUD
type Collection struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Items       []map[string]interface{} `json:"items"`
	CreatedAt   time.Time                `json:"createdAt"`
	UpdatedAt   time.Time                `json:"updatedAt"`
}

// Validate checks collection integrity
func (c *Collection) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return errors.New("collection ID cannot be empty")
	}
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("collection Name cannot be empty")
	}
	c.Name = strings.ToLower(strings.TrimSpace(c.Name))
	if c.Items == nil {
		c.Items = make([]map[string]interface{}, 0)
	}
	return nil
}

// TrafficEntry represents an intercepted HTTP request/response log
type TrafficEntry struct {
	ID               string              `json:"id"`
	Timestamp        time.Time           `json:"timestamp"`
	ClientIP         string              `json:"clientIp"`
	Method           string              `json:"method"`
	Path             string              `json:"path"`
	QueryParams      map[string][]string `json:"queryParams,omitempty"`
	RequestHeaders   map[string][]string `json:"requestHeaders,omitempty"`
	RequestBody      string              `json:"requestBody,omitempty"`
	ResponseStatus   int                 `json:"responseStatus"`
	ResponseHeaders  map[string][]string `json:"responseHeaders,omitempty"`
	ResponseBody     string              `json:"responseBody,omitempty"`
	DurationMs       int64               `json:"durationMs"`
	SimulatedDelayMs int64               `json:"simulatedDelayMs"`
	MatchedEndpoint  string              `json:"matchedEndpoint,omitempty"`
	MatchedCollection string             `json:"matchedCollection,omitempty"`
	Error            string              `json:"error,omitempty"`
}

// Workspace represents a full export/import package
type Workspace struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	ExportedAt  time.Time    `json:"exportedAt"`
	Endpoints   []Endpoint   `json:"endpoints"`
	Collections []Collection `json:"collections"`
}
