package engine_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"apimock/pkg/engine"
	"apimock/pkg/models"
)

func TestMockEngineServer(t *testing.T) {
	srv := engine.NewServer("")
	srv.SeedDemoData()

	t.Run("Dispatch Dynamic Mock Endpoint with Parameter Interpolation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/users/user-456", nil)
		rec := httptest.NewRecorder()

		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
		}

		var body map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode response JSON: %v", err)
		}

		if body["id"] != "user-456" {
			t.Errorf("expected param interpolation 'user-456', got %v", body["id"])
		}
		if name, ok := body["name"].(string); !ok || name == "" {
			t.Errorf("expected generated faker name, got %v", body["name"])
		}
	})

	t.Run("Enforce Auth Guard on Protected Route", func(t *testing.T) {
		// 1. Without Token -> 401
		reqUnauth := httptest.NewRequest("GET", "/api/v1/analytics/overview", nil)
		recUnauth := httptest.NewRecorder()
		srv.ServeHTTP(recUnauth, reqUnauth)
		if recUnauth.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for missing token, got %d", recUnauth.Code)
		}

		// 2. With Valid Bearer Token -> 200
		reqAuth := httptest.NewRequest("GET", "/api/v1/analytics/overview", nil)
		reqAuth.Header.Set("Authorization", "Bearer mockforge-super-token-2026")
		recAuth := httptest.NewRecorder()
		srv.ServeHTTP(recAuth, reqAuth)
		if recAuth.Code != http.StatusOK {
			t.Errorf("expected 200 OK for valid bearer token, got %d", recAuth.Code)
		}
	})

	t.Run("Simulate Latency Delay", func(t *testing.T) {
		_, _ = srv.AddEndpoint(models.Endpoint{
			ID:      "ep-latency-test",
			Path:    "/api/test/delayed",
			Method:  "GET",
			Enabled: true,
			Response: models.ResponseMock{
				StatusCode: 200,
				Body:       `{"status":"delayed_ok"}`,
			},
			Latency: models.LatencyConfig{
				Enabled: true,
				Mode:    models.LatencyModeFixed,
				FixedMs: 100,
			},
		})

		start := time.Now()
		req := httptest.NewRequest("GET", "/api/test/delayed", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		elapsed := time.Since(start)
		if elapsed < 90*time.Millisecond {
			t.Errorf("expected response to be delayed by at least ~100ms, elapsed: %v", elapsed)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rec.Code)
		}
	})

	t.Run("Simulate Chaos Failure Injection", func(t *testing.T) {
		_, _ = srv.AddEndpoint(models.Endpoint{
			ID:      "ep-chaos-test",
			Path:    "/api/test/chaos",
			Method:  "GET",
			Enabled: true,
			Response: models.ResponseMock{
				StatusCode: 200,
				Body:       `{"status":"normal"}`,
			},
			Chaos: models.ChaosConfig{
				Enabled:      true,
				Rate:         1.0, // 100% failure rate
				StatusCode:   503,
				ResponseBody: `{"error":"Service Unavailable - Chaos"}`,
			},
		})

		req := httptest.NewRequest("GET", "/api/test/chaos", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("expected 503 from chaos simulation, got %d", rec.Code)
		}
	})

	t.Run("Stateful Collection Auto-CRUD via HTTP", func(t *testing.T) {
		// 1. List items
		reqList := httptest.NewRequest("GET", "/api/resources/products?category=hardware", nil)
		recList := httptest.NewRecorder()
		srv.ServeHTTP(recList, reqList)
		if recList.Code != http.StatusOK {
			t.Fatalf("expected 200 for collection list, got %d", recList.Code)
		}

		var items []map[string]interface{}
		_ = json.Unmarshal(recList.Body.Bytes(), &items)
		if len(items) != 2 {
			t.Errorf("expected 2 hardware items, got %d", len(items))
		}

		// 2. Create new item (POST)
		newProduct := `{"title":"Wireless Mouse","price":79.99,"category":"accessories","inStock":true}`
		reqCreate := httptest.NewRequest("POST", "/api/resources/products", bytes.NewBufferString(newProduct))
		recCreate := httptest.NewRecorder()
		srv.ServeHTTP(recCreate, reqCreate)
		if recCreate.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created for collection POST, got %d", recCreate.Code)
		}

		var createdItem map[string]interface{}
		_ = json.Unmarshal(recCreate.Body.Bytes(), &createdItem)
		if createdItem["title"] != "Wireless Mouse" || createdItem["id"] == nil {
			t.Errorf("unexpected created item: %+v", createdItem)
		}
	})

	t.Run("Admin Metrics API", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/admin/metrics", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for admin metrics, got %d", rec.Code)
		}

		var metrics map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &metrics)
		if metrics["activeEndpoints"] == nil || metrics["trafficStats"] == nil {
			t.Errorf("missing expected metrics keys: %+v", metrics)
		}
	})

	t.Run("Admin Auth Token Generator API", func(t *testing.T) {
		body := `{"secret":"super-secret-jwt","durationMinutes":60,"claims":{"sub":"admin-1","role":"admin"}}`
		req := httptest.NewRequest("POST", "/api/admin/auth/generate-token", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for token generator, got %d", rec.Code)
		}

		var res map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if token, ok := res["token"].(string); !ok || token == "" {
			t.Errorf("expected generated JWT token string, got %+v", res)
		}
	})
}
