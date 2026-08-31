package models_test

import (
	"testing"
	"time"

	"apimock/pkg/models"
)

func TestEndpointValidate(t *testing.T) {
	t.Run("Valid default endpoint", func(t *testing.T) {
		ep := models.Endpoint{
			ID:   "ep-1",
			Path: "api/users",
		}
		if err := ep.Validate(); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
		if ep.Path != "/api/users" {
			t.Errorf("expected path to start with '/', got %s", ep.Path)
		}
		if ep.Method != models.MethodGET {
			t.Errorf("expected default method GET, got %s", ep.Method)
		}
		if ep.Response.StatusCode != 200 {
			t.Errorf("expected default status code 200, got %d", ep.Response.StatusCode)
		}
		if ep.Auth.Type != models.AuthTypeNone {
			t.Errorf("expected default auth type none, got %s", ep.Auth.Type)
		}
	})

	t.Run("Empty ID or Path error", func(t *testing.T) {
		epNoID := models.Endpoint{Path: "/test"}
		if err := epNoID.Validate(); err == nil {
			t.Error("expected error for empty ID, got nil")
		}

		epNoPath := models.Endpoint{ID: "1", Path: "   "}
		if err := epNoPath.Validate(); err == nil {
			t.Error("expected error for empty Path, got nil")
		}
	})

	t.Run("Sanitize Chaos and Latency bounds", func(t *testing.T) {
		ep := models.Endpoint{
			ID:   "ep-2",
			Path: "/test",
			Chaos: models.ChaosConfig{
				Enabled: true,
				Rate:    1.5,
			},
			Latency: models.LatencyConfig{
				Enabled: true,
				Mode:    models.LatencyModeRandom,
				MinMs:   500,
				MaxMs:   100, // Invalid max < min
			},
		}
		if err := ep.Validate(); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
		if ep.Chaos.Rate != 1.0 {
			t.Errorf("expected chaos rate clamped to 1.0, got %f", ep.Chaos.Rate)
		}
		if ep.Chaos.StatusCode != 500 {
			t.Errorf("expected default chaos status 500, got %d", ep.Chaos.StatusCode)
		}
		if ep.Latency.MaxMs != 500 {
			t.Errorf("expected latency max clamped to min 500, got %d", ep.Latency.MaxMs)
		}
	})
}

func TestCollectionValidate(t *testing.T) {
	t.Run("Valid collection", func(t *testing.T) {
		col := models.Collection{
			ID:        "col-1",
			Name:      " Users ",
			CreatedAt: time.Now(),
		}
		if err := col.Validate(); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
		if col.Name != "users" {
			t.Errorf("expected lowercased trimmed name 'users', got '%s'", col.Name)
		}
		if col.Items == nil {
			t.Error("expected initialized empty items slice, got nil")
		}
	})

	t.Run("Empty ID or Name error", func(t *testing.T) {
		colNoID := models.Collection{Name: "users"}
		if err := colNoID.Validate(); err == nil {
			t.Error("expected error for empty ID, got nil")
		}

		colNoName := models.Collection{ID: "1", Name: "   "}
		if err := colNoName.Validate(); err == nil {
			t.Error("expected error for empty Name, got nil")
		}
	})
}
