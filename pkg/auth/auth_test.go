package auth_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"apimock/pkg/auth"
	"apimock/pkg/models"
)

func TestValidateRequest(t *testing.T) {
	t.Run("None Auth Type always passes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		cfg := models.AuthConfig{Type: models.AuthTypeNone}
		res := auth.ValidateRequest(req, cfg)
		if !res.Authenticated || res.StatusCode != http.StatusOK {
			t.Errorf("expected None auth to pass, got %+v", res)
		}
	})

	t.Run("Bearer Auth validation", func(t *testing.T) {
		cfg := models.AuthConfig{Type: models.AuthTypeBearer, Token: "super-secret-123"}

		// Missing header
		req1 := httptest.NewRequest("GET", "/api/test", nil)
		res1 := auth.ValidateRequest(req1, cfg)
		if res1.Authenticated || res1.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 for missing header, got %+v", res1)
		}

		// Wrong token
		req2 := httptest.NewRequest("GET", "/api/test", nil)
		req2.Header.Set("Authorization", "Bearer wrong-token")
		res2 := auth.ValidateRequest(req2, cfg)
		if res2.Authenticated || res2.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 for wrong token, got %+v", res2)
		}

		// Valid token
		req3 := httptest.NewRequest("GET", "/api/test", nil)
		req3.Header.Set("Authorization", "Bearer super-secret-123")
		res3 := auth.ValidateRequest(req3, cfg)
		if !res3.Authenticated || res3.StatusCode != http.StatusOK {
			t.Errorf("expected valid bearer token to pass, got %+v", res3)
		}
	})

	t.Run("API Key Header and Query validation", func(t *testing.T) {
		cfg := models.AuthConfig{
			Type:       models.AuthTypeAPIKey,
			Token:      "api-key-999",
			HeaderName: "X-Custom-Key",
			QueryParam: "api_key",
		}

		// Header match
		req1 := httptest.NewRequest("GET", "/api/test", nil)
		req1.Header.Set("X-Custom-Key", "api-key-999")
		res1 := auth.ValidateRequest(req1, cfg)
		if !res1.Authenticated {
			t.Errorf("expected API Key header to pass, got %+v", res1)
		}

		// Query param match
		req2 := httptest.NewRequest("GET", "/api/test?api_key=api-key-999", nil)
		res2 := auth.ValidateRequest(req2, cfg)
		if !res2.Authenticated {
			t.Errorf("expected API Key query to pass, got %+v", res2)
		}

		// Missing
		req3 := httptest.NewRequest("GET", "/api/test", nil)
		res3 := auth.ValidateRequest(req3, cfg)
		if res3.Authenticated || res3.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 for missing api key, got %+v", res3)
		}
	})

	t.Run("Basic Auth validation", func(t *testing.T) {
		cfg := models.AuthConfig{
			Type:          models.AuthTypeBasic,
			BasicUsername: "admin",
			BasicPassword: "secretpassword",
		}

		req := httptest.NewRequest("GET", "/api/test", nil)
		basicVal := base64.StdEncoding.EncodeToString([]byte("admin:secretpassword"))
		req.Header.Set("Authorization", "Basic "+basicVal)

		res := auth.ValidateRequest(req, cfg)
		if !res.Authenticated || res.User != "admin" {
			t.Errorf("expected Basic Auth to pass for admin, got %+v", res)
		}

		// Wrong password
		reqWrong := httptest.NewRequest("GET", "/api/test", nil)
		wrongVal := base64.StdEncoding.EncodeToString([]byte("admin:wrongpass"))
		reqWrong.Header.Set("Authorization", "Basic "+wrongVal)
		resWrong := auth.ValidateRequest(reqWrong, cfg)
		if resWrong.Authenticated || resWrong.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 for wrong basic password, got %+v", resWrong)
		}
	})

	t.Run("JWT Auth validation and Claims checking", func(t *testing.T) {
		secret := "my-jwt-secret-32-chars-long-1234"

		// Generate valid token
		token, err := auth.GenerateMockJWT(secret, map[string]interface{}{
			"sub":   "user-100",
			"email": "user100@example.com",
			"role":  "admin",
		}, time.Hour)
		if err != nil {
			t.Fatalf("failed to generate mock JWT: %v", err)
		}

		cfg := models.AuthConfig{
			Type:      models.AuthTypeJWT,
			JWTSecret: secret,
			RequiredClaims: map[string]interface{}{
				"role": "admin",
			},
		}

		// Valid request
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		res := auth.ValidateRequest(req, cfg)
		if !res.Authenticated || res.User != "user-100" {
			t.Errorf("expected JWT auth to pass with user-100, got %+v", res)
		}

		// Claim mismatch (Forbidden 403)
		cfgViewer := models.AuthConfig{
			Type:      models.AuthTypeJWT,
			JWTSecret: secret,
			RequiredClaims: map[string]interface{}{
				"role": "superadmin", // token only has role: admin
			},
		}
		resViewer := auth.ValidateRequest(req, cfgViewer)
		if resViewer.Authenticated || resViewer.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for missing claim, got %+v", resViewer)
		}

		// Signature mismatch (Wrong Secret)
		cfgWrongSecret := models.AuthConfig{
			Type:      models.AuthTypeJWT,
			JWTSecret: "completely-different-secret-key-000",
		}
		resWrongSecret := auth.ValidateRequest(req, cfgWrongSecret)
		if resWrongSecret.Authenticated || resWrongSecret.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 for signature mismatch, got %+v", resWrongSecret)
		}
	})
}
