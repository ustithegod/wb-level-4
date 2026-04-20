package cut

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type InputLine struct {
	Seq     int    `json:"seq"`
	File    string `json:"file"`
	LineNum int    `json:"line_num"`
	Text    string `json:"text"`
}

type OutputLine struct {
	Seq     int    `json:"seq"`
	File    string `json:"file"`
	LineNum int    `json:"line_num"`
	Text    string `json:"text"`
}

func LoadInputs(files []string, stdin io.Reader) ([]InputLine, error) {
	if len(files) == 0 {
		return loadReader(stdin, "stdin", 0)
	}

	var all []InputLine
	seq := 0
	for _, file := range files {
		fh, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", file, err)
		}

		lines, err := loadReader(fh, file, seq)
		closeErr := fh.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close %s: %w", file, closeErr)
		}

		all = append(all, lines...)
		seq += len(lines)
	}

	return all, nil
}

func loadReader(r io.Reader, file string, seqStart int) ([]InputLine, error) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lines := make([]InputLine, 0, 128)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		lines = append(lines, InputLine{
			Seq:     seqStart + len(lines),
			File:    file,
			LineNum: lineNum,
			Text:    scanner.Text(),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	return lines, nil
}

func ProcessInputs(inputs []InputLine, opts Options) ([]OutputLine, error) {
	selector, err := parseSelector(opts.List())
	if err != nil {
		return nil, err
	}

	out := make([]OutputLine, 0, len(inputs))
	for _, input := range inputs {
		text, keep := processLine(input.Text, opts, selector)
		if !keep {
			continue
		}
		out = append(out, OutputLine{
			Seq:     input.Seq,
			File:    input.File,
			LineNum: input.LineNum,
			Text:    text,
		})
	}
	return out, nil
}

func processLine(line string, opts Options, selector selector) (string, bool) {
	switch opts.Mode() {
	case ModeBytes:
		return cutBytes(line, selector), true
	case ModeChars:
		return cutChars(line, selector), true
	default:
		return cutFields(line, opts, selector)
	}
}

func cutBytes(line string, selector selector) string {
	data := []byte(line)
	var b strings.Builder
	for i, value := range data {
		if selector.contains(i + 1) {
			b.WriteByte(value)
		}
	}
	return b.String()
}

func cutChars(line string, selector selector) string {
	runes := []rune(line)
	var b strings.Builder
	for i, value := range runes {
		if selector.contains(i + 1) {
			b.WriteRune(value)
		}
	}
	return b.String()
}

func cutFields(line string, opts Options, selector selector) (string, bool) {
	delim := opts.Delimiter
	if delim == "" {
		delim = "\t"
	}
	if !strings.Contains(line, delim) {
		if opts.SuppressNoDelim {
			return "", false
		}
		return line, true
	}

	fields := strings.Split(line, delim)
	selected := make([]string, 0, len(fields))
	for i, field := range fields {
		if selector.contains(i + 1) {
			selected = append(selected, field)
		}
	}

	outputDelim := opts.OutputDelimiter
	if outputDelim == "" {
		outputDelim = delim
	}

	return strings.Join(selected, outputDelim), true
}

func WriteResults(w io.Writer, results []OutputLine) error {
	for _, result := range results {
		if _, err := fmt.Fprintln(w, result.Text); err != nil {
			return err
		}
	}
	return nil
}
