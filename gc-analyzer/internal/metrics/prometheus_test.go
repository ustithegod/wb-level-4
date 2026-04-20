package metrics

import (
	"bytes"
	"strings"
	"testing"
)

func TestWritePrometheus(t *testing.T) {
	t.Parallel()

	input := []Metric{
		{
			Name:  "example_metric_total",
			Help:  "Example counter.",
			Type:  "counter",
			Value: 42,
		},
		{
			Name:  "example_gauge",
			Help:  "Example gauge.",
			Type:  "gauge",
			Value: 3.14,
		},
	}

	var buf bytes.Buffer
	if err := WritePrometheus(&buf, input); err != nil {
		t.Fatalf("WritePrometheus() error = %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"# HELP example_metric_total Example counter.",
		"# TYPE example_metric_total counter",
		"example_metric_total 42",
		"# HELP example_gauge Example gauge.",
		"# TYPE example_gauge gauge",
		"example_gauge 3.14",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output does not contain %q\nfull output:\n%s", want, output)
		}
	}
}
