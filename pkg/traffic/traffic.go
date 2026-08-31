package traffic

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"apimock/pkg/models"
)

// TrafficStats captures aggregate request metrics
type TrafficStats struct {
	TotalRequests     int64   `json:"totalRequests"`
	SuccessRequests   int64   `json:"successRequests"`
	ErrorRequests     int64   `json:"errorRequests"`
	AvgDurationMs     float64 `json:"avgDurationMs"`
	AvgSimulatedDelay float64 `json:"avgSimulatedDelay"`
}

// TrafficLogger implements a thread-safe circular buffer with SSE pub/sub
type TrafficLogger struct {
	mu          sync.RWMutex
	entries     []models.TrafficEntry
	capacity    int
	head        int
	count       int
	totalLogged int64
	totalErrors int64
	totalDurMs  int64
	totalDelay  int64

	subMu       sync.RWMutex
	subscribers map[chan models.TrafficEntry]struct{}
}

// NewTrafficLogger creates a new logger with circular buffer capacity
func NewTrafficLogger(capacity int) *TrafficLogger {
	if capacity <= 0 {
		capacity = 200
	}
	return &TrafficLogger{
		entries:     make([]models.TrafficEntry, capacity),
		capacity:    capacity,
		subscribers: make(map[chan models.TrafficEntry]struct{}),
	}
}

// Log records a new traffic entry and broadcasts to SSE subscribers
func (tl *TrafficLogger) Log(entry models.TrafficEntry) {
	if entry.ID == "" {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(900000))
		entry.ID = fmt.Sprintf("req_%d", 100000+nBig.Int64())
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	tl.mu.Lock()
	tl.entries[tl.head] = entry
	tl.head = (tl.head + 1) % tl.capacity
	if tl.count < tl.capacity {
		tl.count++
	}
	tl.totalLogged++
	if entry.ResponseStatus >= 400 {
		tl.totalErrors++
	}
	tl.totalDurMs += entry.DurationMs
	tl.totalDelay += entry.SimulatedDelayMs
	tl.mu.Unlock()

	// Broadcast to active SSE subscribers (non-blocking)
	tl.subMu.RLock()
	defer tl.subMu.RUnlock()
	for ch := range tl.subscribers {
		select {
		case ch <- entry:
		default:
			// Subscriber buffer full, skip to avoid stalling pipeline
		}
	}
}

// GetHistory returns the most recent traffic entries (newest first)
func (tl *TrafficLogger) GetHistory(limit int) []models.TrafficEntry {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	if limit <= 0 || limit > tl.count {
		limit = tl.count
	}

	res := make([]models.TrafficEntry, 0, limit)
	for i := 0; i < limit; i++ {
		idx := (tl.head - 1 - i + tl.capacity) % tl.capacity
		res = append(res, tl.entries[idx])
	}
	return res
}

// GetEntry searches for a traffic log by ID
func (tl *TrafficLogger) GetEntry(id string) (models.TrafficEntry, bool) {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	for i := 0; i < tl.count; i++ {
		idx := (tl.head - 1 - i + tl.capacity) % tl.capacity
		if tl.entries[idx].ID == id {
			return tl.entries[idx], true
		}
	}
	return models.TrafficEntry{}, false
}

// Clear resets the log buffer
func (tl *TrafficLogger) Clear() {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	tl.entries = make([]models.TrafficEntry, tl.capacity)
	tl.head = 0
	tl.count = 0
	tl.totalLogged = 0
	tl.totalErrors = 0
	tl.totalDurMs = 0
	tl.totalDelay = 0
}

// Subscribe registers an SSE channel listener
func (tl *TrafficLogger) Subscribe() (chan models.TrafficEntry, func()) {
	tl.subMu.Lock()
	ch := make(chan models.TrafficEntry, 50)
	tl.subscribers[ch] = struct{}{}
	tl.subMu.Unlock()

	unsubscribe := func() {
		tl.subMu.Lock()
		delete(tl.subscribers, ch)
		close(ch)
		tl.subMu.Unlock()
	}

	return ch, unsubscribe
}

// Stats returns aggregate metrics
func (tl *TrafficLogger) Stats() TrafficStats {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	stats := TrafficStats{
		TotalRequests:   tl.totalLogged,
		ErrorRequests:   tl.totalErrors,
		SuccessRequests: tl.totalLogged - tl.totalErrors,
	}

	if tl.totalLogged > 0 {
		stats.AvgDurationMs = float64(tl.totalDurMs) / float64(tl.totalLogged)
		stats.AvgSimulatedDelay = float64(tl.totalDelay) / float64(tl.totalLogged)
	}

	return stats
}
