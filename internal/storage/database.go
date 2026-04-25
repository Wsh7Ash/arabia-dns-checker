package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

type Node struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Location  string    `json:"location"`
	ISP       string    `json:"isp"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CheckResult struct {
	ID        int       `json:"id"`
	CheckID   string    `json:"check_id"`
	Node      string    `json:"node"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	Ping      json.RawMessage `json:"ping"`
	DNS       json.RawMessage `json:"dns"`
	HTTP      json.RawMessage `json:"http"`
	CreatedAt time.Time `json:"created_at"`
}

type Topology struct {
	ID          int       `json:"id"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	ISP         string    `json:"isp"`
	ASPath      []int     `json:"as_path"`
	GeoPath     []GeoPoint `json:"geographic_path"`
	NetworkHops []Hop     `json:"network_hops"`
	CreatedAt   time.Time `json:"created_at"`
}

type GeoPoint struct {
	City    string  `json:"city"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

type Hop struct {
	Hop          int    `json:"hop"`
	IP           string `json:"ip"`
	Description  string `json:"description"`
}

func NewDatabase(databaseURL string) (*Database, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetAllNodes() ([]Node, error) {
	rows, err := d.db.Query("SELECT id, name, host, port, location, isp, status, created_at FROM nodes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var nodes []Node
	for rows.Next() {
		var node Node
		if err := rows.Scan(&node.ID, &node.Name, &node.Host, &node.Port, &node.Location, &node.ISP, &node.Status, &node.CreatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	
	return nodes, nil
}

func (d *Database) GetHistory(domain string, period string) ([]CheckResult, error) {
	var results []CheckResult
	
	query := `SELECT id, check_id, node, url, status, ping, dns, http, created_at 
	          FROM check_results 
	          WHERE url LIKE $1 
	          ORDER BY created_at DESC 
	          LIMIT 100`
	
	rows, err := d.db.Query(query, "%"+domain+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var result CheckResult
		if err := rows.Scan(&result.ID, &result.CheckID, &result.Node, &result.URL, &result.Status, &result.Ping, &result.DNS, &result.HTTP, &result.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	
	return results, nil
}

func (d *Database) GetTopology(source string, target string) (*Topology, error) {
	var topology Topology
	
	query := `SELECT id, source, target, isp, as_path, geographic_path, network_hops, created_at 
	          FROM topologies 
	          WHERE source = $1 AND target = $2 
	          ORDER BY created_at DESC 
	          LIMIT 1`
	
	var asPathJSON, geoPathJSON, hopsJSON []byte
	if err := d.db.QueryRow(query, source, target).Scan(
		&topology.ID, &topology.Source, &topology.Target, &topology.ISP,
		&asPathJSON, &geoPathJSON, &hopsJSON, &topology.CreatedAt,
	); err != nil {
		return nil, err
	}
	
	json.Unmarshal(asPathJSON, &topology.ASPath)
	json.Unmarshal(geoPathJSON, &topology.GeoPath)
	json.Unmarshal(hopsJSON, &topology.NetworkHops)
	
	return &topology, nil
}

func (d *Database) CompareRouting(domain string, isps []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	
	for _, isp := range isps {
		var latency float64
		var path string
		
		query := `SELECT AVG(latency), array_agg(hop_ip ORDER BY hop_number) 
		          FROM routing_data 
		          WHERE isp = $1 AND target_domain LIKE $2 
		          GROUP BY isp`
		
		if err := d.db.QueryRow(query, isp, "%"+domain+"%").Scan(&latency, &path); err != nil {
			log.Printf("Error fetching routing data for %s: %v", isp, err)
			continue
		}
		
		results[isp] = map[string]interface{}{
			"average_latency": latency,
			"path": path,
		}
	}
	
	return results, nil
}

func (d *Database) AnalyzeDNS(domain string, nodes string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	
	query := `SELECT node, resolution_time, ip_addresses, success 
	          FROM dns_results 
	          WHERE domain LIKE $1 
	          ORDER BY created_at DESC 
	          LIMIT 50`
	
	rows, err := d.db.Query(query, "%"+domain+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var node string
		var resolutionTime float64
		var ipAddresses string
		var success bool
		
		if err := rows.Scan(&node, &resolutionTime, &ipAddresses, &success); err != nil {
			return nil, err
		}
		
		results[node] = map[string]interface{}{
			"resolution_time": resolutionTime,
			"ip_addresses": ipAddresses,
			"success": success,
		}
	}
	
	return results, nil
}

func (d *Database) SaveCheckResult(result CheckResult) error {
	query := `INSERT INTO check_results (check_id, node, url, status, ping, dns, http, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := d.db.Exec(query, result.CheckID, result.Node, result.URL, result.Status, result.Ping, result.DNS, result.HTTP, result.CreatedAt)
	return err
}

func (d *Database) SaveTopology(topology Topology) error {
	asPathJSON, _ := json.Marshal(topology.ASPath)
	geoPathJSON, _ := json.Marshal(topology.GeoPath)
	hopsJSON, _ := json.Marshal(topology.NetworkHops)
	
	query := `INSERT INTO topologies (source, target, isp, as_path, geographic_path, network_hops, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := d.db.Exec(query, topology.Source, topology.Target, topology.ISP, asPathJSON, geoPathJSON, hopsJSON, topology.CreatedAt)
	return err
}

func (d *Database) UpdateNodeStatus(nodeID int, status string) error {
	query := `UPDATE nodes SET status = $1 WHERE id = $2`
	_, err := d.db.Exec(query, status, nodeID)
	return err
}
