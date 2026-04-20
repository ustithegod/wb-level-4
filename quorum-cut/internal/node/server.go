package node

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ustithegod/wb-level-4/quorum-cut/internal/cut"
	"github.com/ustithegod/wb-level-4/quorum-cut/internal/protocol"
)

type Server struct {
	nodeID string
}

func NewServer(nodeID string) *Server {
	if nodeID == "" {
		nodeID = "node"
	}
	return &Server{nodeID: nodeID}
}

func (s *Server) ListenAndServe(listenAddr string) error {
	server := &http.Server{
		Addr:    listenAddr,
		Handler: s.Handler(),
	}
	return server.ListenAndServe()
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/cut", s.handleCut)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleCut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("decode request: %v", err), http.StatusBadRequest)
		return
	}

	results, err := cut.ProcessInputs(req.Lines, req.Options)
	resp := protocol.SearchResponse{
		NodeID:  s.nodeID,
		ChunkID: req.ChunkID,
		Results: results,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
