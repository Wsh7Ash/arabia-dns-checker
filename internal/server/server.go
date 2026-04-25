package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"arabia-dns-checker/internal/storage"
	"arabia-dns-checker/pkg/dns"
	"arabia-dns-checker/pkg/ping"
	"arabia-dns-checker/pkg/traceroute"

	"github.com/gorilla/mux"
)

type Server struct {
	db     *storage.Database
	config *Config
}

type Config struct {
	Port        int
	DatabaseURL string
	RedisURL    string
	CheckInterval time.Duration
	Timeout      time.Duration
	RetryCount   int
}

func LoadConfig() *Config {
	return &Config{
		Port:        8080,
		DatabaseURL: "postgresql://user:pass@localhost/arabia_dns",
		RedisURL:    "redis://localhost:6379",
		CheckInterval: 30 * time.Second,
		Timeout:      10 * time.Second,
		RetryCount:   3,
	}
}

func NewServer(db *storage.Database, config *Config) *Server {
	return &Server{
		db:     db,
		config: config,
	}
}

func (s *Server) SetupRoutes(router *mux.Router) {
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Monitoring endpoints
	api.HandleFunc("/check/website", s.CheckWebsite).Methods("POST")
	api.HandleFunc("/status/nodes", s.GetNodeStatus).Methods("GET")
	api.HandleFunc("/history/website", s.GetHistory).Methods("GET")
	
	// Network analysis endpoints
	api.HandleFunc("/network/topology", s.GetTopology).Methods("GET")
	api.HandleFunc("/network/compare", s.CompareRouting).Methods("GET")
	api.HandleFunc("/dns/analyze", s.AnalyzeDNS).Methods("GET")
	
	// WebSocket for real-time updates
	router.HandleFunc("/ws/alerts", s.AlertWebSocket)
}

func (s *Server) CheckWebsite(w http.ResponseWriter, r *http.Request) {
	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	results := make(map[string]NodeResult)
	
	for _, node := range req.Nodes {
		result := s.performNodeCheck(req.URL, node, req.Checks)
		results[node] = result
	}
	
	response := CheckResponse{
		CheckID:   generateCheckID(),
		URL:       req.URL,
		Timestamp: time.Now().UTC(),
		Results:   results,
		Summary:   s.generateSummary(results),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) performNodeCheck(url string, node string, checks []string) NodeResult {
	result := NodeResult{
		Node: node,
	}
	
	for _, check := range checks {
		switch check {
		case "ping":
			result.Ping = ping.Ping(url, s.config.Timeout)
		case "dns":
			result.DNS = dns.Resolve(url, s.config.Timeout)
		case "http":
			result.HTTP = s.checkHTTP(url, s.config.Timeout)
		case "traceroute":
			result.Traceroute = traceroute.Trace(url, s.config.Timeout)
		}
	}
	
	return result
}

func (s *Server) checkHTTP(url string, timeout time.Duration) HTTPResult {
	client := &http.Client{
		Timeout: timeout,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return HTTPResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()
	
	return HTTPResult{
		Success:     true,
		StatusCode:  resp.StatusCode,
		ResponseTime: time.Since(time.Now()).Milliseconds(),
	}
}

func (s *Server) generateSummary(results map[string]NodeResult) CheckSummary {
	total := len(results)
	accessible := 0
	blocked := 0
	var totalTime int64
	
	for _, result := range results {
		if result.HTTP.Success {
			accessible++
			totalTime += result.HTTP.ResponseTime
		} else {
			blocked++
		}
	}
	
	avgTime := int64(0)
	if accessible > 0 {
		avgTime = totalTime / int64(accessible)
	}
	
	return CheckSummary{
		TotalNodes:         total,
		AccessibleNodes:    accessible,
		BlockedNodes:       blocked,
		AverageResponseTime: avgTime,
	}
}

func (s *Server) GetNodeStatus(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.db.GetAllNodes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

func (s *Server) GetHistory(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	period := r.URL.Query().Get("period")
	
	history, err := s.db.GetHistory(domain, period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (s *Server) GetTopology(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	target := r.URL.Query().Get("target")
	
	topology, err := s.db.GetTopology(source, target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topology)
}

func (s *Server) CompareRouting(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	isps := r.URL.Query()["isps"]
	
	comparison, err := s.db.CompareRouting(domain, isps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

func (s *Server) AnalyzeDNS(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	nodes := r.URL.Query().Get("nodes")
	
	analysis, err := s.db.AnalyzeDNS(domain, nodes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

func (s *Server) AlertWebSocket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket for real-time alerts
	log.Println("WebSocket connection requested")
}

func generateCheckID() string {
	return fmt.Sprintf("chk_%d", time.Now().UnixNano())
}

type CheckRequest struct {
	URL    string   `json:"url"`
	Nodes  []string `json:"nodes"`
	Checks []string `json:"checks"`
}

type CheckResponse struct {
	CheckID   string               `json:"check_id"`
	URL       string               `json:"url"`
	Timestamp time.Time            `json:"timestamp"`
	Results   map[string]NodeResult `json:"results"`
	Summary   CheckSummary         `json:"summary"`
}

type NodeResult struct {
	Node       string          `json:"node"`
	Status     string          `json:"status"`
	Ping       ping.Result     `json:"ping,omitempty"`
	DNS        dns.Result      `json:"dns,omitempty"`
	HTTP       HTTPResult      `json:"http,omitempty"`
	Traceroute traceroute.Result `json:"traceroute,omitempty"`
}

type HTTPResult struct {
	Success      bool   `json:"success"`
	StatusCode   int    `json:"status_code,omitempty"`
	ResponseTime int64  `json:"response_time,omitempty"`
	Error        string `json:"error,omitempty"`
}

type CheckSummary struct {
	TotalNodes          int    `json:"total_nodes"`
	AccessibleNodes     int    `json:"accessible_nodes"`
	BlockedNodes        int    `json:"blocked_nodes"`
	AverageResponseTime int64  `json:"average_response_time"`
}
