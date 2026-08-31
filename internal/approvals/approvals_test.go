package approvals

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up to the directory holding .agentic/, so tests run from anywhere.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".agentic")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no .agentic/ found above the test directory")
		}
		dir = parent
	}
}

// TestGoReproducesRecordedDigests is the check that makes the second implementation of
// the hash schemes safe to have.
//
// scripts/check_gates.py implements legacy/truncated-32 and legacy/file-32 in Python; this
// package implements them again in Go. Neither can be deleted -- CI needs the Python one
// and devctl needs one that does not shell out -- so the duplication is pinned to a shared
// corpus instead. Every digest in .agentic/approvals/ was computed by the Python side and
// written down by a human. If Go reproduces them, the two implementations agree.
//
// The assertion is per PATH, not per record: for each artifact path, SOME record must
// carry the digest Go computes for it today. Requiring every record to match would be
// wrong, because records are historical. platform_config has been closed four times as the
// registries changed, so APR-0008, APR-0009 and APR-0010 hold digests of gates.yaml that
// are deliberately no longer current; only APR-0011 describes what is on disk. A test that
// failed on those would be asserting that approval history cannot exist.
//
// This reads the working tree where check_gates.py reads HEAD. That difference is real and
// intended -- check_gates evaluates a diff between commits, this evaluates what is on disk
// now -- and the two coincide on a clean tree, which is what CI always has.
func TestGoReproducesRecordedDigests(t *testing.T) {
	root := repoRoot(t)
	records, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("no approval records found; this test would pass vacuously")
	}

	// path -> the digest Go computes now, plus every digest any record recorded for it.
	computed := map[string]string{}
	recorded := map[string][]string{}
	for _, rec := range records {
		for _, art := range rec.Artifacts {
			got, ok := Digest(root, art)
			if !ok {
				// Unverifiable is a legitimate outcome, not a failure: legacy/truncated-32
				// is undefined over a YAML file. Skip it rather than count it as agreement.
				continue
			}
			computed[art.Path] = got
			recorded[art.Path] = append(recorded[art.Path],
				rec.ID+"="+recordedDigest(art.ContentHash))
		}
	}
	if len(computed) == 0 {
		t.Fatal("every artifact was unverifiable; the schemes were never exercised")
	}

	for path, got := range computed {
		matched := false
		for _, entry := range recorded[path] {
			if strings.HasSuffix(entry, "="+got) {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("%s: Go computes %s, which no record carries.\n"+
				"  recorded: %s\n"+
				"  Either the Go and Python hash implementations have diverged, or every "+
				"record naming this path is stale and the gate is genuinely open.",
				path, got, strings.Join(recorded[path], " "))
		}
	}
	t.Logf("reproduced the recorded digest for %d path(s) across %d record(s)",
		len(computed), len(records))
}

// TestBothSchemesAreExercised guards the test above from decaying into a one-scheme
// check. If the corpus ever loses all of one kind, the agreement test would keep passing
// while covering half of what it claims to.
func TestBothSchemesAreExercised(t *testing.T) {
	root := repoRoot(t)
	records, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	var config, document int
	for _, rec := range records {
		for _, art := range rec.Artifacts {
			if _, ok := Digest(root, art); !ok {
				continue
			}
			if art.Kind == "config" {
				config++
			} else {
				document++
			}
		}
	}
	if config == 0 {
		t.Error("no verifiable config artifact: legacy/file-32 is untested")
	}
	if document == 0 {
		t.Error("no verifiable document artifact: legacy/truncated-32 is untested")
	}
}

// --- Cover, against a fixture rather than the live corpus ----------------------------
//
// The live records are the right corpus for the agreement test, because agreement is a
// claim about them. Cover's behaviour is not: pinning it to real records would make the
// test fail whenever someone approves something, which trains people to edit tests.

func fixture(t *testing.T) (root string, doc Artifact) {
	t.Helper()
	root = t.TempDir()
	path := filepath.Join(root, "docs", "thing.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("---\nid: X\n---\nbody text\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	art := Artifact{ID: "X", Path: "docs/thing.md"}
	sum, ok := Digest(root, art)
	if !ok {
		t.Fatal("fixture document produced no digest")
	}
	art.ContentHash = "legacy/truncated-32:sha256:" + sum
	return root, art
}

func TestCoverReportsCoveredWhenHashMatches(t *testing.T) {
	root, art := fixture(t)
	recs := []Record{{ID: "APR-0001", Gate: "g", Artifacts: []Artifact{art}}}
	got := Cover(root, "g", recs)
	if got.Coverage != Covered {
		t.Fatalf("got %q (%s), want covered", got.Coverage, got.Detail)
	}
	if got.RecordID != "APR-0001" {
		t.Errorf("record %q, want APR-0001", got.RecordID)
	}
}

func TestCoverReportsStaleWhenBodyChanges(t *testing.T) {
	root, art := fixture(t)
	recs := []Record{{ID: "APR-0001", Gate: "g", Artifacts: []Artifact{art}}}
	if err := os.WriteFile(filepath.Join(root, art.Path),
		[]byte("---\nid: X\n---\ndifferent body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Cover(root, "g", recs)
	if got.Coverage != Stale {
		t.Fatalf("got %q, want stale", got.Coverage)
	}
	if !strings.Contains(got.Detail, "earlier version") {
		t.Errorf("detail %q does not say the approval is of an earlier version", got.Detail)
	}
}

// Front matter is excluded from the hash on purpose: approving a document sets its own
// approval fields, so a hash covering them could never match once recorded.
func TestCoverIgnoresFrontMatterEdits(t *testing.T) {
	root, art := fixture(t)
	recs := []Record{{ID: "APR-0001", Gate: "g", Artifacts: []Artifact{art}}}
	if err := os.WriteFile(filepath.Join(root, art.Path),
		[]byte("---\nid: X\nstatus: approved\n---\nbody text\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Cover(root, "g", recs); got.Coverage != Covered {
		t.Fatalf("got %q (%s), want covered: front matter is not part of the hash",
			got.Coverage, got.Detail)
	}
}

func TestCoverReportsOpenWhenNoRecordNamesTheGate(t *testing.T) {
	root, art := fixture(t)
	recs := []Record{{ID: "APR-0001", Gate: "other", Artifacts: []Artifact{art}}}
	if got := Cover(root, "g", recs); got.Coverage != Open {
		t.Fatalf("got %q, want open", got.Coverage)
	}
}

// A gate re-approved after each material change accumulates records; platform_config
// already carries four. A current record must win over an earlier stale one regardless
// of the order they are read in.
func TestCoverPrefersACurrentRecordOverAnEarlierStaleOne(t *testing.T) {
	root, art := fixture(t)
	old := art
	old.ContentHash = "legacy/truncated-32:sha256:" + strings.Repeat("0", 32)
	recs := []Record{
		{ID: "APR-0001", Gate: "g", Artifacts: []Artifact{old}},
		{ID: "APR-0002", Gate: "g", Artifacts: []Artifact{art}},
	}
	got := Cover(root, "g", recs)
	if got.Coverage != Covered || got.RecordID != "APR-0002" {
		t.Fatalf("got %q from %s, want covered from APR-0002", got.Coverage, got.RecordID)
	}
}

// A record approves its artifacts as a set. One changed file invalidates the approval of
// the pair, even though the other is untouched.
func TestCoverIsStaleWhenOnlyOneArtifactOfASetChanges(t *testing.T) {
	root, a := fixture(t)
	bPath := filepath.Join(root, "docs", "other.md")
	if err := os.WriteFile(bPath, []byte("---\nid: Y\n---\nsecond body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b := Artifact{ID: "Y", Path: "docs/other.md"}
	sum, _ := Digest(root, b)
	b.ContentHash = "legacy/truncated-32:sha256:" + sum

	recs := []Record{{ID: "APR-0001", Gate: "g", Artifacts: []Artifact{a, b}}}
	if got := Cover(root, "g", recs); got.Coverage != Covered {
		t.Fatalf("baseline: got %q, want covered", got.Coverage)
	}
	if err := os.WriteFile(bPath, []byte("---\nid: Y\n---\nedited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Cover(root, "g", recs); got.Coverage != Stale {
		t.Fatalf("got %q, want stale: one changed artifact invalidates the set", got.Coverage)
	}
}

// A config artifact hashes whole; a document hashes its body. Picking the scheme from the
// file extension instead of the record's kind would get this backwards.
func TestConfigKindHashesTheWholeFileIncludingFrontMatter(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "cfg.md") // .md, but kind: config
	if err := os.WriteFile(path, []byte("---\nid: X\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	art := Artifact{ID: "X", Kind: "config", Path: "cfg.md"}
	before, ok := Digest(root, art)
	if !ok {
		t.Fatal("config artifact produced no digest")
	}
	if err := os.WriteFile(path, []byte("---\nid: X\nextra: 1\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	after, _ := Digest(root, art)
	if before == after {
		t.Fatal("front-matter edit did not change the digest; config hashed as a document")
	}
}
