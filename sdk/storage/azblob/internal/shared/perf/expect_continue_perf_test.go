// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

// This program benchmarks the Expect: 100-continue policy against a real Azure
// Storage account. It uploads blocks concurrently using blockblob.StageBlock
// under three modes (Off, Auto, On) and reports per-mode latency statistics to
// show the overhead (or benefit) of each mode.
//
// Usage:
//
//	export AZURE_STORAGE_CONNECTION_STRING="DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net"
//	go test -v -run TestExpectContinuePerf -timeout 10m ./sdk/storage/azblob/internal/shared/perf/
//
// Optional environment variables:
//   - PERF_CONTAINER_NAME   — container to use (default: "expect100perf")
//   - PERF_BLOCK_SIZE_BYTES — per-block body size in bytes (default: 8388608 = 8 MiB)
//   - PERF_CONCURRENCY      — number of parallel goroutines (default: 16)
//   - PERF_REQUESTS          — total number of StageBlock calls per mode (default: 200)

package perf

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/shared"
)

// readSeekCloser wraps a *bytes.Reader so it satisfies io.ReadSeekCloser.
type readSeekCloser struct {
	*bytes.Reader
}

func (r readSeekCloser) Close() error { return nil }

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}

// blockID returns a base64-encoded block ID padded to a fixed width.
func blockID(n int) string {
	s := fmt.Sprintf("%06d", n)
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// latencyStats holds collected durations for a single test run.
type latencyStats struct {
	durations []time.Duration
}

func (ls *latencyStats) add(d time.Duration) {
	ls.durations = append(ls.durations, d)
}

func (ls *latencyStats) sort() {
	sort.Slice(ls.durations, func(i, j int) bool {
		return ls.durations[i] < ls.durations[j]
	})
}

func (ls *latencyStats) p(pct float64) time.Duration {
	if len(ls.durations) == 0 {
		return 0
	}
	idx := int(math.Ceil(pct/100*float64(len(ls.durations)))) - 1
	if idx < 0 {
		idx = 0
	}
	return ls.durations[idx]
}

func (ls *latencyStats) mean() time.Duration {
	if len(ls.durations) == 0 {
		return 0
	}
	var total time.Duration
	for _, d := range ls.durations {
		total += d
	}
	return total / time.Duration(len(ls.durations))
}

func (ls *latencyStats) total() time.Duration {
	var t time.Duration
	for _, d := range ls.durations {
		t += d
	}
	return t
}

// runMode executes `totalRequests` StageBlock calls with `concurrency` parallel
// workers, using the supplied blockblob client. Returns collected per-request
// latencies.
func runMode(t *testing.T, client *blockblob.Client, body []byte, concurrency, totalRequests int) *latencyStats {
	t.Helper()

	stats := &latencyStats{durations: make([]time.Duration, 0, totalRequests)}
	var mu sync.Mutex

	type work struct {
		id int
	}
	ch := make(chan work, totalRequests)
	for i := 0; i < totalRequests; i++ {
		ch <- work{id: i}
	}
	close(ch)

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for w := 0; w < concurrency; w++ {
		go func() {
			defer wg.Done()
			for job := range ch {
				reader := readSeekCloser{bytes.NewReader(body)}
				start := time.Now()
				_, err := client.StageBlock(context.Background(), blockID(job.id), reader, nil)
				elapsed := time.Since(start)
				if err != nil {
					t.Logf("StageBlock error (id=%d): %v", job.id, err)
					continue
				}
				mu.Lock()
				stats.add(elapsed)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	stats.sort()
	return stats
}

func TestExpectContinuePerf(t *testing.T) {
	connStr := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if connStr == "" {
		t.Skip("AZURE_STORAGE_CONNECTION_STRING not set — skipping perf test")
	}

	containerName := envOrDefault("PERF_CONTAINER_NAME", "expect100perf")
	blockSize := envIntOrDefault("PERF_BLOCK_SIZE_BYTES", 8*1024*1024)
	concurrency := envIntOrDefault("PERF_CONCURRENCY", 16)
	totalRequests := envIntOrDefault("PERF_REQUESTS", 200)

	t.Logf("=== Perf configuration ===")
	t.Logf("  Container:      %s", containerName)
	t.Logf("  Block size:     %d bytes (%.1f MiB)", blockSize, float64(blockSize)/(1024*1024))
	t.Logf("  Concurrency:    %d", concurrency)
	t.Logf("  Requests/mode:  %d", totalRequests)
	t.Logf("==========================")

	// Ensure the container exists.
	cntClient, err := container.NewClientFromConnectionString(connStr, containerName, nil)
	if err != nil {
		t.Fatalf("creating container client: %v", err)
	}
	_, err = cntClient.Create(context.Background(), nil)
	if err != nil && !strings.Contains(err.Error(), "ContainerAlreadyExists") {
		t.Fatalf("creating container: %v", err)
	}

	// Generate random body once and reuse across all modes.
	body := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, body); err != nil {
		t.Fatalf("generating random body: %v", err)
	}

	// Define the three modes to benchmark.
	modes := []struct {
		name string
		opts *shared.ExpectContinueOptions
	}{
		{
			name: "Off",
			opts: &shared.ExpectContinueOptions{Mode: shared.ExpectContinueModeOff},
		},
		{
			name: "Auto (default)",
			opts: nil, // nil = default Auto mode
		},
		{
			name: "On (always)",
			opts: &shared.ExpectContinueOptions{Mode: shared.ExpectContinueModeOn},
		},
	}

	results := make(map[string]*latencyStats)

	for _, m := range modes {
		blobName := fmt.Sprintf("perfblob-%s-%d", strings.ReplaceAll(strings.ToLower(m.name), " ", "-"), time.Now().UnixNano())

		clientOpts := &blockblob.ClientOptions{}
		if m.opts != nil {
			clientOpts.ExpectContinue = m.opts
		}

		client, err := blockblob.NewClientFromConnectionString(connStr, containerName, blobName, clientOpts)
		if err != nil {
			t.Fatalf("[%s] creating blockblob client: %v", m.name, err)
		}

		t.Logf("\n--- Running mode: %s ---", m.name)
		wallStart := time.Now()
		stats := runMode(t, client, body, concurrency, totalRequests)
		wallElapsed := time.Since(wallStart)

		results[m.name] = stats

		succeeded := len(stats.durations)
		failed := totalRequests - succeeded
		throughputMBps := float64(int64(succeeded)*int64(blockSize)) / wallElapsed.Seconds() / (1024 * 1024)

		t.Logf("  Succeeded:   %d / %d (failed: %d)", succeeded, totalRequests, failed)
		t.Logf("  Wall time:   %s", wallElapsed.Round(time.Millisecond))
		t.Logf("  Throughput:  %.1f MiB/s", throughputMBps)
		t.Logf("  Mean:        %s", stats.mean().Round(time.Millisecond))
		t.Logf("  P50:         %s", stats.p(50).Round(time.Millisecond))
		t.Logf("  P90:         %s", stats.p(90).Round(time.Millisecond))
		t.Logf("  P99:         %s", stats.p(99).Round(time.Millisecond))
		t.Logf("  Min:         %s", stats.p(0).Round(time.Millisecond))
		t.Logf("  Max:         %s", stats.p(100).Round(time.Millisecond))
	}

	// Print comparison summary.
	t.Logf("\n====================================================")
	t.Logf("  COMPARISON SUMMARY")
	t.Logf("====================================================")
	t.Logf("%-20s %10s %10s %10s %10s", "Mode", "Mean", "P50", "P90", "P99")
	t.Logf("%-20s %10s %10s %10s %10s", "----", "----", "---", "---", "---")

	offStats := results["Off"]
	for _, m := range modes {
		s := results[m.name]
		overhead := ""
		if offStats != nil && s.mean() > 0 && offStats.mean() > 0 {
			pct := (float64(s.mean()) - float64(offStats.mean())) / float64(offStats.mean()) * 100
			if m.name != "Off" {
				overhead = fmt.Sprintf(" (%+.1f%%)", pct)
			}
		}
		t.Logf("%-20s %10s %10s %10s %10s%s",
			m.name,
			s.mean().Round(time.Millisecond),
			s.p(50).Round(time.Millisecond),
			s.p(90).Round(time.Millisecond),
			s.p(99).Round(time.Millisecond),
			overhead,
		)
	}
	t.Logf("====================================================")
	t.Logf("Note: 'On (always)' sends Expect: 100-continue on every request,")
	t.Logf("causing an extra round-trip. The overhead %% vs Off shows the cost.")
	t.Logf("Auto mode should show ~0%% overhead in a non-throttled scenario.")
}
