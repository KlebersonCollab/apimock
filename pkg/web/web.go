package web

import (
	"io/fs"
	"net/http"
	"strings"
)

// Handler serves static assets from embedded filesystem and routes API requests to the mock server
type Handler struct {
	fileServer http.Handler
	mockServer http.Handler
	staticFS   fs.FS
}

// NewHandler creates a web handler wrapping embedded static files and mock server
func NewHandler(staticFS fs.FS, mockServer http.Handler) *Handler {
	var subFS fs.FS
	if staticFS != nil {
		// If fs has 'web' subfolder, strip it
		if sub, err := fs.Sub(staticFS, "web"); err == nil {
			subFS = sub
		} else {
			subFS = staticFS
		}
	}

	var fServer http.Handler
	if subFS != nil {
		fServer = http.FileServer(http.FS(subFS))
	}

	return &Handler{
		fileServer: fServer,
		mockServer: mockServer,
		staticFS:   subFS,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 1. All API and resource requests are forwarded to Mock Engine
	if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/resources/") {
		h.mockServer.ServeHTTP(w, r)
		return
	}

	// 2. Static file serving (CSS, JS, assets)
	if h.staticFS != nil {
		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		// Check if exact static file exists
		if _, err := fs.Stat(h.staticFS, cleanPath); err == nil {
			h.fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for Single-Page App (SPA) client-side routes
		if !strings.Contains(path, ".") {
			r.URL.Path = "/"
			h.fileServer.ServeHTTP(w, r)
			return
		}
	}

	// 3. Fallback to mock server
	h.mockServer.ServeHTTP(w, r)
}
