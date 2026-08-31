package work

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/approvals"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
)

// gitRepo builds a throwaway repository whose commit messages are the fixture.
//
// A real repository rather than a stubbed command runner: the thing under test is a
// claim about what `git log` returns for a squash-merged trunk, and a stub would only
// test that the parser reads the stub.
func gitRepo(t *testing.T, messages ...string) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q", "-b", "main")
	for i, msg := range messages {
		if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte(msg), 0o644); err != nil {
			t.Fatal(err)
		}
		run("add", "-A")
		run("commit", "-q", "-m", msg)
		_ = i
	}
	return dir
}

func TestMergedItemsReadsTheClosesTrailer(t *testing.T) {
	dir := gitRepo(t,
		"feat(x): first\n\nbody\n\nCloses WI-0001\n",
		"docs(y): second\n\nCloses WI-0002\n",
	)
	got, err := MergedItems(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"WI-0001", "WI-0002"} {
		if _, ok := got[id]; !ok {
			t.Errorf("%s not found; got %v", id, got)
		}
	}
	if len(got) != 2 {
		t.Errorf("found %d items, want 2: %v", len(got), got)
	}
}

// The failure this whole derivation is built around: PRs #8 and #9 merged with the
// trailer only in the pull request description, so it never reached the trunk. An item
// like that must come back absent, never guessed at from the subject line.
func TestMergedItemsIgnoresACommitWithNoTrailer(t *testing.T) {
	dir := gitRepo(t, "docs(work): mark WI-0009 done\n\nno trailer here\n")
	got, err := MergedItems(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("found %v, want nothing: the subject mentions an item but closes none", got)
	}
}

func TestMergedItemsFindsSeveralItemsInOneCommit(t *testing.T) {
	dir := gitRepo(t, "spec(a): batch\n\nCloses WI-0004\nCloses WI-0005\n")
	got, err := MergedItems(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got["WI-0004"] != got["WI-0005"] {
		t.Fatalf("got %v, want both items pointing at the same commit", got)
	}
}

// --- Plan ---------------------------------------------------------------------------

func wi(id, state string, gates ...string) config.WorkItem {
	return config.WorkItem{ID: id, WorkState: state, RequiredGates: gates}
}

func TestPlanAdvancesAMergedUngatedItemToDone(t *testing.T) {
	items := []config.WorkItem{wi("WI-0001", "review")}
	merged := map[string]string{"WI-0001": "abc123"}
	got := Plan(t.TempDir(), items, merged, nil)
	if len(got) != 1 || got[0].To != "done" {
		t.Fatalf("got %+v, want one change to done", got)
	}
}

func TestPlanLeavesAnUnmergedItemAlone(t *testing.T) {
	items := []config.WorkItem{wi("WI-0001", "review")}
	if got := Plan(t.TempDir(), items, map[string]string{}, nil); len(got) != 0 {
		t.Fatalf("got %+v, want no change: nothing merged it", got)
	}
}

// A gate with no record cannot advance an item to done. This is the half of the rule
// that carries the weight -- the other half would let a merge close a human gate, which
// ADR-019 says a merge never does.
func TestPlanSendsAMergedItemWithAnOpenGateToAwaitingHuman(t *testing.T) {
	items := []config.WorkItem{wi("WI-0001", "review", "platform_config")}
	merged := map[string]string{"WI-0001": "abc123"}
	got := Plan(t.TempDir(), items, merged, nil)
	if len(got) != 1 || got[0].To != "awaiting_human" {
		t.Fatalf("got %+v, want awaiting_human", got)
	}
	if len(got[0].Why) != 1 || !strings.Contains(got[0].Why[0], "open") {
		t.Errorf("why = %v, want it to name the open gate", got[0].Why)
	}
}

// A stale record must not advance anything. If it could, binding an approval to a
// content hash would buy nothing.
func TestPlanTreatsAStaleRecordAsNotCovered(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "doc.md")
	if err := os.WriteFile(path, []byte("---\nid: X\n---\noriginal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	art := approvals.Artifact{ID: "X", Path: "doc.md"}
	sum, _ := approvals.Digest(root, art)
	art.ContentHash = "legacy/truncated-32:sha256:" + sum
	recs := []approvals.Record{{ID: "APR-0001", Gate: "g", Artifacts: []approvals.Artifact{art}}}

	items := []config.WorkItem{wi("WI-0001", "review", "g")}
	merged := map[string]string{"WI-0001": "abc123"}

	if got := Plan(root, items, merged, recs); len(got) != 1 || got[0].To != "done" {
		t.Fatalf("baseline: got %+v, want done while the record is current", got)
	}
	if err := os.WriteFile(path, []byte("---\nid: X\n---\nedited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Plan(root, items, merged, recs)
	if len(got) != 1 || got[0].To != "awaiting_human" {
		t.Fatalf("got %+v, want awaiting_human once the artifact changed", got)
	}
}

func TestPlanSkipsTerminalStates(t *testing.T) {
	items := []config.WorkItem{wi("WI-0001", "done"), wi("WI-0002", "cancelled")}
	merged := map[string]string{"WI-0001": "a", "WI-0002": "b"}
	if got := Plan(t.TempDir(), items, merged, nil); len(got) != 0 {
		t.Fatalf("got %+v, want nothing: done and cancelled are terminal", got)
	}
}

func TestPlanIsOrderedByID(t *testing.T) {
	items := []config.WorkItem{wi("WI-0009", "review"), wi("WI-0002", "review"), wi("WI-0005", "review")}
	merged := map[string]string{"WI-0009": "a", "WI-0002": "b", "WI-0005": "c"}
	got := Plan(t.TempDir(), items, merged, nil)
	if len(got) != 3 {
		t.Fatalf("got %d changes, want 3", len(got))
	}
	for i, want := range []string{"WI-0002", "WI-0005", "WI-0009"} {
		if got[i].ID != want {
			t.Fatalf("position %d is %s, want %s: %+v", i, got[i].ID, want, got)
		}
	}
}

// --- Apply --------------------------------------------------------------------------

const itemYAML = `id: WI-0001
# a comment that a YAML round-trip would delete
work_state: review
title: Something
description: |
  A block scalar whose line breaks and indentation
  a re-marshal would reflow.

    Including this indented paragraph.
required_gates: []
`

func TestApplyChangesOnlyTheStateLine(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".agentic", "work")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "WI-0001.yaml")
	if err := os.WriteFile(path, []byte(itemYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Apply(root, []Change{{ID: "WI-0001", From: "review", To: "done"}}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(itemYAML, "work_state: review", "work_state: done", 1)
	if string(got) != want {
		t.Errorf("Apply changed more than the state line.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// `work_state` also appears inside descriptions and as a registry axis name. Only a
// top-level key may be rewritten; matching loosely would corrupt prose.
func TestApplyIgnoresTheWordInsideADescription(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".agentic", "work")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// The dangerous shape, and a real one: WI-0013's description discusses the very field
	// this rewrites. An indented mention that ends the line looks exactly like the key.
	body := "id: WI-0002\n" +
		"work_state: review\n" +
		"description: |\n" +
		"  The binding advances a merged item to:\n" +
		"    work_state: done\n" +
		"  and nothing else.\n"
	path := filepath.Join(dir, "WI-0002.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Apply(root, []Change{{ID: "WI-0002", To: "done"}}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "    work_state: done\n  and nothing else.") {
		t.Errorf("the indented mention was rewritten or the block reflowed:\n%s", got)
	}
	if !strings.Contains(string(got), "\nwork_state: done\n") {
		t.Errorf("the top-level key was not rewritten:\n%s", got)
	}
}

func TestApplyRefusesAFileWithNoStateKey(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".agentic", "work")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "WI-0003.yaml"), []byte("id: WI-0003\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Apply(root, []Change{{ID: "WI-0003", To: "done"}})
	if err == nil {
		t.Fatal("Apply accepted a file with no work_state key")
	}
	if !strings.Contains(err.Error(), "want exactly 1") {
		t.Errorf("error %q does not say what was wrong", err)
	}
}
