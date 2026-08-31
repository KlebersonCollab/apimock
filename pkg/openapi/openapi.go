package openapi

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"apimock/pkg/models"
)

// ImportOpenAPI parses an OpenAPI 3.0 JSON specification and generates MockForge endpoints
func ImportOpenAPI(specBytes []byte) ([]models.Endpoint, error) {
	if len(specBytes) == 0 {
		return nil, errors.New("empty OpenAPI specification")
	}

	var root map[string]interface{}
	if err := json.Unmarshal(specBytes, &root); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI JSON: %w", err)
	}

	pathsObj, ok := root["paths"].(map[string]interface{})
	if !ok || len(pathsObj) == 0 {
		return nil, errors.New("no 'paths' found in OpenAPI specification")
	}

	endpoints := make([]models.Endpoint, 0)
	now := time.Now()

	for rawPath, pathItemRaw := range pathsObj {
		pathItem, ok := pathItemRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Convert OpenAPI {param} to :param for router consistency
		convertedPath := convertOpenAPIPath(rawPath)

		methods := []string{"get", "post", "put", "patch", "delete", "options", "head"}
		for _, method := range methods {
			opRaw, ok := pathItem[method]
			if !ok {
				continue
			}

			op, ok := opRaw.(map[string]interface{})
			if !ok {
				continue
			}

			epID := generateID()
			epName := fmt.Sprintf("%s %s", strings.ToUpper(method), rawPath)
			if summary, ok := op["summary"].(string); ok && summary != "" {
				epName = summary
			}

			desc := ""
			if d, ok := op["description"].(string); ok {
				desc = d
			}

			tags := make([]string, 0)
			if rawTags, ok := op["tags"].([]interface{}); ok {
				for _, t := range rawTags {
					if ts, ok := t.(string); ok {
						tags = append(tags, ts)
					}
				}
			}

			statusCode := 200
			contentType := "application/json"
			bodyStr := "{}"

			// Parse response
			if responses, ok := op["responses"].(map[string]interface{}); ok {
				for codeStr, respRaw := range responses {
					cInt, err := strconv.Atoi(codeStr)
					if err == nil && cInt >= 200 && cInt < 300 {
						statusCode = cInt
					}
					if respObj, ok := respRaw.(map[string]interface{}); ok {
						if content, ok := respObj["content"].(map[string]interface{}); ok {
							for ct, mediaRaw := range content {
								contentType = ct
								if mediaObj, ok := mediaRaw.(map[string]interface{}); ok {
									if example, ok := mediaObj["example"]; ok {
										if exBytes, err := json.MarshalIndent(example, "", "  "); err == nil {
											bodyStr = string(exBytes)
										}
									} else if schema, ok := mediaObj["schema"].(map[string]interface{}); ok {
										bodyStr = synthesizeSchemaExample(schema)
									}
								}
								break
							}
						}
					}
					break
				}
			}

			// Parse security
			authCfg := models.AuthConfig{Type: models.AuthTypeNone}
			if sec, ok := op["security"].([]interface{}); ok && len(sec) > 0 {
				authCfg.Type = models.AuthTypeBearer
			}

			ep := models.Endpoint{
				ID:          epID,
				Name:        epName,
				Description: desc,
				Method:      strings.ToUpper(method),
				Path:        convertedPath,
				Enabled:     true,
				Tags:        tags,
				Response: models.ResponseMock{
					StatusCode:  statusCode,
					ContentType: contentType,
					Headers:     map[string]string{"Content-Type": contentType},
					Body:        bodyStr,
				},
				Latency: models.LatencyConfig{
					Enabled: false,
					Mode:    models.LatencyModeFixed,
					FixedMs: 100,
				},
				Chaos: models.ChaosConfig{
					Enabled:    false,
					Rate:       0.1,
					StatusCode: 500,
				},
				Auth:      authCfg,
				CreatedAt: now,
				UpdatedAt: now,
			}

			endpoints = append(endpoints, ep)
		}
	}

	return endpoints, nil
}

// ExportOpenAPI generates an OpenAPI 3.0 JSON specification from endpoints
func ExportOpenAPI(endpoints []models.Endpoint, title, version string) (map[string]interface{}, error) {
	if title == "" {
		title = "MockForge Simulated API"
	}
	if version == "" {
		version = "1.0.0"
	}

	paths := make(map[string]interface{})

	for _, ep := range endpoints {
		openAPIPath := revertToOpenAPIPath(ep.Path)
		var pathItem map[string]interface{}
		if existing, ok := paths[openAPIPath].(map[string]interface{}); ok {
			pathItem = existing
		} else {
			pathItem = make(map[string]interface{})
			paths[openAPIPath] = pathItem
		}

		method := strings.ToLower(ep.Method)
		statusStr := strconv.Itoa(ep.Response.StatusCode)
		if statusStr == "0" {
			statusStr = "200"
		}

		var parsedBody interface{}
		if err := json.Unmarshal([]byte(ep.Response.Body), &parsedBody); err != nil {
			parsedBody = ep.Response.Body
		}

		op := map[string]interface{}{
			"summary":     ep.Name,
			"description": ep.Description,
			"tags":        ep.Tags,
			"responses": map[string]interface{}{
				statusStr: map[string]interface{}{
					"description": "Mocked response",
					"content": map[string]interface{}{
						ep.Response.ContentType: map[string]interface{}{
							"example": parsedBody,
						},
					},
				},
			},
		}

		if ep.Auth.Type != models.AuthTypeNone && ep.Auth.Type != "" {
			op["security"] = []map[string][]string{
				{ep.Auth.Type: {}},
			}
		}

		pathItem[method] = op
	}

	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       title,
			"version":     version,
			"description": "Exported from MockForge API Studio",
		},
		"paths": paths,
	}

	return spec, nil
}

// ExportWorkspace packages all endpoints and collections into a Workspace struct
func ExportWorkspace(name string, endpoints []models.Endpoint, collections []models.Collection) models.Workspace {
	if name == "" {
		name = "MockForge Workspace"
	}
	return models.Workspace{
		Name:        name,
		Version:     "1.0.0",
		ExportedAt:  time.Now(),
		Endpoints:   endpoints,
		Collections: collections,
	}
}

// ImportWorkspace parses a workspace JSON byte slice
func ImportWorkspace(data []byte) (*models.Workspace, error) {
	var ws models.Workspace
	if err := json.Unmarshal(data, &ws); err != nil {
		return nil, fmt.Errorf("invalid workspace JSON: %w", err)
	}
	return &ws, nil
}

func convertOpenAPIPath(path string) string {
	re := regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)
	return re.ReplaceAllString(path, ":$1")
}

func revertToOpenAPIPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func generateID() string {
	nBig, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("ep_%d", 100000+nBig.Int64())
}

func synthesizeSchemaExample(schema map[string]interface{}) string {
	schemaType, _ := schema["type"].(string)
	switch schemaType {
	case "array":
		return "[\n  {\n    \"id\": \"{{faker.id}}\",\n    \"name\": \"{{faker.name}}\"\n  }\n]"
	case "object":
		props, ok := schema["properties"].(map[string]interface{})
		if !ok || len(props) == 0 {
			return "{\n  \"status\": \"success\"\n}"
		}
		obj := make(map[string]interface{})
		for propName, propRaw := range props {
			pObj, ok := propRaw.(map[string]interface{})
			if !ok {
				obj[propName] = "{{faker.word}}"
				continue
			}
			pType, _ := pObj["type"].(string)
			switch pType {
			case "integer", "number":
				obj[propName] = 100
			case "boolean":
				obj[propName] = true
			case "string":
				lowerName := strings.ToLower(propName)
				if strings.Contains(lowerName, "email") {
					obj[propName] = "{{faker.email}}"
				} else if strings.Contains(lowerName, "name") {
					obj[propName] = "{{faker.name}}"
				} else if strings.Contains(lowerName, "id") || strings.Contains(lowerName, "uuid") {
					obj[propName] = "{{faker.uuid}}"
				} else if strings.Contains(lowerName, "date") {
					obj[propName] = "{{faker.datetime}}"
				} else {
					obj[propName] = "{{faker.sentence}}"
				}
			default:
				obj[propName] = "value"
			}
		}
		b, _ := json.MarshalIndent(obj, "", "  ")
		return string(b)
	default:
		return "{\n  \"message\": \"success\"\n}"
	}
}
