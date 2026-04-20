package coordinator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ustithegod/wb-level-4/quorum-cut/internal/cut"
	"github.com/ustithegod/wb-level-4/quorum-cut/internal/node"
)

func TestSearchAll(t *testing.T) {
	t.Parallel()

	srv1 := httptest.NewServer(testNodeHandler("n1"))
	defer srv1.Close()
	srv2 := httptest.NewServer(testNodeHandler("n2"))
	defer srv2.Close()

	client := New([]string{
		strings.TrimPrefix(srv1.URL, "http://"),
		strings.TrimPrefix(srv2.URL, "http://"),
	}, "all", 1)

	inputs := []cut.InputLine{
		{Seq: 0, Text: "a:b:c"},
		{Seq: 1, Text: "x:y:z"},
	}
	opts := cut.Options{FieldList: "2", Delimiter: ":"}

	results, err := client.Search(context.Background(), inputs, opts)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Text != "b" || results[1].Text != "y" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestSearchQuorum(t *testing.T) {
	t.Parallel()

	srv1 := httptest.NewServer(testNodeHandler("n1"))
	defer srv1.Close()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv2.Close()
	srv3 := httptest.NewServer(testNodeHandler("n3"))
	defer srv3.Close()

	client := New([]string{
		strings.TrimPrefix(srv1.URL, "http://"),
		strings.TrimPrefix(srv2.URL, "http://"),
		strings.TrimPrefix(srv3.URL, "http://"),
	}, "quorum", 1)

	inputs := []cut.InputLine{
		{Seq: 0, Text: "a:b:c"},
		{Seq: 1, Text: "x:y:z"},
		{Seq: 2, Text: "1:2:3"},
	}
	opts := cut.Options{FieldList: "2", Delimiter: ":"}

	results, err := client.Search(context.Background(), inputs, opts)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func testNodeHandler(nodeID string) http.Handler {
	server := node.NewServer(nodeID)
	return server.Handler()
}
