package cut

import (
	"errors"
	"fmt"
	"strconv"
)

type Mode string

const (
	ModeBytes  Mode = "bytes"
	ModeChars  Mode = "chars"
	ModeFields Mode = "fields"
)

type Options struct {
	ByteList        string `json:"byte_list"`
	CharList        string `json:"char_list"`
	FieldList       string `json:"field_list"`
	Delimiter       string `json:"delimiter"`
	SuppressNoDelim bool   `json:"suppress_no_delim"`
	OutputDelimiter string `json:"output_delimiter"`
}

func (o Options) Validate() error {
	selected := 0
	if o.ByteList != "" {
		selected++
	}
	if o.CharList != "" {
		selected++
	}
	if o.FieldList != "" {
		selected++
	}
	if selected != 1 {
		return errors.New("exactly one of -b, -c, or -f must be specified")
	}
	if o.Delimiter == "" {
		o.Delimiter = "\t"
	}
	return nil
}

func (o Options) Mode() Mode {
	switch {
	case o.ByteList != "":
		return ModeBytes
	case o.CharList != "":
		return ModeChars
	default:
		return ModeFields
	}
}

func (o Options) List() string {
	switch o.Mode() {
	case ModeBytes:
		return o.ByteList
	case ModeChars:
		return o.CharList
	default:
		return o.FieldList
	}
}

func ParsePositiveInt(raw string, flagName string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", flagName, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive", flagName)
	}
	return value, nil
}
