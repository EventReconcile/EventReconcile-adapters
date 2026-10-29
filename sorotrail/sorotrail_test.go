package sorotrail

import (
	"os"
	"strings"
	"testing"

	"github.com/EventReconcile/EventReconcile-core/reconcile"
)

const compatible = `{"id":"0000000000000001-0000000001","contract_id":"CABC","ledger":100,"type":"contract","tx_hash":"tx-a","tx_index":0,"op_index":0,"event_index":0,"topics":[{"symbol":"transfer"}],"value":{"i128":"10"},"topics_xdr":["AAA="],"value_xdr":"BBB="}`

func TestNormalizeCompatibleNDJSON(t *testing.T) {
	r, err := NormalizeNDJSON(strings.NewReader(compatible + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Events) != 1 || len(r.RawNDJSON) != 1 || r.Events[0].EventIndex != 0 || r.Coverage.Ranges[0].From != 100 || r.Source != "sorotrail-ndjson" || r.IdentityMode != "canonical" || r.ContractID != "CABC" {
		t.Fatalf("unexpected result %#v", r)
	}
}

func TestRejectsMixedContracts(t *testing.T) {
	second := strings.Replace(compatible, "\"CABC\"", "\"CDEF\"", 1)
	if _, err := NormalizeNDJSON(strings.NewReader(compatible + "\n" + second + "\n")); err == nil || !strings.Contains(err.Error(), "mixed contract") {
		t.Fatalf("got %v", err)
	}
}

func TestRejectsNonContractEvents(t *testing.T) {
	nonContract := strings.Replace(compatible, "\"contract\"", "\"diagnostic\"", 1)
	if _, err := NormalizeNDJSON(strings.NewReader(nonContract + "\n")); err == nil {
		t.Fatal("expected type validation error")
	}
}

func TestNormalizesDocumentedShapeAsSourceScoped(t *testing.T) {
	documented := `{"id":"0001099511627776-0000000001","contract_id":"CABC","ledger":256000,"type":"contract","tx_hash":"tx-a","tx_index":1,"op_index":0,"topics":[{"symbol":"transfer"}],"value":{"i128":"10"}}`
	r, err := NormalizeNDJSON(strings.NewReader(documented + "\n"))
	if err != nil || r.IdentityMode != "source-scoped" || r.Events[0].SourceID == "" {
		t.Fatalf("expected source-scoped result, got %#v %v", r, err)
	}
}

func TestSourceScopedResultFeedsCoreWithExplicitInconclusiveIdentity(t *testing.T) {
	documented := `{"id":"0001099511627776-0000000001","contract_id":"CABC","ledger":256000,"type":"contract","tx_hash":"tx-a","tx_index":1,"op_index":0,"topics":[{"symbol":"transfer"}],"value":{"i128":"10"}}`
	r, err := NormalizeNDJSON(strings.NewReader(documented + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	left, err := reconcile.Reconcile(reconcile.Input{Reference: r.Events, Observed: r.Events, Coverage: r.Coverage, IdentityMode: r.IdentityMode})
	if err != nil {
		t.Fatal(err)
	}
	if left.Overall != "INCONCLUSIVE" || left.Assertions[1].Status != "INCONCLUSIVE" {
		t.Fatalf("unexpected report %#v", left)
	}
}

func TestRejectsMalformedRecord(t *testing.T) {
	if _, err := NormalizeNDJSON(strings.NewReader("not-json\n")); err == nil {
		t.Fatal("expected malformed JSON error")
	}
}

func TestRejectsEmptyExport(t *testing.T) {
	if _, err := NormalizeNDJSON(strings.NewReader("")); err == nil {
		t.Fatal("expected empty export error")
	}
}

func TestRejectsOversizedLine(t *testing.T) {
	if _, err := NormalizeNDJSON(strings.NewReader(strings.Repeat("x", maxLineBytes) + "\n")); err == nil {
		t.Fatal("expected oversized line error")
	}
}

func TestRejectsTooManyRecords(t *testing.T) {
	data := strings.Repeat(compatible+"\n", maxLines+1)
	if _, err := NormalizeNDJSON(strings.NewReader(data)); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected record limit error, got %v", err)
	}
}

func TestLoadRejectsOversizedFile(t *testing.T) {
	d := t.TempDir()
	p := d + string(os.PathSeparator) + "oversized.ndjson"
	if err := os.WriteFile(p, make([]byte, maxFileBytes+1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadNDJSONFile(p); err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("expected file limit error, got %v", err)
	}
}
