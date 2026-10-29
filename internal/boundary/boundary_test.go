package boundary

import (
	"testing"

	"github.com/EventReconcile/EventReconcile-core/reconcile"
)

func TestCoreContractIsReachable(t *testing.T) {
	if reconcile.SchemaVersion == "" {
		t.Fatal("core schema version is empty")
	}
}
