package metrics

import (
	"fmt"
	"io"
	"strconv"
)

func WritePrometheus(w io.Writer, metrics []Metric) error {
	for _, metric := range metrics {
		if _, err := fmt.Fprintf(w, "# HELP %s %s\n", metric.Name, metric.Help); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "# TYPE %s %s\n", metric.Name, metric.Type); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s %s\n", metric.Name, strconv.FormatFloat(metric.Value, 'f', -1, 64)); err != nil {
			return err
		}
	}

	return nil
}
