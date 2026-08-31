package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"apimock/pkg/models"
)

// AuthResult holds verification outcome
type AuthResult struct {
	Authenticated bool                   `json:"authenticated"`
	User          string                 `json:"user,omitempty"`
	Claims        map[string]interface{} `json:"claims,omitempty"`
	StatusCode    int                    `json:"statusCode"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
}

// ValidateRequest checks if an incoming HTTP request satisfies the endpoint's AuthConfig
func ValidateRequest(req *http.Request, config models.AuthConfig) AuthResult {
	if !models.IsAuthRequired(config.Type) {
		return AuthResult{Authenticated: true, StatusCode: http.StatusOK}
	}

	switch config.Type {
	case models.AuthTypeBearer:
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  "Missing Authorization header",
			}
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  "Invalid Authorization header format. Expected 'Bearer <token>'",
			}
		}
		token := strings.TrimSpace(parts[1])
		if config.Token != "" && token != config.Token {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  "Invalid Bearer token",
			}
		}
		return AuthResult{Authenticated: true, StatusCode: http.StatusOK, User: "bearer-user"}

	case models.AuthTypeAPIKey:
		hName := config.HeaderName
		if hName == "" {
			hName = "X-API-Key"
		}
		qName := config.QueryParam
		if qName == "" {
			qName = "api_key"
		}

		keyVal := req.Header.Get(hName)
		if keyVal == "" {
			keyVal = req.URL.Query().Get(qName)
		}

		if keyVal == "" {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  fmt.Sprintf("Missing API Key (expected header '%s' or query '%s')", hName, qName),
			}
		}
		if config.Token != "" && keyVal != config.Token {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  "Invalid API Key",
			}
		}
		return AuthResult{Authenticated: true, StatusCode: http.StatusOK, User: "api-key-user"}

	case models.AuthTypeBasic:
		user, pass, ok := req.BasicAuth()
		if !ok {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  "Missing or malformed Basic Authorization header",
			}
		}
		if config.BasicUsername != "" && user != config.BasicUsername {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  "Invalid Basic Auth username",
			}
		}
		if config.BasicPassword != "" && pass != config.BasicPassword {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  "Invalid Basic Auth password",
			}
		}
		return AuthResult{Authenticated: true, StatusCode: http.StatusOK, User: user}

	case models.AuthTypeJWT:
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  "Missing Authorization header for JWT verification",
			}
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  "Expected 'Bearer <jwt_token>' header",
			}
		}

		rawToken := strings.TrimSpace(parts[1])
		claims, err := VerifyJWT(rawToken, config.JWTSecret)
		if err != nil {
			return AuthResult{
				Authenticated: false,
				StatusCode:    http.StatusUnauthorized,
				ErrorMessage:  fmt.Sprintf("JWT Verification Failed: %v", err),
			}
		}

		// Verify required claims
		if len(config.RequiredClaims) > 0 {
			for reqKey, reqVal := range config.RequiredClaims {
				claimVal, ok := claims[reqKey]
				if !ok {
					return AuthResult{
						Authenticated: false,
						StatusCode:    http.StatusForbidden,
						ErrorMessage:  fmt.Sprintf("Forbidden: missing required claim '%s'", reqKey),
					}
				}
				if fmt.Sprintf("%v", claimVal) != fmt.Sprintf("%v", reqVal) {
					return AuthResult{
						Authenticated: false,
						StatusCode:    http.StatusForbidden,
						ErrorMessage:  fmt.Sprintf("Forbidden: claim '%s' expected '%v' but got '%v'", reqKey, reqVal, claimVal),
					}
				}
			}
		}

		user := "jwt-user"
		if sub, ok := claims["sub"].(string); ok {
			user = sub
		} else if email, ok := claims["email"].(string); ok {
			user = email
		}

		return AuthResult{
			Authenticated: true,
			StatusCode:    http.StatusOK,
			User:          user,
			Claims:        claims,
		}
	}

	return AuthResult{Authenticated: true, StatusCode: http.StatusOK}
}

// VerifyJWT parses and validates an HS256 JWT
func VerifyJWT(tokenString, secret string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format (must have 3 dot-separated parts)")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT header: %w", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid JWT header JSON: %w", err)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("invalid JWT payload JSON: %w", err)
	}

	// Verify HMAC-SHA256 signature if secret is specified
	if secret != "" {
		signingInput := parts[0] + "." + parts[1]
		expectedSig := computeHMACSHA256(signingInput, secret)
		actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
		if err != nil {
			return nil, fmt.Errorf("failed to decode JWT signature: %w", err)
		}
		if !hmac.Equal(expectedSig, actualSig) {
			return nil, fmt.Errorf("JWT signature mismatch")
		}
	}

	// Verify expiration `exp` claim if present
	if expVal, ok := claims["exp"]; ok {
		var expUnix int64
		switch v := expVal.(type) {
		case float64:
			expUnix = int64(v)
		case int64:
			expUnix = v
		case json.Number:
			expUnix, _ = v.Int64()
		}
		if expUnix > 0 && time.Now().Unix() > expUnix {
			return nil, fmt.Errorf("JWT token has expired")
		}
	}

	return claims, nil
}

// GenerateMockJWT generates a signed HS256 JWT
func GenerateMockJWT(secret string, claims map[string]interface{}, duration time.Duration) (string, error) {
	if claims == nil {
		claims = make(map[string]interface{})
	}

	now := time.Now()
	if _, ok := claims["iat"]; !ok {
		claims["iat"] = now.Unix()
	}
	if duration > 0 {
		if _, ok := claims["exp"]; !ok {
			claims["exp"] = now.Add(duration).Unix()
		}
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := encodedHeader + "." + encodedPayload
	var encodedSig string
	if secret != "" {
		sig := computeHMACSHA256(signingInput, secret)
		encodedSig = base64.RawURLEncoding.EncodeToString(sig)
	} else {
		encodedSig = base64.RawURLEncoding.EncodeToString([]byte("unsigned"))
	}

	return fmt.Sprintf("%s.%s.%s", encodedHeader, encodedPayload, encodedSig), nil
}

func computeHMACSHA256(data, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return h.Sum(nil)
}
