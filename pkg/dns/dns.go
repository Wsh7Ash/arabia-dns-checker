package dns

import (
	"context"
	"fmt"
	"net"
	"time"
)

type Result struct {
	Success        bool     `json:"success"`
	ResolutionTime float64  `json:"resolution_time"`
	IPAddresses    []string `json:"ip_addresses"`
	Error          string   `json:"error,omitempty"`
}

func Resolve(domain string, timeout time.Duration) Result {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Remove protocol prefix if present
	if len(domain) > 8 && domain[:8] == "https://" {
		domain = domain[8:]
	} else if len(domain) > 7 && domain[:7] == "http://" {
		domain = domain[7:]
	}
	
	// Remove path if present
	for i, c := range domain {
		if c == '/' {
			domain = domain[:i]
			break
		}
	}
	
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return Result{
			Success: false,
			Error:   err.Error(),
		}
	}
	
	ipStrings := make([]string, len(ips))
	for i, ip := range ips {
		ipStrings[i] = ip.IP.String()
	}
	
	return Result{
		Success:        true,
		ResolutionTime: time.Since(start).Seconds() * 1000, // Convert to milliseconds
		IPAddresses:    ipStrings,
	}
}

func ResolveWithServer(domain string, dnsServer string, timeout time.Duration) Result {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: timeout,
			}
			return d.DialContext(ctx, "udp", dnsServer)
		},
	}
	
	ips, err := resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return Result{
			Success: false,
			Error:   err.Error(),
		}
	}
	
	ipStrings := make([]string, len(ips))
	for i, ip := range ips {
		ipStrings[i] = ip.IP.String()
	}
	
	return Result{
		Success:        true,
		ResolutionTime: time.Since(start).Seconds() * 1000,
		IPAddresses:    ipStrings,
	}
}

func CheckDNSConsistency(domain string, servers []string, timeout time.Duration) map[string]Result {
	results := make(map[string]Result)
	
	for _, server := range servers {
		results[server] = ResolveWithServer(domain, server, timeout)
	}
	
	return results
}

func GetMXRecords(domain string, timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	records, err := net.DefaultResolver.LookupMX(ctx, domain)
	if err != nil {
		return nil, err
	}
	
	mxRecords := make([]string, len(records))
	for i, record := range records {
		mxRecords[i] = record.Host
	}
	
	return mxRecords, nil
}

func GetTXTRecords(domain string, timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	records, err := net.DefaultResolver.LookupTXT(ctx, domain)
	if err != nil {
		return nil, err
	}
	
	return records, nil
}

func ReverseLookup(ip string, timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	names, err := net.DefaultResolver.LookupAddr(ctx, ip)
	if err != nil {
		return nil, err
	}
	
	return names, nil
}
