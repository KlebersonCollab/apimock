package traffic_test

import (
	"fmt"
	"testing"
	"time"

	"apimock/pkg/models"
	"apimock/pkg/traffic"
)

func TestTrafficLogger(t *testing.T) {
	t.Run("Circular buffer capacity and ordering", func(t *testing.T) {
		logger := traffic.NewTrafficLogger(3)

		logger.Log(models.TrafficEntry{ID: "req-1", Path: "/api/1", ResponseStatus: 200, DurationMs: 10})
		logger.Log(models.TrafficEntry{ID: "req-2", Path: "/api/2", ResponseStatus: 200, DurationMs: 20})
		logger.Log(models.TrafficEntry{ID: "req-3", Path: "/api/3", ResponseStatus: 500, DurationMs: 30})
		logger.Log(models.TrafficEntry{ID: "req-4", Path: "/api/4", ResponseStatus: 201, DurationMs: 40}) // Overwrites req-1

		history := logger.GetHistory(10)
		if len(history) != 3 {
			t.Fatalf("expected capacity 3 entries, got %d", len(history))
		}

		// Newest first
		if history[0].ID != "req-4" || history[1].ID != "req-3" || history[2].ID != "req-2" {
			t.Errorf("unexpected history order: [%s, %s, %s]", history[0].ID, history[1].ID, history[2].ID)
		}

		// Lookup by ID
		entry, found := logger.GetEntry("req-3")
		if !found || entry.Path != "/api/3" {
			t.Errorf("expected to find req-3, got found=%v, entry=%+v", found, entry)
		}

		_, foundOld := logger.GetEntry("req-1")
		if foundOld {
			t.Error("expected overwritten req-1 to not be found")
		}

		// Stats
		stats := logger.Stats()
		if stats.TotalRequests != 4 {
			t.Errorf("expected total requests 4, got %d", stats.TotalRequests)
		}
		if stats.ErrorRequests != 1 {
			t.Errorf("expected error requests 1 (status 500), got %d", stats.ErrorRequests)
		}
		if stats.SuccessRequests != 3 {
			t.Errorf("expected success requests 3, got %d", stats.SuccessRequests)
		}
	})

	t.Run("SSE Pub/Sub Broadcast", func(t *testing.T) {
		logger := traffic.NewTrafficLogger(10)
		ch, unsub := logger.Subscribe()
		defer unsub()

		testEntry := models.TrafficEntry{
			ID:             "req-live-1",
			Path:           "/api/realtime",
			ResponseStatus: 200,
		}

		logger.Log(testEntry)

		select {
		case received := <-ch:
			if received.ID != "req-live-1" {
				t.Errorf("expected received entry ID 'req-live-1', got '%s'", received.ID)
			}
		case <-time.After(500 * time.Millisecond):
			t.Error("timed out waiting for broadcast event")
		}
	})

	t.Run("Clear logger", func(t *testing.T) {
		logger := traffic.NewTrafficLogger(5)
		for i := 0; i < 5; i++ {
			logger.Log(models.TrafficEntry{ID: fmt.Sprintf("req-%d", i)})
		}
		if len(logger.GetHistory(10)) != 5 {
			t.Errorf("expected 5 items before clear")
		}

		logger.Clear()
		if len(logger.GetHistory(10)) != 0 {
			t.Errorf("expected 0 items after clear")
		}
	})
}
