package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSumHandlerOptimized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/sum", strings.NewReader(`{"numbers":[10,20,-5]}`))
	rec := httptest.NewRecorder()

	sumHandlerOptimized(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if got, want := strings.TrimSpace(string(body)), `{"sum":25,"count":3}`; got != want {
		t.Fatalf("body = %s, want %s", got, want)
	}
}

func TestParseNumbers(t *testing.T) {
	sum, count, err := parseNumbers([]byte(`{"numbers":[1, 2, -3, 4]}`))
	if err != nil {
		t.Fatalf("parseNumbers returned error: %v", err)
	}
	if sum != 4 {
		t.Fatalf("sum = %d, want 4", sum)
	}
	if count != 4 {
		t.Fatalf("count = %d, want 4", count)
	}
}
