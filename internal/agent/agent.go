package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"arabia-dns-checker/pkg/dns"
	"arabia-dns-checker/pkg/ping"
	"arabia-dns-checker/pkg/traceroute"
)

type Agent struct {
	config    *Config
	client    *http.Client
	stopChan  chan struct{}
	targets   []Target
}

type Config struct {
	ServerURL     string
	NodeName      string
	NodeLocation  string
	NodeISP       string
	CheckInterval time.Duration
	Timeout       time.Duration
	RetryCount    int
	HealthPort    int
}

type Target struct {
	Domain    string
	CheckType []string
	Interval  time.Duration
}

type CheckResult struct {
	CheckID   string      `json:"check_id"`
	Node      string      `json:"node"`
	URL       string      `json:"url"`
	Status    string      `json:"status"`
	Ping      ping.Result `json:"ping,omitempty"`
	DNS       dns.Result  `json:"dns,omitempty"`
	HTTP      HTTPResult  `json:"http,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type HTTPResult struct {
	Success      bool   `json:"success"`
	StatusCode   int    `json:"status_code,omitempty"`
	ResponseTime int64  `json:"response_time,omitempty"`
	Error        string `json:"error,omitempty"`
}

func NewAgent(config *Config) *Agent {
	return &Agent{
		config:   config,
		client:   &http.Client{Timeout: 30 * time.Second},
		stopChan: make(chan struct{}),
		targets:  loadTargets(),
	}
}

func (a *Agent) RunMonitoring() {
	ticker := time.NewTicker(a.config.CheckInterval)
	defer ticker.Stop()
	
	log.Printf("Starting monitoring for node: %s (%s)", a.config.NodeName, a.config.NodeLocation)
	
	for {
		select {
		case <-ticker.C:
			a.performChecks()
		case <-a.stopChan:
			return
		}
	}
}

func (a *Agent) performChecks() {
	for _, target := range a.targets {
		result := a.checkTarget(target)
		a.sendResult(result)
	}
}

func (a *Agent) checkTarget(target Target) CheckResult {
	result := CheckResult{
		CheckID:   generateCheckID(),
		Node:      a.config.NodeName,
		URL:       target.Domain,
		Timestamp: time.Now().UTC(),
	}
	
	for _, checkType := range target.CheckType {
		switch checkType {
		case "ping":
			result.Ping = ping.Ping(target.Domain, a.config.Timeout)
		case "dns":
			result.DNS = dns.Resolve(target.Domain, a.config.Timeout)
		case "http":
			result.HTTP = a.checkHTTP(target.Domain)
		case "traceroute":
			// Traceroute is expensive, run sparingly
			if time.Now().Minute()%5 == 0 {
				tracerouteResult := traceroute.Trace(target.Domain, a.config.Timeout)
				// Handle traceroute result
				log.Printf("Traceroute for %s: %d hops", target.Domain, tracerouteResult.Hops)
			}
		}
	}
	
	// Determine overall status
	if result.HTTP.Success {
		result.Status = "accessible"
	} else {
		result.Status = "unreachable"
	}
	
	return result
}

func (a *Agent) checkHTTP(url string) HTTPResult {
	start := time.Now()
	resp, err := a.client.Get(url)
	if err != nil {
		return HTTPResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()
	
	return HTTPResult{
		Success:      resp.StatusCode >= 200 && resp.StatusCode < 400,
		StatusCode:   resp.StatusCode,
		ResponseTime: time.Since(start).Milliseconds(),
	}
}

func (a *Agent) sendResult(result CheckResult) {
	payload := map[string]interface{}{
		"node":      result.Node,
		"check_id":  result.CheckID,
		"url":       result.URL,
		"status":    result.Status,
		"ping":      result.Ping,
		"dns":       result.DNS,
		"http":      result.HTTP,
		"timestamp": result.Timestamp,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling result: %v", err)
		return
	}
	
	resp, err := a.client.Post(
		fmt.Sprintf("%s/api/v1/results", a.config.ServerURL),
		"application/json",
		jsonData,
	)
	if err != nil {
		log.Printf("Error sending result: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("Server returned status: %d", resp.StatusCode)
	}
}

func (a *Agent) Stop(ctx context.Context) {
	close(a.stopChan)
	
	select {
	case <-ctx.Done():
		log.Println("Agent shutdown timeout")
	case <-time.After(5 * time.Second):
		log.Println("Agent stopped gracefully")
	}
}

func loadTargets() []Target {
	return []Target{
		{
			Domain:    "https://google.com",
			CheckType: []string{"ping", "dns", "http"},
			Interval:  30 * time.Second,
		},
		{
			Domain:    "https://local-website.sa",
			CheckType: []string{"ping", "dns", "http", "ssl"},
			Interval:  60 * time.Second,
		},
		{
			Domain:    "https://government-portal.jo",
			CheckType: []string{"ping", "dns", "http"},
			Interval:  120 * time.Second,
		},
	}
}

func generateCheckID() string {
	return fmt.Sprintf("chk_%d", time.Now().UnixNano())
}
