package protocol

import "github.com/ustithegod/wb-level-4/quorum-cut/internal/cut"

type SearchRequest struct {
	RequestID string          `json:"request_id"`
	ChunkID   int             `json:"chunk_id"`
	Options   cut.Options     `json:"options"`
	Lines     []cut.InputLine `json:"lines"`
}

type SearchResponse struct {
	NodeID  string           `json:"node_id"`
	ChunkID int              `json:"chunk_id"`
	Results []cut.OutputLine `json:"results"`
	Error   string           `json:"error,omitempty"`
}
