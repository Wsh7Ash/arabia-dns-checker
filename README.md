# Arabia DNS Checker

A distributed uptime monitoring system using cheap VPS nodes across Jeddah, Amman, and Manama to check website accessibility and network topology mapping for Gulf region infrastructure.

## 🌟 Features

- **Distributed Monitoring**: VPS nodes in Jeddah, Amman, and Manama
- **Network Topology Mapping**: Visualize routing differences across Gulf ISPs
- **DNS Resolution Tracking**: Monitor DNS propagation and consistency
- **Real-time Alerts**: Instant notifications for outages and routing changes
- **Historical Analytics**: Track performance trends and ISP behavior
- **Firewall Detection**: Identify ISP-level blocking and filtering
- **Raspberry Pi Support**: Community nodes for expanded coverage
- **Open-source Agents**: Go-based monitoring agents for easy deployment

## 🚀 Quick Start

### Installation

```bash
git clone https://github.com/Wsh7Ash/arabia-dns-checker
cd arabia-dns-checker
go mod download
```

### Configuration

```bash
# Copy configuration template
cp config/config.example.yaml config/config.yaml

# Edit with your settings
nano config/config.yaml
```

### Running the System

```bash
# Start the central server
go run cmd/server/main.go

# Deploy monitoring agents
./deploy-agent.sh jeddah
./deploy-agent.sh amman
./deploy-agent.sh manama

# Start the web dashboard
go run cmd/web/main.go
```

## 📍 Monitoring Nodes

### Primary VPS Nodes

| Location | ISP | Coordinates | Node Type |
|----------|-----|-------------|-----------|
| Jeddah | STC | 21.4225°N, 39.8262°E | Primary |
| Amman | Zain Jordan | 31.9539°N, 35.9106°E | Primary |
| Manama | Batelco | 26.0667°N, 50.5577°E | Primary |

### Community Nodes

| Location | ISP | Node Type | Status |
|----------|-----|-----------|--------|
| Dubai | Etisalat | Community | Active |
| Riyadh | Mobily | Community | Active |
| Kuwait | Zain Kuwait | Community | Active |
| Doha | Ooredoo | Community | Active |

## 📡 API Endpoints

### Monitoring

#### Check Website from Multiple Locations
```http
POST /api/v1/check/website
Content-Type: application/json

{
  "url": "https://example.com",
  "nodes": ["jeddah", "amman", "manama"],
  "checks": ["ping", "dns", "http", "traceroute"]
}
```

#### Get Real-time Status
```http
GET /api/v1/status/nodes
```

#### Get Historical Data
```http
GET /api/v1/history/website?domain=example.com&period=24h
```

### Network Analysis

#### Get Network Topology
```http
GET /api/v1/network/topology?source=jeddah&target=example.com
```

#### Compare ISP Routing
```http
GET /api/v1/network/compare?domain=example.com&isps=stc,zain,batelco
```

#### Detect DNS Issues
```http
GET /api/v1/dns/analyze?domain=example.com&nodes=all
```

## 📊 Response Examples

### Multi-location Check Response

```json
{
  "check_id": "chk_1234567890",
  "url": "https://example.com",
  "timestamp": "2026-04-25T10:30:00Z",
  "results": {
    "jeddah": {
      "status": "accessible",
      "response_time": 245,
      "ping": {
        "success": true,
        "time_ms": 89,
        "packet_loss": 0
      },
      "dns": {
        "success": true,
        "resolution_time": 12,
        "ip_addresses": ["93.184.216.34"]
      },
      "http": {
        "success": true,
        "status_code": 200,
        "response_time": 245,
        "ssl_valid": true
      },
      "traceroute": {
        "hops": 12,
        "path": [
          {"hop": 1, "ip": "10.0.0.1", "time": 1.2},
          {"hop": 2, "ip": "213.42.0.1", "time": 15.3}
        ]
      }
    },
    "amman": {
      "status": "accessible",
      "response_time": 312,
      "ping": {"success": true, "time_ms": 125, "packet_loss": 0},
      "dns": {"success": true, "resolution_time": 18, "ip_addresses": ["93.184.216.34"]},
      "http": {"success": true, "status_code": 200, "response_time": 312, "ssl_valid": true}
    },
    "manama": {
      "status": "blocked",
      "response_time": null,
      "ping": {"success": false, "error": "timeout"},
      "dns": {"success": true, "resolution_time": 14, "ip_addresses": ["93.184.216.34"]},
      "http": {"success": false, "error": "connection_refused"}
    }
  },
  "summary": {
    "total_nodes": 3,
    "accessible_nodes": 2,
    "blocked_nodes": 1,
    "average_response_time": 278.5,
    "issues_detected": ["blocking_in_manama"]
  }
}
```

### Network Topology Response

```json
{
  "source": "jeddah",
  "target": "example.com",
  "topology": {
    "isp": "STC",
    "as_path": [3549, 6453, 8075, 15169],
    "geographic_path": [
      {"city": "Jeddah", "country": "SA", "lat": 21.4225, "lng": 39.8262},
      {"city": "Riyadh", "country": "SA", "lat": 24.7136, "lng": 46.6753},
      {"city": "Frankfurt", "country": "DE", "lat": 50.1109, "lng": 8.6821}
    ],
    "network_hops": [
      {"hop": 1, "ip": "10.0.0.1", "description": "Local gateway"},
      {"hop": 2, "ip": "213.42.0.1", "description": "STC edge router"},
      {"hop": 3, "ip": "84.23.87.1", "description": "STC core router"},
      {"hop": 8, "ip": "80.81.192.1", "description": "International gateway"}
    ]
  },
  "performance": {
    "total_latency": 89,
    "isp_latency": 45,
    "international_latency": 44
  }
}
```

## 🏗️ Architecture

```
arabia-dns-checker/
├── cmd/
│   ├── server/               # Central server
│   │   └── main.go
│   ├── agent/                # Monitoring agent
│   │   └── main.go
│   └── web/                  # Web dashboard
│       └── main.go
├── internal/
│   ├── server/               # Server logic
│   ├── agent/                # Agent logic
│   ├── monitoring/           # Monitoring operations
│   ├── network/              # Network analysis
│   └── storage/              # Database operations
├── pkg/
│   ├── dns/                  # DNS operations
│   ├── traceroute/           # Traceroute functionality
│   ├── ping/                 # Ping operations
│   └── geo/                  # Geographic utilities
├── config/
│   ├── config.yaml           # Main configuration
│   └── nodes.yaml            # Node configurations
├── deployments/
│   ├── docker/               # Docker configurations
│   ├── systemd/              # Systemd service files
│   └── kubernetes/           # K8s manifests
├── web/
│   ├── static/               # Static assets
│   └── templates/            # HTML templates
├── scripts/
│   ├── deploy-agent.sh       # Agent deployment script
│   ├── setup-node.sh         # Node setup script
│   └── backup.sh             # Backup script
├── tests/
├── go.mod
├── go.sum
└── README.md
```

## 🤖 Monitoring Agent

### Agent Features

```go
// Main agent monitoring loop
func (a *Agent) RunMonitoring() {
    ticker := time.NewTicker(a.config.CheckInterval)
    
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
    for _, target := range a.config.Targets {
        result := a.checkTarget(target)
        a.sendResult(result)
    }
}
```

### Supported Checks

- **Ping**: ICMP ping with packet loss tracking
- **DNS**: DNS resolution with multiple servers
- **HTTP**: HTTP/HTTPS connectivity checks
- **Traceroute**: Network path tracing
- **SSL**: Certificate validation
- **TCP Port**: Custom port connectivity

### Agent Deployment

```bash
# Deploy agent to remote server
./deploy-agent.sh --host jeddah.example.com --user root

# Deploy to Raspberry Pi
./deploy-agent.sh --host pi.local --user pi --device raspberry-pi

# Deploy Docker container
docker run -d --name dns-agent \
  -v /config:/app/config \
  arabia-dns-checker/agent:latest
```

## 🌍 Geographic Coverage

### ISP Coverage Map

| Country | ISPs Covered | Node Count | Coverage Type |
|---------|--------------|-------------|---------------|
| Saudi Arabia | STC, Mobily, Zain | 3 | Primary |
| Jordan | Zain Jordan, Orange | 2 | Primary |
| Bahrain | Batelco, Zain Bahrain | 2 | Primary |
| UAE | Etisalat, du | 2 | Community |
| Kuwait | Zain Kuwait, Ooredoo | 2 | Community |
| Qatar | Ooredoo Qatar | 1 | Community |

### Network Insights

#### Common Routing Patterns

```json
{
  "saudi_to_europe": {
    "typical_path": "STC -> Riyadh -> Frankfurt -> Target",
    "latency": "45-60ms",
    "common_isps": ["STC", "Mobily"]
  },
  "jordan_to_gcc": {
    "typical_path": "Zain -> Amman -> Dubai -> Target",
    "latency": "25-35ms",
    "common_isps": ["Zain Jordan", "Orange"]
  }
}
```

#### Firewall Detection

```go
func DetectFirewall(pingResult PingResult, tracerouteResult TracerouteResult) FirewallInfo {
    if pingResult.PacketLoss > 50 && tracerouteResult.StopsAtISP {
        return FirewallInfo{
            Detected: true,
            Type: "isp_firewall",
            Location: tracerouteResult.LastHop,
            Severity: "high"
        }
    }
    return FirewallInfo{Detected: false}
}
```

## 📱 Web Dashboard

### Real-time Monitoring

- **Node Status**: Live status of all monitoring nodes
- **Target Health**: Real-time health of monitored websites
- **Network Maps**: Visual representation of network paths
- **Alert Center**: Active alerts and notifications

### Historical Analytics

- **Performance Trends**: Response time trends over time
- **ISP Comparison**: Performance comparison between ISPs
- **Outage History**: Historical outage data and analysis
- **Geographic Analysis**: Regional performance patterns

### Alert System

```javascript
// Real-time alert updates
const ws = new WebSocket('wss://arabia-dns-checker.com/ws/alerts');

ws.onmessage = function(event) {
    const alert = JSON.parse(event.data);
    
    if (alert.severity === 'critical') {
        showCriticalAlert(alert);
        sendSMSNotification(alert);
    }
    
    updateDashboard(alert);
};
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests
go test -tags=integration ./tests/integration/...

# Test specific functionality
go test ./pkg/dns/...
```

## 📈 Performance Metrics

- **Check Frequency**: Every 30 seconds
- **Node Response Time**: < 100ms
- **Database Queries**: < 50ms average
- **Alert Latency**: < 5 seconds
- **Dashboard Updates**: Real-time via WebSocket

## 🔧 Configuration

### Main Configuration

```yaml
# config/config.yaml
server:
  port: 8080
  database_url: "postgresql://user:pass@localhost/arabia_dns"
  redis_url: "redis://localhost:6379"

monitoring:
  check_interval: 30s
  timeout: 10s
  retry_count: 3

nodes:
  jeddah:
    host: "jeddah.arabia-dns-checker.com"
    port: 9001
    location: "Jeddah, Saudi Arabia"
    isp: "STC"
    
  amman:
    host: "amman.arabia-dns-checker.com"
    port: 9001
    location: "Amman, Jordan"
    isp: "Zain Jordan"

alerts:
  email_enabled: true
  sms_enabled: true
  webhook_url: "https://hooks.slack.com/..."
  thresholds:
    response_time: 1000  # ms
    packet_loss: 10      # percent
```

### Node Configuration

```yaml
# config/nodes.yaml
targets:
  - domain: "google.com"
    checks: ["ping", "dns", "http"]
    interval: 30s
    
  - domain: "local-website.sa"
    checks: ["ping", "dns", "http", "ssl"]
    interval: 60s
    
  - domain: "government-portal.jo"
    checks: ["ping", "dns", "http"]
    interval: 120s
```

## 🚀 Deployment

### Docker Deployment

```bash
# Build and deploy with Docker Compose
docker-compose up -d

# Scale agents
docker-compose up -d --scale agent=5
```

### Kubernetes Deployment

```yaml
# deployments/kubernetes/agent-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dns-agent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: dns-agent
  template:
    metadata:
      labels:
        app: dns-agent
    spec:
      containers:
      - name: agent
        image: arabia-dns-checker/agent:latest
        env:
        - name: NODE_LOCATION
          value: "jeddah"
```

### Raspberry Pi Deployment

```bash
# Setup Raspberry Pi as monitoring node
curl -sSL https://get.arabia-dns-checker.com/pi | bash

# Configure as community node
sudo systemctl enable dns-agent
sudo systemctl start dns-agent
```

## 🔒 Security Features

- **Agent Authentication**: Mutual TLS between server and agents
- **Data Encryption**: Encrypted communication channels
- **Access Control**: Role-based access to monitoring data
- **Audit Logging**: Complete audit trail of all operations
- **Rate Limiting**: Prevent abuse of monitoring resources

## 📄 License

MIT License - see LICENSE file for details

## 🤝 Contributing

We welcome contributions! See CONTRIBUTING.md for guidelines.

### Adding New Nodes

1. Add node configuration in `config/nodes.yaml`
2. Deploy agent using `deploy-agent.sh`
3. Update node status in database
4. Add monitoring targets for the region

### Adding New Check Types

1. Implement check interface in `pkg/checks/`
2. Add check configuration options
3. Update agent to support new check type
4. Add tests for new functionality

## 🙏 Acknowledgments

- Gulf ISPs for network transparency
- Open-source monitoring tools and libraries
- Community node operators for expanded coverage
- Network engineers for routing insights

## 📞 Support

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Email**: dns-checker@example.com
- **Status Page**: https://status.arabia-dns-checker.com

---

**Note**: This monitoring system relies on community participation for comprehensive coverage. Consider running a community node to help improve Gulf region network visibility.
