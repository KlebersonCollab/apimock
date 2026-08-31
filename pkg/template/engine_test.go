package template_test

import (
	"encoding/json"
	"strings"
	"testing"

	"apimock/pkg/template"
)

func TestFakerMethods(t *testing.T) {
	f := template.NewFaker()

	if f.Name() == "" {
		t.Error("expected non-empty name")
	}
	if !strings.Contains(f.Email(), "@") {
		t.Errorf("expected valid email with @, got %s", f.Email())
	}
	if len(f.UUID()) != 36 {
		t.Errorf("expected standard 36-char UUID, got %s", f.UUID())
	}
	num := f.Number(10, 20)
	if num < 10 || num > 20 {
		t.Errorf("expected number between 10 and 20, got %d", num)
	}
	price := f.Price(5.0, 10.0)
	if price < 5.0 || price > 10.0 {
		t.Errorf("expected price between 5.0 and 10.0, got %f", price)
	}
}

func TestTemplateEngineEvaluation(t *testing.T) {
	eng := template.NewEngine()

	ctx := &template.RequestContext{
		Params: map[string]string{
			"id": "101",
		},
		Query: map[string][]string{
			"filter": {"active"},
		},
		Headers: map[string][]string{
			"Authorization": {"Bearer abc-123"},
		},
		Body:   `{"name":"Alice","role":"admin","profile":{"age":30}}`,
		Method: "POST",
		Path:   "/api/v1/users/101",
	}

	t.Run("Interpolate request params, query, headers and body", func(t *testing.T) {
		tpl := `{"userId":"{{req.params.id}}","filter":"{{req.query.filter}}","auth":"{{req.headers.authorization}}","bodyName":"{{req.body.name}}","bodyAge":{{req.body.profile.age}},"method":"{{req.method}}"}`
		evaluated := eng.Evaluate(tpl, ctx)

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(evaluated), &parsed); err != nil {
			t.Fatalf("failed to unmarshal evaluated JSON: %v, raw:\n%s", err, evaluated)
		}

		if parsed["userId"] != "101" {
			t.Errorf("expected userId 101, got %v", parsed["userId"])
		}
		if parsed["filter"] != "active" {
			t.Errorf("expected filter active, got %v", parsed["filter"])
		}
		if parsed["auth"] != "Bearer abc-123" {
			t.Errorf("expected auth Bearer abc-123, got %v", parsed["auth"])
		}
		if parsed["bodyName"] != "Alice" {
			t.Errorf("expected bodyName Alice, got %v", parsed["bodyName"])
		}
		if parsed["bodyAge"] != float64(30) {
			t.Errorf("expected bodyAge 30, got %v", parsed["bodyAge"])
		}
		if parsed["method"] != "POST" {
			t.Errorf("expected method POST, got %v", parsed["method"])
		}
	})

	t.Run("Interpolate faker tags", func(t *testing.T) {
		tpl := `{"name":"{{faker.name}}","email":"{{faker.email}}","uuid":"{{faker.uuid}}","price":{{faker.price(10, 50)}}}`
		evaluated := eng.Evaluate(tpl, ctx)

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(evaluated), &parsed); err != nil {
			t.Fatalf("failed to unmarshal evaluated JSON: %v, raw:\n%s", err, evaluated)
		}

		name, _ := parsed["name"].(string)
		if name == "" || strings.Contains(name, "{{") {
			t.Errorf("expected evaluated name, got %s", name)
		}
		email, _ := parsed["email"].(string)
		if !strings.Contains(email, "@") {
			t.Errorf("expected evaluated email, got %s", email)
		}
	})

	t.Run("Process repeat array blocks", func(t *testing.T) {
		tpl := `[
  {{#repeat 3}}
  {
    "id": {{@iteration}},
    "index": {{@index}},
    "name": "{{faker.name}}",
    "email": "{{faker.email}}"
  }
  {{/repeat}}
]`
		evaluated := eng.Evaluate(tpl, ctx)

		var list []map[string]interface{}
		if err := json.Unmarshal([]byte(evaluated), &list); err != nil {
			t.Fatalf("failed to unmarshal repeat evaluated JSON: %v, raw:\n%s", err, evaluated)
		}

		if len(list) != 3 {
			t.Fatalf("expected 3 items, got %d", len(list))
		}
		if list[0]["id"] != float64(1) || list[0]["index"] != float64(0) {
			t.Errorf("expected item 0 to have id 1, index 0, got %v, %v", list[0]["id"], list[0]["index"])
		}
		if list[2]["id"] != float64(3) || list[2]["index"] != float64(2) {
			t.Errorf("expected item 2 to have id 3, index 2, got %v, %v", list[2]["id"], list[2]["index"])
		}
	})
}
