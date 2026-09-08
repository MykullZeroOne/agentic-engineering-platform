package loop

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/runtime"
	"gopkg.in/yaml.v3"
)

// fullFakeRun runs the loop to its step-7 park (a successful pass with no re-entry,
// no question) against an isolated fixture, and returns the root, run id and record.
func fullFakeRun(t *testing.T) (config.Root, string, *Record) {
	t.Helper()
	root := fixture(t, nil)
	control := &fakeControl{}
	checker := &fakeChecker{Script: [][]Check{{
		{Name: "go-build", Command: "go build ./...", RC: 0},
	}}}
	adapter := &fakeAdapter{Sessions: []fakeSession{{
		SessionID: "sess-1",
		Texts:     []string{"implemented the change"},
		Result:    runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0},
	}}}

	out, err := Run(context.Background(), Options{
		Root:    root,
		Item:    "WI-0001",
		Adapter: adapter,
		Control: control,
		Checker: checker,
		Now:     fixedNow(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.Exit != 0 {
		t.Fatalf("Exit = %d, want 0", out.Exit)
	}
	return root, out.RunID, out.Record
}

func TestC19_RecordCarriesRequiredKeysAndNoContextManifest(t *testing.T) {
	root, runID, _ := fullFakeRun(t)

	b, err := os.ReadFile(RecordPath(root, runID))
	if err != nil {
		t.Fatal(err)
	}
	var tree map[string]any
	if err := yaml.Unmarshal(b, &tree); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{
		"schema", "run_id", "agent_identity", "role_version", "runtime", "work",
		"state", "passes", "steps", "evidence", "handoff", "stubbed",
	} {
		if _, ok := tree[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	rt, ok := tree["runtime"].(map[string]any)
	if !ok {
		t.Fatal("runtime is not a map")
	}
	if rt["model"] != "provider-current" {
		t.Errorf("runtime.model = %v, want provider-current", rt["model"])
	}
	if _, ok := rt["provider"]; !ok {
		t.Error("runtime.provider missing")
	}
	if _, ok := rt["override"]; !ok {
		t.Error("runtime.override missing")
	}

	work, ok := tree["work"].(map[string]any)
	if !ok {
		t.Fatal("work is not a map")
	}
	if work["item"] != "WI-0001" {
		t.Errorf("work.item = %v, want WI-0001", work["item"])
	}

	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			if _, bad := x["context_manifest"]; bad {
				t.Fatal("found a context_manifest key somewhere in the record")
			}
			for _, vv := range x {
				walk(vv)
			}
		case []any:
			for _, vv := range x {
				walk(vv)
			}
		}
	}
	walk(tree)
}

func TestC19_StateIsAnAgentStateToken(t *testing.T) {
	root, _, rec := fullFakeRun(t)

	states, err := config.LoadStates(root, ".agentic/registries/states.yaml")
	if err != nil {
		t.Fatal(err)
	}
	tokens, ok := states.Tokens("agent_state")
	if !ok {
		t.Fatal("fixture states.yaml has no agent_state axis")
	}
	if !tokens[rec.State] {
		t.Errorf("record state %q is not an agent_state token", rec.State)
	}
}

func TestC5_EvidenceReadsFromDiskWithNoSessionAlive(t *testing.T) {
	root, runID, _ := fullFakeRun(t)

	// A fresh LoadRecord call: no session, no in-memory state from the run above.
	rec, err := LoadRecord(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Evidence) == 0 {
		t.Fatal("expected at least one evidence entry")
	}
	for _, e := range rec.Evidence {
		if e.RunID != rec.RunID {
			t.Errorf("evidence %s has RunID %q, want %q", e.ID, e.RunID, rec.RunID)
		}
		p := filepath.Join(RunDir(root, runID), e.Path)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("evidence %s path %q does not resolve to a file on disk: %v", e.ID, e.Path, err)
		}
	}
}
