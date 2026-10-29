package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/EventReconcile/EventReconcile-adapters/sorotrail"
)

type inputDocument struct {
	Reference    any    `json:"reference"`
	Observed     any    `json:"observed"`
	Coverage     any    `json:"coverage"`
	IdentityMode string `json:"identityMode"`
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("sorotrail-normalize", flag.ContinueOnError)
	flags.SetOutput(stderr)
	referencePath := flags.String("reference", "", "reference SoroTrail NDJSON export")
	observedPath := flags.String("observed", "", "observed SoroTrail NDJSON export")
	outputPath := flags.String("output", "", "core input JSON destination; existing files are never overwritten")
	if err := flags.Parse(args); err != nil || *referencePath == "" || *observedPath == "" || *outputPath == "" || flags.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: sorotrail-normalize --reference ref.ndjson --observed observed.ndjson --output input.json")
		return 2
	}
	if filepath.Clean(*referencePath) == filepath.Clean(*observedPath) || filepath.Clean(*referencePath) == filepath.Clean(*outputPath) || filepath.Clean(*observedPath) == filepath.Clean(*outputPath) {
		fmt.Fprintln(stderr, "reference, observed and output paths must differ")
		return 2
	}
	reference, err := sorotrail.LoadNDJSONFile(*referencePath)
	if err != nil {
		fmt.Fprintf(stderr, "REFERENCE_ERROR: %v\n", err)
		return 2
	}
	observed, err := sorotrail.LoadNDJSONFile(*observedPath)
	if err != nil {
		fmt.Fprintf(stderr, "OBSERVED_ERROR: %v\n", err)
		return 2
	}
	if reference.ContractID != observed.ContractID {
		fmt.Fprintln(stderr, "INPUT_ERROR: reference and observed contract IDs differ")
		return 2
	}
	mode := reference.IdentityMode
	if observed.IdentityMode == "source-scoped" {
		mode = "source-scoped"
	}
	document := inputDocument{Reference: reference.Events, Observed: observed.Events, Coverage: reference.Coverage, IdentityMode: mode}
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "OUTPUT_ERROR: %v\n", err)
		return 2
	}
	if err := writeNew(*outputPath, append(data, '\n')); err != nil {
		fmt.Fprintf(stderr, "OUTPUT_ERROR: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "NORMALIZED: %s\n", filepath.Clean(*outputPath))
	return 0
}

func writeNew(path string, data []byte) error {
	parent := filepath.Dir(filepath.Clean(path))
	if err := os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}
