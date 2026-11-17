package buildinfo

import "testing"

func TestDeterministicIdentity(t *testing.T) {
	if Version != "0.1.0-dev" || Phase != "phase2" {
		t.Fatal("unexpected Phase 2 identity")
	}
}
