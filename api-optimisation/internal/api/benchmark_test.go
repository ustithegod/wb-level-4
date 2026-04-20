package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var benchmarkBody = strings.NewReader(`{"numbers":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20]}`)

func BenchmarkSumHandlerNaive(b *testing.B) {
	benchmarkHandler(b, sumHandlerNaive)
}

func BenchmarkSumHandlerOptimized(b *testing.B) {
	benchmarkHandler(b, sumHandlerOptimized)
}

func benchmarkHandler(b *testing.B, handler http.HandlerFunc) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/sum", strings.NewReader(`{"numbers":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20]}`))
		rec := httptest.NewRecorder()
		handler(rec, req)

		if rec.Code != http.StatusOK {
			b.Fatalf("status = %d", rec.Code)
		}
	}
}
