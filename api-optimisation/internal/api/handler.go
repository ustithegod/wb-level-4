package api

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
)

const maxBodySize = 1 << 20

var responseBufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 128)
		return &buf
	},
}

func sumHandlerOptimized(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "empty body")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	sum, count, err := parseNumbers(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	bufPtr := responseBufferPool.Get().(*[]byte)
	buf := (*bufPtr)[:0]
	buf = append(buf, `{"sum":`...)
	buf = strconv.AppendInt(buf, sum, 10)
	buf = append(buf, `,"count":`...)
	buf = strconv.AppendInt(buf, int64(count), 10)
	buf = append(buf, '}')

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(buf)

	*bufPtr = buf[:0]
	responseBufferPool.Put(bufPtr)
}

func parseNumbers(body []byte) (int64, int, error) {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return 0, 0, errors.New("empty body")
	}

	const prefix = `{"numbers":[`
	if len(body) < len(prefix)+2 || !bytes.HasPrefix(body, []byte(prefix)) || body[len(body)-2] != ']' || body[len(body)-1] != '}' {
		return 0, 0, errors.New(`expected payload like {"numbers":[1,2,3]}`)
	}

	payload := body[len(prefix) : len(body)-2]
	payload = bytes.TrimSpace(payload)
	if len(payload) == 0 {
		return 0, 0, nil
	}

	var (
		sum   int64
		count int
		start = 0
	)

	for start < len(payload) {
		for start < len(payload) && payload[start] == ' ' {
			start++
		}
		end := start
		for end < len(payload) && payload[end] != ',' {
			end++
		}

		token := bytes.TrimSpace(payload[start:end])
		if len(token) == 0 {
			return 0, 0, errors.New("empty number token")
		}

		value, err := strconv.ParseInt(string(token), 10, 64)
		if err != nil {
			return 0, 0, errors.New("invalid integer in numbers")
		}
		sum += value
		count++

		start = end + 1
	}

	return sum, count, nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}
