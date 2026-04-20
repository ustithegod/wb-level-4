package cut

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type span struct {
	start int
	end   int
}

type selector struct {
	spans []span
}

func parseSelector(spec string) (selector, error) {
	parts := strings.Split(spec, ",")
	spans := make([]span, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return selector{}, fmt.Errorf("invalid empty range in %q", spec)
		}

		rangeParts := strings.Split(part, "-")
		switch len(rangeParts) {
		case 1:
			value, err := parseIndex(rangeParts[0])
			if err != nil {
				return selector{}, err
			}
			spans = append(spans, span{start: value, end: value})
		case 2:
			var start, end int
			if rangeParts[0] == "" {
				start = 1
			} else {
				value, err := parseIndex(rangeParts[0])
				if err != nil {
					return selector{}, err
				}
				start = value
			}
			if rangeParts[1] == "" {
				end = -1
			} else {
				value, err := parseIndex(rangeParts[1])
				if err != nil {
					return selector{}, err
				}
				end = value
			}
			if end != -1 && start > end {
				return selector{}, fmt.Errorf("invalid descending range %q", part)
			}
			spans = append(spans, span{start: start, end: end})
		default:
			return selector{}, fmt.Errorf("invalid range %q", part)
		}
	}

	sort.Slice(spans, func(i, j int) bool {
		if spans[i].start == spans[j].start {
			return spans[i].end < spans[j].end
		}
		return spans[i].start < spans[j].start
	})

	return selector{spans: spans}, nil
}

func parseIndex(raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse index %q: %w", raw, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("index must be positive: %q", raw)
	}
	return value, nil
}

func (s selector) contains(index int) bool {
	for _, item := range s.spans {
		if index < item.start {
			return false
		}
		if item.end == -1 {
			if index >= item.start {
				return true
			}
			continue
		}
		if index >= item.start && index <= item.end {
			return true
		}
	}
	return false
}
