package traceroute

import (
	"context"
	"fmt"
	"net"
	"time"
)

type Result struct {
	Success bool   `json:"success"`
	Hops    int    `json:"hops"`
	Path    []Hop  `json:"path"`
	Error   string `json:"error,omitempty"`
}

type Hop struct {
	HopNumber int      `json:"hop"`
	IP        string   `json:"ip"`
	Host      string   `json:"host,omitempty"`
	RTT       []float64 `json:"rtt_ms"`
}

func Trace(target string, timeout time.Duration) Result {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Remove protocol prefix if present
	if len(target) > 8 && target[:8] == "https://" {
		target = target[8:]
	} else if len(target) > 7 && target[:7] == "http://" {
		target = target[7:]
	}
	
	// Remove path if present
	for i, c := range target {
		if c == '/' {
			target = target[:i]
			break
		}
	}
	
	// Resolve target to IP
	ipAddr, err := net.ResolveIPAddr("ip", target)
	if err != nil {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Failed to resolve %s: %v", target, err),
		}
	}
	
	targetIP := ipAddr.IP.String()
	
	// Perform traceroute (simplified implementation)
	hops, err := performTraceroute(ctx, targetIP, 30)
	if err != nil {
		return Result{
			Success: false,
			Error:   err.Error(),
		}
	}
	
	return Result{
		Success: true,
		Hops:    len(hops),
		Path:    hops,
	}
}

func performTraceroute(ctx context.Context, targetIP string, maxHops int) ([]Hop, error) {
	hops := make([]Hop, 0)
	
	for ttl := 1; ttl <= maxHops; ttl++ {
		select {
		case <-ctx.Done():
			return hops, ctx.Err()
		default:
			hop, reached := traceHop(ctx, targetIP, ttl)
			hops = append(hops, hop)
			
			if reached {
				return hops, nil
			}
		}
	}
	
	return hops, fmt.Errorf("max hops reached without reaching target")
}

func traceHop(ctx context.Context, targetIP string, ttl int) (Hop, bool) {
	hop := Hop{
		HopNumber: ttl,
		RTT:       make([]float64, 0, 3),
	}
	
	// Send 3 probes for each hop to get average RTT
	for i := 0; i < 3; i++ {
		start := time.Now()
		
		// Create UDP connection with specific TTL
		conn, err := net.Dial("udp", fmt.Sprintf("%s:33434", targetIP))
		if err != nil {
			continue
		}
		defer conn.Close()
		
		// This is a simplified traceroute - in production use golang.org/x/net/ipv4
		// to properly set TTL and receive ICMP Time Exceeded messages
		
		rtt := time.Since(start).Seconds() * 1000
		hop.RTT = append(hop.RTT, rtt)
		
		time.Sleep(50 * time.Millisecond)
	}
	
	// In a real implementation, we would receive ICMP responses
	// For now, return a placeholder hop
	hop.IP = "*"
	hop.Host = "*"
	
	return hop, false
}

type TracerouteStats struct {
	TotalHops      int     `json:"total_hops"`
	TotalLatency   float64 `json:"total_latency_ms"`
	AvgHopLatency  float64 `json:"avg_hop_latency_ms"`
	ASPath         []int   `json:"as_path"`
	ISPPath        []string `json:"isp_path"`
}

func AnalyzeTraceroute(result Result) TracerouteStats {
	stats := TracerouteStats{
		TotalHops: result.Hops,
	}
	
	if len(result.Path) > 0 {
		totalLatency := 0.0
		for _, hop := range result.Path {
			if len(hop.RTT) > 0 {
				avgRTT := 0.0
				for _, rtt := range hop.RTT {
					avgRTT += rtt
				}
				avgRTT /= float64(len(hop.RTT))
				totalLatency += avgRTT
			}
		}
		
		stats.TotalLatency = totalLatency
		if result.Hops > 0 {
			stats.AvgHopLatency = totalLatency / float64(result.Hops)
		}
	}
	
	// In a real implementation, we would look up AS numbers and ISP names
	// for each hop IP address
	
	return stats
}

type FirewallDetection struct {
	Detected  bool   `json:"detected"`
	Type      string `json:"type,omitempty"`
	Location  string `json:"location,omitempty"`
	Severity  string `json:"severity,omitempty"`
}

func DetectFirewall(pingResult interface{}, tracerouteResult Result) FirewallDetection {
	// Analyze patterns to detect ISP-level firewalls
	// This is a simplified detection logic
	
	if tracerouteResult.Hops > 0 {
		lastHop := tracerouteResult.Path[len(tracerouteResult.Path)-1]
		
		// Check if traceroute stops at a specific hop (firewall signature)
		if lastHop.IP == "*" || lastHop.IP == "" {
			return FirewallDetection{
				Detected: true,
				Type:     "isp_firewall",
				Location: fmt.Sprintf("Hop %d", lastHop.HopNumber),
				Severity: "high",
			}
		}
	}
	
	return FirewallDetection{
		Detected: false,
	}
}
