package ping

import (
	"fmt"
	"net"
	"time"
)

type Result struct {
	Success    bool    `json:"success"`
	TimeMs     float64 `json:"time_ms"`
	PacketLoss float64 `json:"packet_loss"`
	Error      string  `json:"error,omitempty"`
}

func Ping(host string, timeout time.Duration) Result {
	start := time.Now()
	
	// Remove protocol prefix if present
	if len(host) > 8 && host[:8] == "https://" {
		host = host[8:]
	} else if len(host) > 7 && host[:7] == "http://" {
		host = host[7:]
	}
	
	// Remove path if present
	for i, c := range host {
		if c == '/' {
			host = host[:i]
			break
		}
	}
	
	// Resolve hostname to IP
	ip, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Failed to resolve %s: %v", host, err),
		}
	}
	
	// Create connection
	conn, err := net.DialTimeout("ip:icmp", ip.String(), timeout)
	if err != nil {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Failed to connect to %s: %v", host, err),
		}
	}
	defer conn.Close()
	
	// Set deadline
	conn.SetDeadline(time.Now().Add(timeout))
	
	// Send ICMP packet (simplified - in production use golang.org/x/net/icmp)
	// This is a basic TCP ping implementation as ICMP requires elevated privileges
	tcpConn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", host), timeout)
	if err != nil {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("TCP connection failed to %s: %v", host, err),
		}
	}
	defer tcpConn.Close()
	
	elapsed := time.Since(start)
	
	return Result{
		Success:    true,
		TimeMs:     elapsed.Seconds() * 1000,
		PacketLoss: 0,
	}
}

func PingMultiple(host string, count int, timeout time.Duration) []Result {
	results := make([]Result, count)
	
	for i := 0; i < count; i++ {
		results[i] = Ping(host, timeout)
		time.Sleep(100 * time.Millisecond)
	}
	
	return results
}

func CalculatePacketLoss(results []Result) float64 {
	if len(results) == 0 {
		return 0
	}
	
	failed := 0
	for _, result := range results {
		if !result.Success {
			failed++
		}
	}
	
	return (float64(failed) / float64(len(results))) * 100
}

func CalculateAverageTime(results []Result) float64 {
	if len(results) == 0 {
		return 0
	}
	
	total := 0.0
	count := 0
	
	for _, result := range results {
		if result.Success {
			total += result.TimeMs
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	
	return total / float64(count)
}

type PingStats struct {
	Host         string  `json:"host"`
	PacketsSent  int     `json:"packets_sent"`
	PacketsRecv  int     `json:"packets_received"`
	PacketLoss   float64 `json:"packet_loss_percent"`
	MinTime      float64 `json:"min_time_ms"`
	MaxTime      float64 `json:"max_time_ms"`
	AvgTime      float64 `json:"avg_time_ms"`
}

func GetPingStats(host string, count int, timeout time.Duration) PingStats {
	results := PingMultiple(host, count, timeout)
	
	stats := PingStats{
		Host:        host,
		PacketsSent: count,
	}
	
	successful := make([]Result, 0)
	for _, result := range results {
		if result.Success {
			successful = append(successful, result)
		}
	}
	
	stats.PacketsRecv = len(successful)
	stats.PacketLoss = CalculatePacketLoss(results)
	
	if len(successful) > 0 {
		min := successful[0].TimeMs
		max := successful[0].TimeMs
		
		for _, result := range successful {
			if result.TimeMs < min {
				min = result.TimeMs
			}
			if result.TimeMs > max {
				max = result.TimeMs
			}
		}
		
		stats.MinTime = min
		stats.MaxTime = max
		stats.AvgTime = CalculateAverageTime(results)
	}
	
	return stats
}
