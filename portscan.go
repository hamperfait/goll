// +build windows linux
package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// MAXWORKERS Need worker pool because running 1 goroutine per port exhausts file descriptors
const MAXWORKERS = 100

//PortRange is the range of ports
type PortRange struct {
	Start uint64
	End   uint64
}

func (pr *PortRange) String() string {
	return fmt.Sprintf("[%v,%v)", pr.Start, pr.End)
}

// Run the port scanner
func scan(host string, PortRange string, debug bool) string {
	if host == "" || PortRange == "" {
		return "Usage: scan <host>  port|start-end,[port|start-end ...] [-debug]"
	}
	prs, err := parseRanges(PortRange)
	if err != nil {
		log.Fatal(err)
	}

	// Format results
	report := ""
	for _, pr := range prs {
		results := ScanPorts(host, pr)
		for port, success := range results {
			if success || debug {
				report += fmt.Sprintf("%v: %v\n", port, success)
			}
		}
	}
	return report
}

// Parse port ranges spec
func parseRanges(RangesStr string) ([]*PortRange, error) {
	parts := strings.Split(RangesStr, ",")
	ranges := make([]*PortRange, 0)
	for _, part := range parts {
		rg, err := parseRange(part)
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, rg)
	}
	return ranges, nil
}

//TODO: check overflow
func parseRange(RangeStr string) (*PortRange, error) {
	parts := strings.SplitN(RangeStr, "-", 2)
	nums := make([]uint64, len(parts))
	for i, v := range parts {
		n, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return nil, err
		}
		nums[i] = n
	}
	switch len(nums) {
	case 1:
		return &PortRange{
			Start: nums[0],
			End:   nums[0] + 1,
		}, nil
	case 2:
		return &PortRange{
			Start: nums[0],
			End:   nums[1],
		}, nil
	default:
		return nil, fmt.Errorf("Invalid Port Specification")
	}
}

// ScanResult is the container for scan results from workers
type ScanResult struct {
	Port    uint64
	Success bool
	Err     error
}

// ScanPorts runs the scan with a worker pool; memory usage grows in proportion
// with number of ports scanned to prevent deadlock from blocking channels
func ScanPorts(host string, pr *PortRange) map[uint64]bool {
	numPorts := pr.End - pr.Start + 1
	results := make(map[uint64]bool)
	jobpipe := make(chan uint64, numPorts)
	respipe := make(chan *ScanResult, numPorts)

	// Start workers
	for worker := 0; worker < MAXWORKERS; worker++ {
		go scanWorker(host, jobpipe, respipe)
	}

	// Seed w/ jobs
	for port := pr.Start; port < pr.End+1; port++ {
		jobpipe <- port
	}

	// Receive results
	received := uint64(0)
	for received < pr.End-pr.Start {
		res := <-respipe
		results[res.Port] = res.Success
		received++
	}
	return results
}

// Worker function; pull from job queue forever and return results on result
// queue
func scanWorker(host string, jobpipe chan uint64, respipe chan *ScanResult) {
	for job := <-jobpipe; ; job = <-jobpipe {
		respipe <- scanPort(host, job)
	}
}

// Simple scan of a single port
//	- Just tries to connect to <host>:<port> over TCP and checks for error
func scanPort(host string, port uint64) *ScanResult {
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.Dial("tcp", fmt.Sprintf("%v:%v", host, port))
	result := ScanResult{
		Port:    port,
		Success: err == nil,
		Err:     err,
	}
	if conn != nil {
		conn.Close()
	}
	return &result
}
