package sorotrail

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/EventReconcile/EventReconcile-core/reconcile"
)

const maxLines = 10000
const maxLineBytes = 4 * 1024 * 1024
const maxFileBytes int64 = 16 * 1024 * 1024

type sourceEvent struct {
	ID         string   `json:"id"`
	ContractID string   `json:"contract_id"`
	Type       string   `json:"type"`
	Ledger     uint64   `json:"ledger"`
	TxHash     string   `json:"tx_hash"`
	TxIndex    uint64   `json:"tx_index"`
	OpIndex    uint64   `json:"op_index"`
	EventIndex *uint64  `json:"event_index"`
	Topics     []any    `json:"topics"`
	Value      any      `json:"value"`
	TopicsXDR  []string `json:"topics_xdr"`
	ValueXDR   string   `json:"value_xdr"`
}

type Result struct {
	Events       []reconcile.Event
	Coverage     reconcile.Coverage
	Source       string
	IdentityMode string
	RawRecordIDs []string
	RawNDJSON    [][]byte
	ContractID   string
}

// NormalizeNDJSON validates SoroTrail's event export and converts only records
// that expose the canonical event_index. It never substitutes op_index or the
// opaque SoroTrail id for event_index.
func NormalizeNDJSON(r io.Reader) (Result, error) {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64*1024), maxLineBytes)
	out := Result{Events: []reconcile.Event{}, RawRecordIDs: []string{}, RawNDJSON: [][]byte{}, Source: "sorotrail-ndjson", IdentityMode: "canonical"}
	var from, to uint64
	for line := 1; s.Scan(); line++ {
		if line > maxLines {
			return Result{}, fmt.Errorf("sorotrail export exceeds %d records", maxLines)
		}
		var e sourceEvent
		if err := json.Unmarshal(s.Bytes(), &e); err != nil {
			return Result{}, fmt.Errorf("line %d: invalid JSON: %w", line, err)
		}
		if e.ID == "" || e.ContractID == "" || e.Type != "contract" || e.Ledger == 0 || e.TxHash == "" || e.Topics == nil {
			return Result{}, fmt.Errorf("line %d: malformed SoroTrail event", line)
		}
		if out.ContractID == "" { out.ContractID = e.ContractID } else if out.ContractID != e.ContractID { return Result{}, fmt.Errorf("line %d: mixed contract IDs are not supported", line) }
		event := reconcile.Event{LedgerSeq: e.Ledger, TxHash: e.TxHash, TxApplicationOrder: e.TxIndex, Topics: e.Topics, Data: e.Value, SourceID: e.ID}
		if e.EventIndex != nil {
			event.EventIndex = *e.EventIndex
		} else {
			out.IdentityMode = "source-scoped"
		}
		out.Events = append(out.Events, event)
		out.RawRecordIDs = append(out.RawRecordIDs, e.ID)
		out.RawNDJSON = append(out.RawNDJSON, append([]byte(nil), s.Bytes()...))
		if from == 0 || e.Ledger < from {
			from = e.Ledger
		}
		if e.Ledger > to {
			to = e.Ledger
		}
	}
	if err := s.Err(); err != nil {
		return Result{}, err
	}
	if len(out.Events) == 0 {
		return Result{}, fmt.Errorf("sorotrail export is empty")
	}
	out.Coverage = reconcile.Coverage{Ranges: []reconcile.CoverageRange{{From: from, To: to}}, Provenance: out.Source, Completeness: "unknown"}
	return out, nil
}

func LoadNDJSONFile(path string) (Result, error) {
	f, err := os.Open(path)
	if err != nil { return Result{}, err }
	defer f.Close()
	st, err := f.Stat()
	if err != nil { return Result{}, err }
	if st.Size() > maxFileBytes { return Result{}, fmt.Errorf("sorotrail export exceeds %d-byte limit", maxFileBytes) }
	data, err := io.ReadAll(io.LimitReader(f, maxFileBytes+1))
	if err != nil { return Result{}, err }
	if int64(len(data)) > maxFileBytes { return Result{}, fmt.Errorf("sorotrail export exceeds %d-byte limit", maxFileBytes) }
	return NormalizeNDJSON(bytes.NewReader(data))
}
