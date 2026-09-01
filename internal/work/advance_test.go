package work

import (
	"errors"
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
	got, err := MergedItems(dir, "main", nil, nil)
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
	got, err := MergedItems(dir, "main", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("found %v, want nothing: the subject mentions an item but closes none", got)
	}
}

func TestMergedItemsFindsSeveralItemsInOneCommit(t *testing.T) {
	dir := gitRepo(t, "spec(a): batch\n\nCloses WI-0004\nCloses WI-0005\n")
	got, err := MergedItems(dir, "main", nil, nil)
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
	merged := map[string]Merge{"WI-0001": {"abc123", "trailer"}}
	got := Plan(t.TempDir(), items, merged, nil)
	if len(got) != 1 || got[0].To != "done" {
		t.Fatalf("got %+v, want one change to done", got)
	}
}

func TestPlanLeavesAnUnmergedItemAlone(t *testing.T) {
	items := []config.WorkItem{wi("WI-0001", "review")}
	if got := Plan(t.TempDir(), items, map[string]Merge{}, nil); len(got) != 0 {
		t.Fatalf("got %+v, want no change: nothing merged it", got)
	}
}

// A gate with no record cannot advance an item to done. This is the half of the rule
// that carries the weight -- the other half would let a merge close a human gate, which
// ADR-019 says a merge never does.
func TestPlanSendsAMergedItemWithAnOpenGateToAwaitingHuman(t *testing.T) {
	items := []config.WorkItem{wi("WI-0001", "review", "platform_config")}
	merged := map[string]Merge{"WI-0001": {"abc123", "trailer"}}
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
	merged := map[string]Merge{"WI-0001": {"abc123", "trailer"}}

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
	merged := map[string]Merge{"WI-0001": {"a", "trailer"}, "WI-0002": {"b", "trailer"}}
	if got := Plan(t.TempDir(), items, merged, nil); len(got) != 0 {
		t.Fatalf("got %+v, want nothing: done and cancelled are terminal", got)
	}
}

func TestPlanIsOrderedByID(t *testing.T) {
	items := []config.WorkItem{wi("WI-0009", "review"), wi("WI-0002", "review"), wi("WI-0005", "review")}
	merged := map[string]Merge{"WI-0009": {"a", "trailer"}, "WI-0002": {"b", "trailer"}, "WI-0005": {"c", "trailer"}}
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

// --- the pull-request fallback ------------------------------------------------------
//
// The trailer has failed eight times on this repository's own trunk, six of them
// consecutively, with every branch commit carrying it and the squash setting on
// COMMIT_MESSAGES. These tests cover the path that works anyway.

type fakeResolver struct {
	branches map[int]string
	err      error
}

func (f fakeResolver) MergedBranches() (map[int]string, error) { return f.branches, f.err }

func TestMergedItemsFallsBackToThePullRequestNumber(t *testing.T) {
	dir := gitRepo(t, "spec(x): a change with no trailer at all (#42)\n\nbody without the word\n")
	items := []config.WorkItem{{ID: "WI-0099", Branch: "spec/99-something"}}
	res := fakeResolver{branches: map[int]string{42: "spec/99-something"}}

	got, err := MergedItems(dir, "main", items, res)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := got["WI-0099"]
	if !ok {
		t.Fatalf("not recovered; got %v", got)
	}
	if m.Source != "pull-request" {
		t.Errorf("source %q, want pull-request: the caller has to be able to tell", m.Source)
	}
}

// The trailer is preferred when present, so an offline machine and a networked one agree
// on every commit that followed the convention.
func TestMergedItemsPrefersTheTrailerOverTheFallback(t *testing.T) {
	dir := gitRepo(t, "spec(x): a change (#42)\n\nCloses WI-0099\n")
	items := []config.WorkItem{{ID: "WI-0099", Branch: "spec/99-something"}}
	res := fakeResolver{branches: map[int]string{42: "spec/99-something"}}

	got, _ := MergedItems(dir, "main", items, res)
	if got["WI-0099"].Source != "trailer" {
		t.Fatalf("source %q, want trailer", got["WI-0099"].Source)
	}
}

// The case above cannot fail if the fallback overrides the trailer, because a commit with a
// trailer never reaches the fallback at all. This one can: two commits, and only the older
// carries the trailer. A trailer-derived entry must never be replaced by a resolved one --
// otherwise the reliable source loses to the one that exists because the reliable source
// failed.
func TestMergedItemsDoesNotLetTheFallbackOverwriteATrailer(t *testing.T) {
	dir := gitRepo(t,
		"spec(a): the merge that closed it (#41)\n\nCloses WI-0099\n",
		"spec(b): a later touch on the same branch (#42)\n\nno trailer here\n")
	items := []config.WorkItem{{ID: "WI-0099", Branch: "spec/99-something"}}
	res := fakeResolver{branches: map[int]string{41: "spec/99-something", 42: "spec/99-something"}}

	got, _ := MergedItems(dir, "main", items, res)
	if got["WI-0099"].Source != "trailer" {
		t.Fatalf("source %q, want trailer: the fallback overwrote a trailer-derived entry",
			got["WI-0099"].Source)
	}
}

// A branch that matches no work item must not be guessed at. This is the guard against the
// fallback inventing a link the way the trailer never could.
func TestMergedItemsIgnoresAnUnmatchedBranch(t *testing.T) {
	dir := gitRepo(t, "spec(x): something else (#42)\n\nno trailer\n")
	items := []config.WorkItem{{ID: "WI-0099", Branch: "spec/99-something"}}
	res := fakeResolver{branches: map[int]string{42: "spec/77-unrelated"}}

	if got, _ := MergedItems(dir, "main", items, res); len(got) != 0 {
		t.Fatalf("got %v, want nothing: the branch matches no work item", got)
	}
}

// A resolver failure degrades to the trailer-only answer. An offline machine should still
// see obvious drift rather than getting an error instead of a report.
func TestMergedItemsDegradesWhenTheResolverFails(t *testing.T) {
	dir := gitRepo(t,
		"spec(a): trailer present (#41)\n\nCloses WI-0098\n",
		"spec(b): trailer missing (#42)\n\nnothing\n")
	items := []config.WorkItem{
		{ID: "WI-0098", Branch: "spec/98-a"},
		{ID: "WI-0099", Branch: "spec/99-b"},
	}
	res := fakeResolver{err: errors.New("no network")}

	got, err := MergedItems(dir, "main", items, res)
	if err != nil {
		t.Fatalf("resolver failure became a run failure: %v", err)
	}
	if _, ok := got["WI-0098"]; !ok {
		t.Error("the trailer-derived item was lost")
	}
	if _, ok := got["WI-0099"]; ok {
		t.Error("an item was recovered with no working resolver")
	}
}

// The number is only trusted at the end of the subject, where GitHub writes it. An issue
// reference in prose is not a merge record.
func TestMergedItemsIgnoresAPullRequestNumberMidSubject(t *testing.T) {
	dir := gitRepo(t, "spec(x): revert the change from (#42) that broke things\n\nno trailer\n")
	items := []config.WorkItem{{ID: "WI-0099", Branch: "spec/99-something"}}
	res := fakeResolver{branches: map[int]string{42: "spec/99-something"}}

	if got, _ := MergedItems(dir, "main", items, res); len(got) != 0 {
		t.Fatalf("got %v: a number mid-subject is not the squash suffix", got)
	}
}

func TestFormatPlanNamesTheFallback(t *testing.T) {
	out := FormatPlan([]Change{{ID: "WI-0099", From: "review", To: "done",
		Commit: "abc123", Source: "pull-request", Why: []string{"no required gates"}}}, false)
	if !strings.Contains(out, "no Closes trailer") {
		t.Errorf("output does not say the trailer was missing:\n%s", out)
	}
}
