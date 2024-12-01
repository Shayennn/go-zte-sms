package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

func startPprofServer(pprofListenAddr string) {
	go func() {
		fmt.Printf("Starting pprof server on %s\n", pprofListenAddr)

		pprofServer := &http.Server{
			Addr:         pprofListenAddr,
			Handler:      http.DefaultServeMux, // Use the default handler for pprof
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting pprof server: %v\n", err)
		}
	}()
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("/getSMS", getSMSHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Add CORS support if needed
	handler := cors.Default().Handler(mux)

	// Start pprof server
	pprofListenAddr := os.Getenv("PPROF_LISTEN_ADDR")
	if pprofListenAddr != "" {
		startPprofServer(pprofListenAddr)
	}

	// Read SERVER_LISTEN_ADDR from environment variables or default to 8080
	listenAddr := os.Getenv("SERVER_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8080"
	}

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Server started on %s", listenAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
