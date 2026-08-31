package openapi_test

import (
	"encoding/json"
	"testing"
	"time"

	"apimock/pkg/models"
	"apimock/pkg/openapi"
)

func TestOpenAPIImportExport(t *testing.T) {
	openAPISample := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Pet Store Mock",
    "version": "1.0.0"
  },
  "paths": {
    "/pets": {
      "get": {
        "summary": "List all pets",
        "tags": ["pets"],
        "responses": {
          "200": {
            "description": "A paged array of pets",
            "content": {
              "application/json": {
                "example": [
                  {"id": 1, "name": "Fido", "tag": "dog"},
                  {"id": 2, "name": "Whiskers", "tag": "cat"}
                ]
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a pet",
        "tags": ["pets"],
        "security": [{"BearerAuth": []}],
        "responses": {
          "201": {
            "description": "Null response"
          }
        }
      }
    },
    "/pets/{petId}": {
      "get": {
        "summary": "Info for a specific pet",
        "tags": ["pets"],
        "responses": {
          "200": {
            "description": "Expected response to a valid request",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "id": {"type": "integer"},
                    "name": {"type": "string"},
                    "ownerEmail": {"type": "string"}
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

	t.Run("Import OpenAPI Specification", func(t *testing.T) {
		endpoints, err := openapi.ImportOpenAPI([]byte(openAPISample))
		if err != nil {
			t.Fatalf("failed to import OpenAPI: %v", err)
		}

		if len(endpoints) != 3 {
			t.Fatalf("expected 3 endpoints (GET /pets, POST /pets, GET /pets/:petId), got %d", len(endpoints))
		}

		// Verify GET /pets
		var getPets *models.Endpoint
		var getPetById *models.Endpoint
		var postPet *models.Endpoint

		for i := range endpoints {
			if endpoints[i].Method == "GET" && endpoints[i].Path == "/pets" {
				getPets = &endpoints[i]
			}
			if endpoints[i].Method == "GET" && endpoints[i].Path == "/pets/:petId" {
				getPetById = &endpoints[i]
			}
			if endpoints[i].Method == "POST" && endpoints[i].Path == "/pets" {
				postPet = &endpoints[i]
			}
		}

		if getPets == nil {
			t.Fatal("GET /pets endpoint not found")
		}
		if getPets.Response.StatusCode != 200 {
			t.Errorf("expected GET /pets status 200, got %d", getPets.Response.StatusCode)
		}

		if getPetById == nil {
			t.Fatal("GET /pets/:petId endpoint not found")
		}
		if getPetById.Path != "/pets/:petId" {
			t.Errorf("expected path parameter converted to :petId, got %s", getPetById.Path)
		}

		if postPet == nil {
			t.Fatal("POST /pets endpoint not found")
		}
		if postPet.Response.StatusCode != 201 {
			t.Errorf("expected POST /pets status 201, got %d", postPet.Response.StatusCode)
		}
		if postPet.Auth.Type != models.AuthTypeBearer {
			t.Errorf("expected POST /pets bearer auth, got %s", postPet.Auth.Type)
		}
	})

	t.Run("Export OpenAPI Specification", func(t *testing.T) {
		eps := []models.Endpoint{
			{
				ID:     "ep-1",
				Name:   "Get User",
				Method: "GET",
				Path:   "/api/v1/users/:id",
				Response: models.ResponseMock{
					StatusCode:  200,
					ContentType: "application/json",
					Body:        `{"id":"1","name":"Alice"}`,
				},
			},
		}

		spec, err := openapi.ExportOpenAPI(eps, "Test API", "2.0.0")
		if err != nil {
			t.Fatalf("failed to export OpenAPI: %v", err)
		}

		paths := spec["paths"].(map[string]interface{})
		if _, ok := paths["/api/v1/users/{id}"]; !ok {
			t.Errorf("expected reverted path /api/v1/users/{id}, got paths: %+v", paths)
		}
	})

	t.Run("Workspace Export and Import", func(t *testing.T) {
		eps := []models.Endpoint{
			{ID: "ep-1", Name: "Ping", Method: "GET", Path: "/ping"},
		}
		cols := []models.Collection{
			{ID: "col-1", Name: "notes", CreatedAt: time.Now()},
		}

		ws := openapi.ExportWorkspace("Demo WS", eps, cols)
		wsBytes, err := json.Marshal(ws)
		if err != nil {
			t.Fatalf("failed to marshal workspace: %v", err)
		}

		imported, err := openapi.ImportWorkspace(wsBytes)
		if err != nil {
			t.Fatalf("failed to import workspace: %v", err)
		}

		if imported.Name != "Demo WS" || len(imported.Endpoints) != 1 || len(imported.Collections) != 1 {
			t.Errorf("unexpected imported workspace: %+v", imported)
		}
	})
}
