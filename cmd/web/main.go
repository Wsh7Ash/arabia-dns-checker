package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

//go:embed static templates
var content embed.FS

func main() {
	port := os.Getenv("WEB_PORT")
	if port == "" {
		port = "8081"
	}
	
	router := setupRoutes()
	
	log.Printf("Starting Arabia DNS Checker web dashboard on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func setupRoutes() *mux.Router {
	router := mux.NewRouter()
	
	// Static files
	router.PathPrefix("/static/").Handler(http.FileServer(http.FS(content)))
	
	// Web routes
	router.HandleFunc("/", dashboardHandler).Methods("GET")
	router.HandleFunc("/nodes", nodesHandler).Methods("GET")
	router.HandleFunc("/history", historyHandler).Methods("GET")
	router.HandleFunc("/alerts", alertsHandler).Methods("GET")
	
	return router
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Serve dashboard template
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Arabia DNS Checker - Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
        <div class="container">
            <a class="navbar-brand" href="/">
                <i class="fas fa-globe"></i> Arabia DNS Checker
            </a>
        </div>
    </nav>
    
    <div class="container mt-4">
        <div class="row">
            <div class="col-md-3">
                <div class="card">
                    <div class="card-body text-center">
                        <h5 class="card-title">Active Nodes</h5>
                        <h2 class="text-primary">5</h2>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card">
                    <div class="card-body text-center">
                        <h5 class="card-title">Monitored Sites</h5>
                        <h2 class="text-success">42</h2>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card">
                    <div class="card-body text-center">
                        <h5 class="card-title">Active Alerts</h5>
                        <h2 class="text-danger">3</h2>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card">
                    <div class="card-body text-center">
                        <h5 class="card-title">Avg Response</h5>
                        <h2 class="text-info">127ms</h2>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="row mt-4">
            <div class="col-md-12">
                <div class="card">
                    <div class="card-header">
                        <h5 class="card-title">Real-time Node Status</h5>
                    </div>
                    <div class="card-body">
                        <div id="node-status">Loading...</div>
                    </div>
                </div>
            </div>
        </div>
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        // Load node status
        fetch('/api/v1/status/nodes')
            .then(response => response.json())
            .then(data => {
                document.getElementById('node-status').innerHTML = JSON.stringify(data, null, 2);
            });
    </script>
</body>
</html>
	`)
}

func nodesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Nodes Page</h1><p>Monitoring nodes status and configuration.</p>")
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>History Page</h1><p>Historical check results and analytics.</p>")
}

func alertsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Alerts Page</h1><p>Active alerts and notifications.</p>")
}
