package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	server *http.Server
	logger *log.Logger
}

func New(addr string, handler http.Handler) *App {
	return &App{
		server: &http.Server{Addr: addr, Handler: handler},
		logger: log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Start starts the server asynchronously and wait for termination
func (a *App) Start() {
	// starts server asynchronously
	go func() {
		a.logger.Printf("Server started at port %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != http.ErrServerClosed {
			a.logger.Fatalf("ListenAndServe: %s", err)
		}
	}()

	// Handle graceful shutdown
	// Channel to listen for an interrupt or terminate signal from the OS.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	// Block waiting for a receive on signal from OS
	s := <-osSignals
	switch s {
	case syscall.SIGTERM:
		d := 10 * time.Second
		a.logger.Printf("SIGTERM received. Sleeping for %s as buffer before stopping server", d)
		// Delay 10 seconds as buffer
		time.Sleep(d)
	}

	// Shutdown gracefully
	a.Stop()
}

// Stop stops the app
func (a *App) Stop() {
	// Create a context to attempt a graceful 5 second shutdown.
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Attempt the graceful shutdown by closing the listener and
	// completing all inflight requests.
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Printf("Could not stop server gracefully: %v", err)
		a.logger.Printf("Initiating hard shutdown")
		if err := a.server.Close(); err != nil {
			a.logger.Printf("Could not stop http server: %v", err)
		}
	}
}
