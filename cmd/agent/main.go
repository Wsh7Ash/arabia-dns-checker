package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"arabia-dns-checker/internal/agent"
	"arabia-dns-checker/pkg/dns"
	"arabia-dns-checker/pkg/ping"
	"arabia-dns-checker/pkg/traceroute"
)

func main() {
	// Load configuration
	config := loadConfig()
	
	// Initialize agent
	a := agent.NewAgent(config)
	
	// Start monitoring
	go a.RunMonitoring()
	
	// Setup health check endpoint
	http.HandleFunc("/health", healthHandler)
	go func() {
		log.Printf("Agent health check on :%d", config.HealthPort)
		http.ListenAndServe(fmt.Sprintf(":%d", config.HealthPort), nil)
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down agent...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	a.Stop(ctx)
	log.Println("Agent stopped")
}

func loadConfig() *agent.Config {
	return &agent.Config{
		ServerURL:      "http://localhost:8080",
		NodeName:       os.Getenv("NODE_NAME"),
		NodeLocation:   os.Getenv("NODE_LOCATION"),
		NodeISP:        os.Getenv("NODE_ISP"),
		CheckInterval:  30 * time.Second,
		Timeout:        10 * time.Second,
		RetryCount:     3,
		HealthPort:     9001,
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}
