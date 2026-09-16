package engine_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"apimock/pkg/engine"
	"apimock/pkg/models"
)

func TestConditionalScenarios(t *testing.T) {
	srv := engine.NewServer("")

	// Setup endpoint with multiple scenarios
	ep := models.Endpoint{
		ID:      "ep-conditional-test",
		Name:    "Conditional User Endpoint",
		Method:  "POST",
		Path:    "/api/v1/auth/login",
		Enabled: true,
		Response: models.ResponseMock{
			StatusCode:  200,
			ContentType: "application/json",
			Body:        `{"status":"default_success","role":"regular_user"}`,
		},
		Scenarios: []models.Scenario{
			{
				ID:        "sc-admin",
				Name:      "Admin Login Scenario",
				Enabled:   true,
				Priority:  1,
				MatchMode: models.MatchModeAll,
				Conditions: []models.Condition{
					{
						Source:   models.ConditionSourceBody,
						Property: "email",
						Operator: models.OperatorEquals,
						Value:    "admin@mockforge.io",
					},
					{
						Source:   models.ConditionSourceHeader,
						Property: "X-Client-Type",
						Operator: models.OperatorEquals,
						Value:    "internal",
					},
				},
				Response: models.ResponseMock{
					StatusCode:  200,
					ContentType: "application/json",
					Body:        `{"status":"admin_granted","token":"admin-super-token-123"}`,
				},
			},
			{
				ID:        "sc-invalid-pass",
				Name:      "Invalid Password 401",
				Enabled:   true,
				Priority:  2,
				MatchMode: models.MatchModeAny,
				Conditions: []models.Condition{
					{
						Source:   models.ConditionSourceBody,
						Property: "password",
						Operator: models.OperatorEquals,
						Value:    "wrong_password",
					},
				},
				Response: models.ResponseMock{
					StatusCode:  401,
					ContentType: "application/json",
					Body:        `{"error":"Invalid credentials","code":"AUTH_INVALID_PASSWORD"}`,
				},
			},
			{
				ID:        "sc-empty-password",
				Name:      "Empty Password 400",
				Enabled:   true,
				Priority:  3,
				MatchMode: models.MatchModeAll,
				Conditions: []models.Condition{
					{
						Source:   models.ConditionSourceBody,
						Property: "password",
						Operator: models.OperatorIsEmpty,
					},
				},
				Response: models.ResponseMock{
					StatusCode:  400,
					ContentType: "application/json",
					Body:        `{"error":"Bad Request","message":"Password is required"}`,
				},
			},
		},
	}

	created, err := srv.AddEndpoint(ep)
	if err != nil {
		t.Fatalf("failed to register endpoint with scenarios: %v", err)
	}
	if len(created.Scenarios) != 3 {
		t.Fatalf("expected 3 scenarios, got %d", len(created.Scenarios))
	}

	t.Run("Match First Scenario: Admin Login with MatchModeAll", func(t *testing.T) {
		reqBody := `{"email":"admin@mockforge.io","password":"my-secret-password"}`
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Client-Type", "internal")

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var res map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if res["status"] != "admin_granted" {
			t.Errorf("expected status 'admin_granted', got %v", res["status"])
		}
		if res["token"] != "admin-super-token-123" {
			t.Errorf("expected admin token, got %v", res["token"])
		}
	})

	t.Run("Match Second Scenario: Invalid Password returning 401", func(t *testing.T) {
		reqBody := `{"email":"other@user.com","password":"wrong_password"}`
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d: %s", rec.Code, rec.Body.String())
		}

		var res map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["code"] != "AUTH_INVALID_PASSWORD" {
			t.Errorf("expected code 'AUTH_INVALID_PASSWORD', got %v", res["code"])
		}
	})

	t.Run("Match Third Scenario: Empty Password returning 400", func(t *testing.T) {
		reqBody := `{"email":"someone@user.com"}`
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Fallback to Default Response when no scenarios match", func(t *testing.T) {
		reqBody := `{"email":"regular@user.com","password":"valid_password"}`
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var res map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["status"] != "default_success" {
			t.Errorf("expected default response status 'default_success', got %v", res["status"])
		}
		if res["role"] != "regular_user" {
			t.Errorf("expected role 'regular_user', got %v", res["role"])
		}
	})
}

func TestScenarioConditionOperators(t *testing.T) {
	srv := engine.NewServer("")

	ep := models.Endpoint{
		ID:      "ep-operators-test",
		Name:    "Operator Testing Endpoint",
		Method:  "GET",
		Path:    "/api/v1/items/:id",
		Enabled: true,
		Response: models.ResponseMock{
			StatusCode: 200,
			Body:       `{"match":"default"}`,
		},
		Scenarios: []models.Scenario{
			{
				ID:        "sc-query-contains",
				Name:      "Query Contains Test",
				Enabled:   true,
				Priority:  1,
				MatchMode: models.MatchModeAll,
				Conditions: []models.Condition{
					{
						Source:   models.ConditionSourceQuery,
						Property: "filter",
						Operator: models.OperatorContains,
						Value:    "urgent",
					},
				},
				Response: models.ResponseMock{
					StatusCode: 200,
					Body:       `{"match":"query_contains"}`,
				},
			},
			{
				ID:        "sc-param-regex",
				Name:      "Param Regex Test",
				Enabled:   true,
				Priority:  2,
				MatchMode: models.MatchModeAll,
				Conditions: []models.Condition{
					{
						Source:   models.ConditionSourceParam,
						Property: "id",
						Operator: models.OperatorRegex,
						Value:    `^vip-\d+$`,
					},
				},
				Response: models.ResponseMock{
					StatusCode: 200,
					Body:       `{"match":"param_regex_vip"}`,
				},
			},
			{
				ID:        "sc-numeric-gt",
				Name:      "Query Numeric GT Test",
				Enabled:   true,
				Priority:  3,
				MatchMode: models.MatchModeAll,
				Conditions: []models.Condition{
					{
						Source:   models.ConditionSourceQuery,
						Property: "score",
						Operator: models.OperatorGt,
						Value:    "90",
					},
				},
				Response: models.ResponseMock{
					StatusCode: 200,
					Body:       `{"match":"high_score"}`,
				},
			},
		},
	}

	_, err := srv.AddEndpoint(ep)
	if err != nil {
		t.Fatalf("failed to add endpoint: %v", err)
	}

	t.Run("Query Contains Operator", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/items/item-1?filter=this-is-urgent-task", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		var res map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["match"] != "query_contains" {
			t.Errorf("expected 'query_contains', got %v", res["match"])
		}
	})

	t.Run("Param Regex Operator", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/items/vip-9988", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		var res map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["match"] != "param_regex_vip" {
			t.Errorf("expected 'param_regex_vip', got %v", res["match"])
		}
	})

	t.Run("Numeric GT Operator", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/items/item-1?score=95", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		var res map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["match"] != "high_score" {
			t.Errorf("expected 'high_score', got %v", res["match"])
		}
	})
}
