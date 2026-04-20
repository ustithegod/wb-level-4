package server

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime/debug"

	"github.com/ustithegod/wb-level-4/gc-analyzer/internal/metrics"
)

type Server struct {
	cfg       Config
	collector *metrics.Collector
	mux       *http.ServeMux
}

func New(cfg Config) *Server {
	debug.SetGCPercent(cfg.GCPercent)

	s := &Server{
		cfg:       cfg,
		collector: metrics.NewCollector(),
		mux:       http.NewServeMux(),
	}

	s.routes()

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/metrics", s.handleMetrics)

	s.mux.HandleFunc("/debug/pprof/", pprof.Index)
	s.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	s.mux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	s.mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	s.mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	s.mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	s.mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	s.mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(
		w,
		"gc-analyzer is running\n\nmetrics: http://%s/metrics\npprof:   http://%s/debug/pprof/\nhealth:  http://%s/healthz\n",
		r.Host,
		r.Host,
		r.Host,
	)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	snapshot := s.collector.Snapshot(s.cfg.GCPercent)
	if err := metrics.WritePrometheus(w, snapshot); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
