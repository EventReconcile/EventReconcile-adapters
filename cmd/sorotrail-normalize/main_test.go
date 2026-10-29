package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const ndjson = `{"id":"event-1","contract_id":"CABC","ledger":100,"type":"contract","tx_hash":"tx-a","tx_index":0,"op_index":0,"event_index":0,"topics":[{"symbol":"transfer"}],"value":{"i128":"10"}}` + "\n"

func TestRunNormalizesPairAndPreservesSafeWrite(t *testing.T) {
	d := t.TempDir()
	ref := filepath.Join(d, "ref.ndjson")
	obs := filepath.Join(d, "obs.ndjson")
	out := filepath.Join(d, "input.json")
	for _, p := range []string{ref, obs} {
		if err := os.WriteFile(p, []byte(ndjson), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr strings.Builder
	if code := run([]string{"--reference", ref, "--observed", obs, "--output", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"--reference", ref, "--observed", obs, "--output", out}, &stdout, &stderr); code == 0 {
		t.Fatal("expected safe-write collision")
	}
}

func TestRunRejectsDifferentContracts(t *testing.T) {
	d := t.TempDir()
	ref, obs, out := filepath.Join(d, "ref.ndjson"), filepath.Join(d, "obs.ndjson"), filepath.Join(d, "input.json")
	if err := os.WriteFile(ref, []byte(ndjson), 0600); err != nil {
		t.Fatal(err)
	}
	other := strings.Replace(ndjson, "CABC", "CDEF", 1)
	if err := os.WriteFile(obs, []byte(other), 0600); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"--reference", ref, "--observed", obs, "--output", out}, &strings.Builder{}, &strings.Builder{}); code == 0 {
		t.Fatal("expected contract mismatch")
	}
}
