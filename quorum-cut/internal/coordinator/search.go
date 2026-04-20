package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/ustithegod/wb-level-4/quorum-cut/internal/cut"
	"github.com/ustithegod/wb-level-4/quorum-cut/internal/protocol"
)

type Client struct {
	nodes       []string
	consistency string
	chunkSize   int
	httpClient  *http.Client
}

func New(nodes []string, consistency string, chunkSize int) *Client {
	return &Client{
		nodes:       append([]string(nil), nodes...),
		consistency: consistency,
		chunkSize:   chunkSize,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *Client) Search(ctx context.Context, inputs []cut.InputLine, opts cut.Options) ([]cut.OutputLine, error) {
	if len(c.nodes) == 0 {
		return nil, errors.New("no nodes configured")
	}
	if len(inputs) == 0 {
		return nil, nil
	}

	chunks := c.chunkInputs(inputs)
	required := len(chunks)
	if c.consistency == "quorum" {
		required = len(chunks)/2 + 1
	}

	resultCh := make(chan nodeResult, len(chunks))
	var wg sync.WaitGroup

	for i, chunk := range chunks {
		wg.Add(1)
		nodeAddr := c.nodes[i%len(c.nodes)]
		go func(chunkID int, lines []cut.InputLine, addr string) {
			defer wg.Done()
			resultCh <- c.dispatch(ctx, chunkID, lines, opts, addr)
		}(i, chunk, nodeAddr)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var (
		successes int
		failures  int
		results   []cut.OutputLine
	)

	for result := range resultCh {
		if result.err != nil {
			failures++
			if len(chunks)-failures < required-successes {
				return nil, fmt.Errorf("quorum unreachable: %w", result.err)
			}
			continue
		}

		successes++
		results = append(results, result.results...)
		if successes >= required {
			if c.consistency == "quorum" {
				break
			}
		}
	}

	if successes < required {
		return nil, fmt.Errorf("insufficient successful chunks: got %d want %d", successes, required)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Seq < results[j].Seq
	})
	return results, nil
}

type nodeResult struct {
	results []cut.OutputLine
	err     error
}

func (c *Client) dispatch(ctx context.Context, chunkID int, lines []cut.InputLine, opts cut.Options, addr string) nodeResult {
	payload, err := json.Marshal(protocol.SearchRequest{
		RequestID: fmt.Sprintf("chunk-%d", chunkID),
		ChunkID:   chunkID,
		Options:   opts,
		Lines:     lines,
	})
	if err != nil {
		return nodeResult{err: fmt.Errorf("marshal request: %w", err)}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+addr+"/cut", bytes.NewReader(payload))
	if err != nil {
		return nodeResult{err: fmt.Errorf("build request: %w", err)}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nodeResult{err: fmt.Errorf("request %s: %w", addr, err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return nodeResult{err: fmt.Errorf("node %s returned %s: %s", addr, resp.Status, string(body))}
	}

	var searchResp protocol.SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nodeResult{err: fmt.Errorf("decode response from %s: %w", addr, err)}
	}
	if searchResp.Error != "" {
		return nodeResult{err: fmt.Errorf("node %s: %s", addr, searchResp.Error)}
	}

	return nodeResult{results: searchResp.Results}
}

func (c *Client) chunkInputs(inputs []cut.InputLine) [][]cut.InputLine {
	size := c.chunkSize
	if size <= 0 {
		size = (len(inputs) + len(c.nodes) - 1) / len(c.nodes)
		if size == 0 {
			size = 1
		}
	}

	var chunks [][]cut.InputLine
	for start := 0; start < len(inputs); start += size {
		end := start + size
		if end > len(inputs) {
			end = len(inputs)
		}
		chunks = append(chunks, inputs[start:end])
	}
	return chunks
}
