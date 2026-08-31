package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"apimock/pkg/engine"
	"apimock/pkg/web"
)

//go:embed all:web
var embeddedWebFS embed.FS

func main() {
	port := flag.Int("port", 8080, "Port for MockForge server")
	host := flag.String("host", "0.0.0.0", "Host interface to bind")
	storePath := flag.String("store", "", "Path to persist stateful collections JSON (optional)")
	noDemo := flag.Bool("no-demo", false, "Disable pre-loading realistic demo endpoints and collections")
	flag.Parse()

	// Initialize Engine Server
	mockServer := engine.NewServer(*storePath)

	if !*noDemo {
		mockServer.SeedDemoData()
	}

	// Initialize Web Handler with embedded SPA assets
	webHandler := web.NewHandler(embeddedWebFS, mockServer)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      webHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Print Startup Banner
	fmt.Println(`
  __  __            _    ______                  
 |  \/  |          | |  |  ____|                 
 | \  / | ___   ___| | _| |__ ___  _ __ __ _  ___ 
 | |\/| |/ _ \ / __| |/ /  __/ _ \| '__/ _` + "`" + ` |/ _ \
 | |  | | (_) | (__|   <| | | (_) | | | (_| |  __/
 |_|  |_|\___/ \___|_|\_\_|  \___/|_|  \__, |\___|
                                        __/ |     
                                       |___/      `)
	fmt.Println("  ⚡ Instant High-Performance API Mock Studio & Frontend Decoupler")
	fmt.Println("  -------------------------------------------------------------")
	fmt.Printf("  🌐 Web Studio Management: http://localhost:%d\n", *port)
	fmt.Printf("  🚀 Mock Routes & Auto-CRUD: http://localhost:%d/api/...\n", *port)
	fmt.Printf("  📡 Live Traffic SSE Stream: http://localhost:%d/api/admin/traffic/stream\n", *port)
	fmt.Println("  -------------------------------------------------------------")
	fmt.Println("  Press CTRL+C to gracefully shutdown.")
	fmt.Println()

	// Run server in goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\n  Shutting down MockForge Studio gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	fmt.Println("  MockForge stopped. Bye!")
}
