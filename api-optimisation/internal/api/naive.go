package api

import (
	"encoding/json"
	"io"
	"net/http"
)

type sumRequest struct {
	Numbers []int64 `json:"numbers"`
}

type sumResponse struct {
	Sum   int64 `json:"sum"`
	Count int   `json:"count"`
}

func sumHandlerNaive(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "empty body")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	var req sumRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	var sum int64
	for _, number := range req.Numbers {
		sum += number
	}

	resp, err := json.Marshal(sumResponse{
		Sum:   sum,
		Count: len(req.Numbers),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encode failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}
