package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var (
		target      = flag.String("target", "http://localhost:8080/sum", "target URL")
		concurrency = flag.Int("c", 32, "number of workers")
		duration    = flag.Duration("d", 15*time.Second, "test duration")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        *concurrency,
			MaxIdleConnsPerHost: *concurrency,
			MaxConnsPerHost:     *concurrency,
		},
	}

	payload := []byte(`{"numbers":[1,2,3,4,5,6,7,8,9,10]}`)

	var (
		totalRequests int64
		totalErrors   int64
		latenciesMu   sync.Mutex
		latencies     = make([]time.Duration, 0, 1024)
		wg            sync.WaitGroup
	)

	for range *concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, *target, bytes.NewReader(payload))
				if err != nil {
					atomic.AddInt64(&totalErrors, 1)
					continue
				}
				req.Header.Set("Content-Type", "application/json")

				start := time.Now()
				resp, err := client.Do(req)
				if err != nil {
					atomic.AddInt64(&totalErrors, 1)
					continue
				}

				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					atomic.AddInt64(&totalErrors, 1)
					continue
				}

				latency := time.Since(start)
				atomic.AddInt64(&totalRequests, 1)

				latenciesMu.Lock()
				latencies = append(latencies, latency)
				latenciesMu.Unlock()
			}
		}()
	}

	wg.Wait()

	reqs := atomic.LoadInt64(&totalRequests)
	errs := atomic.LoadInt64(&totalErrors)

	if len(latencies) == 0 {
		fmt.Fprintln(os.Stderr, "no successful requests recorded")
		os.Exit(1)
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	totalDuration := *duration
	fmt.Printf("requests: %d\n", reqs)
	fmt.Printf("errors: %d\n", errs)
	fmt.Printf("rps: %.2f\n", float64(reqs)/totalDuration.Seconds())
	fmt.Printf("p50: %s\n", percentile(latencies, 0.50))
	fmt.Printf("p95: %s\n", percentile(latencies, 0.95))
	fmt.Printf("p99: %s\n", percentile(latencies, 0.99))
}

func percentile(values []time.Duration, p float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values)-1) * p)
	return values[index]
}
