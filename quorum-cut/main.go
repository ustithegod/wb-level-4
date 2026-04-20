package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ustithegod/wb-level-4/quorum-cut/internal/coordinator"
	"github.com/ustithegod/wb-level-4/quorum-cut/internal/cut"
	"github.com/ustithegod/wb-level-4/quorum-cut/internal/node"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return runLocal(nil, stdin, stdout)
	}

	switch args[0] {
	case "node":
		return runNode(args[1:], stderr)
	case "search":
		return runSearch(args[1:], stdin, stdout)
	case "local":
		return runLocal(args[1:], stdin, stdout)
	case "-h", "--help", "help":
		printUsage(stdout)
		return nil
	default:
		return runLocal(args, stdin, stdout)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  mycut local [cut flags] [file ...]")
	fmt.Fprintln(w, "  mycut search --nodes host1,host2 [cut flags] [file ...]")
	fmt.Fprintln(w, "  mycut node --listen :9001 [--id node-1]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Core cut flags:")
	fmt.Fprintln(w, "  -b list                  select bytes")
	fmt.Fprintln(w, "  -c list                  select characters")
	fmt.Fprintln(w, "  -f list                  select fields")
	fmt.Fprintln(w, "  -d delim                 field delimiter, default tab")
	fmt.Fprintln(w, "  -s                       suppress lines without delimiter in field mode")
	fmt.Fprintln(w, "  --output-delimiter str   delimiter for selected fields")
}

func runLocal(args []string, stdin io.Reader, stdout io.Writer) error {
	opts, files, err := parseCutFlags(args)
	if err != nil {
		return err
	}

	inputs, err := cut.LoadInputs(files, stdin)
	if err != nil {
		return err
	}

	results, err := cut.ProcessInputs(inputs, opts)
	if err != nil {
		return err
	}

	return cut.WriteResults(stdout, results)
}

func runNode(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("node", flag.ContinueOnError)
	fs.SetOutput(stderr)

	listenAddr := fs.String("listen", ":9001", "address to listen on")
	nodeID := fs.String("id", "", "node id")
	if err := fs.Parse(args); err != nil {
		return err
	}

	server := node.NewServer(*nodeID)
	return server.ListenAndServe(*listenAddr)
}

func runSearch(args []string, stdin io.Reader, stdout io.Writer) error {
	searchArgs, nodes, consistency, timeout, chunkSize, err := splitSearchFlags(args)
	if err != nil {
		return err
	}

	opts, files, err := parseCutFlags(searchArgs)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("search mode requires --nodes")
	}

	inputs, err := cut.LoadInputs(files, stdin)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := coordinator.New(nodes, consistency, chunkSize)
	results, err := client.Search(ctx, inputs, opts)
	if err != nil {
		return err
	}

	return cut.WriteResults(stdout, results)
}

func splitSearchFlags(args []string) ([]string, []string, string, time.Duration, int, error) {
	var cutArgs []string
	var nodes []string
	consistency := "quorum"
	timeout := 5 * time.Second
	chunkSize := 0

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--nodes":
			if i+1 >= len(args) {
				return nil, nil, "", 0, 0, errors.New("--nodes requires a value")
			}
			i++
			nodes = splitCSV(args[i])
		case strings.HasPrefix(arg, "--nodes="):
			nodes = splitCSV(strings.TrimPrefix(arg, "--nodes="))
		case arg == "--consistency":
			if i+1 >= len(args) {
				return nil, nil, "", 0, 0, errors.New("--consistency requires a value")
			}
			i++
			consistency = args[i]
		case strings.HasPrefix(arg, "--consistency="):
			consistency = strings.TrimPrefix(arg, "--consistency=")
		case arg == "--timeout":
			if i+1 >= len(args) {
				return nil, nil, "", 0, 0, errors.New("--timeout requires a value")
			}
			i++
			parsed, err := time.ParseDuration(args[i])
			if err != nil {
				return nil, nil, "", 0, 0, fmt.Errorf("parse --timeout: %w", err)
			}
			timeout = parsed
		case strings.HasPrefix(arg, "--timeout="):
			parsed, err := time.ParseDuration(strings.TrimPrefix(arg, "--timeout="))
			if err != nil {
				return nil, nil, "", 0, 0, fmt.Errorf("parse --timeout: %w", err)
			}
			timeout = parsed
		case arg == "--chunk-size":
			if i+1 >= len(args) {
				return nil, nil, "", 0, 0, errors.New("--chunk-size requires a value")
			}
			i++
			parsed, err := cut.ParsePositiveInt(args[i], "--chunk-size")
			if err != nil {
				return nil, nil, "", 0, 0, err
			}
			chunkSize = parsed
		case strings.HasPrefix(arg, "--chunk-size="):
			parsed, err := cut.ParsePositiveInt(strings.TrimPrefix(arg, "--chunk-size="), "--chunk-size")
			if err != nil {
				return nil, nil, "", 0, 0, err
			}
			chunkSize = parsed
		default:
			cutArgs = append(cutArgs, arg)
		}
	}

	switch consistency {
	case "quorum", "all":
	default:
		return nil, nil, "", 0, 0, fmt.Errorf("unsupported consistency %q", consistency)
	}

	return cutArgs, nodes, consistency, timeout, chunkSize, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseCutFlags(args []string) (cut.Options, []string, error) {
	fs := flag.NewFlagSet("cut", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	byteList := fs.String("b", "", "select bytes")
	charList := fs.String("c", "", "select characters")
	fieldList := fs.String("f", "", "select fields")
	delimiter := fs.String("d", "\t", "field delimiter")
	suppress := fs.Bool("s", false, "suppress lines without delimiters")
	outputDelimiter := fs.String("output-delimiter", "", "output delimiter for field mode")

	if err := fs.Parse(args); err != nil {
		return cut.Options{}, nil, err
	}

	opts := cut.Options{
		ByteList:        *byteList,
		CharList:        *charList,
		FieldList:       *fieldList,
		Delimiter:       *delimiter,
		SuppressNoDelim: *suppress,
		OutputDelimiter: *outputDelimiter,
	}

	if err := opts.Validate(); err != nil {
		return cut.Options{}, nil, err
	}

	return opts, fs.Args(), nil
}
