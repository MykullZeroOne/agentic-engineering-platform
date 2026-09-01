package approvals

import (
	"os"
	"os/exec"
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

// pythonDigest computes the same two schemes in Python, using the exact expressions
// scripts/check_gates.py uses, and returns "" when python3 is unavailable.
func pythonDigest(t *testing.T, root string, art Artifact) string {
	t.Helper()
	src := `
import hashlib, sys
path, kind = sys.argv[1], sys.argv[2]
raw = open(path, "rb").read()
if kind == "config":
    print(hashlib.sha256(raw).hexdigest()[:32]); raise SystemExit
s = raw.decode()
if not path.endswith(".md") or not s.startswith("---\n"):
    print(""); raise SystemExit
end = s.find("\n---\n", 4)
print("" if end == -1 else hashlib.sha256(s[end+5:].encode()).hexdigest()[:32])
`
	out, err := exec.Command("python3", "-c", src, filepath.Join(root, art.Path), art.Kind).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// TestGoAgreesWithPython is the check that makes the second implementation of the hash
// schemes safe to have.
//
// scripts/check_gates.py implements legacy/truncated-32 and legacy/file-32 in Python; this
// package implements them again in Go. Neither can be deleted -- CI needs the Python one and
// devctl needs one that does not shell out -- so the two are compared directly over the same
// bytes.
//
// An earlier version compared Go's digest against the digests humans had RECORDED, requiring
// that some record carry the value Go computes. That was wrong in a way that only showed up
// under use: it fails on any legitimately gated edit. Editing an approved document is how a
// gate is opened, and the resulting hash mismatch is the gate working -- not the two
// implementations disagreeing. The test conflated "Go and Python differ", which is a bug,
// with "this document has a pending unapproved edit", which is Tuesday. PRs #24 and #25 both
// went red for it.
//
// Comparing the implementations to each other tests the actual claim and is indifferent to
// whether any document is currently approved.
func TestGoAgreesWithPython(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available; cross-language agreement not checkable here")
	}
	root := repoRoot(t)
	records, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	seen, checked := map[string]bool{}, 0
	for _, rec := range records {
		for _, art := range rec.Artifacts {
			key := art.Kind + ":" + art.Path
			if seen[key] {
				continue
			}
			seen[key] = true

			got, ok := Digest(root, art)
			want := pythonDigest(t, root, art)
			if !ok {
				// Both must agree that the scheme defines no digest here, or one of them
				// is silently hashing something the other refuses to.
				if want != "" {
					t.Errorf("%s (%s): Go computes no digest, Python computes %s", art.Path, art.Kind, want)
				}
				continue
			}
			if want == "" {
				t.Errorf("%s (%s): Go computes %s, Python computes no digest", art.Path, art.Kind, got)
				continue
			}
			if got != want {
				t.Errorf("%s (%s): Go %s, Python %s -- the implementations have diverged",
					art.Path, art.Kind, got, want)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("no artifact was hashed by both; the comparison never ran")
	}
	t.Logf("Go and Python agree on %d artifact digest(s)", checked)
}

// TestSomeRecordedDigestIsReproducible guards the above from passing while Go and Python
// agree on a scheme that is not the one the records were written under. It needs only one
// match, because a path with a pending gated edit legitimately matches nothing.
func TestSomeRecordedDigestIsReproducible(t *testing.T) {
	root := repoRoot(t)
	records, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, rec := range records {
		for _, art := range rec.Artifacts {
			if got, ok := Digest(root, art); ok && got == recordedDigest(art.ContentHash) {
				t.Logf("reproduced %s from %s", art.Path, rec.ID)
				return
			}
		}
	}
	t.Fatal("no recorded digest is reproducible; the schemes do not match the corpus")
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

// touching is the change under evaluation: one work item, and the paths it changed. Before
// WI-0040 Cover took no change at all, which is why every record below had to be assumed
// relevant to whatever was being advanced.
func touching(paths ...string) Change {
	return Change{WorkItem: "WI-0001", Paths: paths}
}

func TestCoverReportsCoveredWhenHashMatches(t *testing.T) {
	root, art := fixture(t)
	recs := []Record{{ID: "APR-0001", Gate: "g", Artifacts: []Artifact{art}}}
	got := Cover(root, "g", recs, touching(art.Path))
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
	got := Cover(root, "g", recs, touching(art.Path))
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
	if got := Cover(root, "g", recs, touching(art.Path)); got.Coverage != Covered {
		t.Fatalf("got %q (%s), want covered: front matter is not part of the hash",
			got.Coverage, got.Detail)
	}
}

func TestCoverReportsOpenWhenNoRecordNamesTheGate(t *testing.T) {
	root, art := fixture(t)
	recs := []Record{{ID: "APR-0001", Gate: "other", Artifacts: []Artifact{art}}}
	if got := Cover(root, "g", recs, touching(art.Path)); got.Coverage != Open {
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
	got := Cover(root, "g", recs, touching(art.Path))
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
	if got := Cover(root, "g", recs, touching(a.Path)); got.Coverage != Covered {
		t.Fatalf("baseline: got %q, want covered", got.Coverage)
	}
	if err := os.WriteFile(bPath, []byte("---\nid: Y\n---\nedited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Cover(root, "g", recs, touching(a.Path)); got.Coverage != Stale {
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

// The defect WI-0040 names. APR-0011 approved two registry files on 2026-08-31 and was
// reported as closing platform_config for a 2026-09-01 change to CLAUDE.md, because the
// only thing matched was the gate id. A current record for the gate is not an approval of
// this change.
func TestCoverReportsUnrelatedWhenTheRecordIsForAnotherChange(t *testing.T) {
	root, art := fixture(t)
	recs := []Record{{ID: "APR-0011", Gate: "g", Artifacts: []Artifact{art}}}

	got := Cover(root, "g", recs, touching("CLAUDE.md"))
	if got.Coverage != Unrelated {
		t.Fatalf("got %q (%s), want unrelated: the record names neither the work item nor a changed path",
			got.Coverage, got.Detail)
	}
	if got.RecordID != "APR-0011" {
		t.Errorf("record %q, want APR-0011 named so the operator knows why it did not count", got.RecordID)
	}
}

// The second key, and the one that makes non-path gates workable: six of the twelve gates
// trigger on classification, finding, lifecycle or capability, so there is no path to match.
func TestCoverAcceptsARecordThatNamesTheWorkItem(t *testing.T) {
	root, art := fixture(t)
	recs := []Record{{ID: "APR-0011", Gate: "g", Artifacts: []Artifact{art}, WorkItems: []string{"WI-0001"}}}

	got := Cover(root, "g", recs, touching("some/other/file.md"))
	if got.Coverage != Covered {
		t.Fatalf("got %q (%s), want covered: the record names WI-0001", got.Coverage, got.Detail)
	}
	if !strings.Contains(got.Detail, "WI-0001") {
		t.Errorf("detail %q does not say why the record counted", got.Detail)
	}
}

// Naming a different work item is not naming this one.
func TestCoverIgnoresARecordNamingAnotherWorkItem(t *testing.T) {
	root, art := fixture(t)
	recs := []Record{{ID: "APR-0011", Gate: "g", Artifacts: []Artifact{art}, WorkItems: []string{"WI-9999"}}}

	if got := Cover(root, "g", recs, touching("CLAUDE.md")); got.Coverage != Unrelated {
		t.Fatalf("got %q, want unrelated", got.Coverage)
	}
}

// Stale outranks unrelated. A record that was written for this change and has gone out of
// date is a different problem from a record that was never about this change, and the first
// is the more useful thing to tell a human: re-approve, rather than approve.
func TestCoverPrefersARelatedStaleRecordOverAnUnrelatedCurrentOne(t *testing.T) {
	root, art := fixture(t)
	stale := art
	stale.ContentHash = "legacy/truncated-32:sha256:" + strings.Repeat("0", 32)

	// A current record for the same gate, over a file this change never touched.
	otherPath := filepath.Join(root, "docs", "elsewhere.md")
	if err := os.WriteFile(otherPath, []byte("---\nid: Z\n---\nelsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	other := Artifact{ID: "Z", Path: "docs/elsewhere.md"}
	sum, _ := Digest(root, other)
	other.ContentHash = "legacy/truncated-32:sha256:" + sum

	recs := []Record{
		{ID: "APR-0001", Gate: "g", Artifacts: []Artifact{stale}},
		{ID: "APR-0002", Gate: "g", Artifacts: []Artifact{other}},
	}
	got := Cover(root, "g", recs, touching(art.Path))
	if got.Coverage != Stale || got.RecordID != "APR-0001" {
		t.Fatalf("got %q from %s, want stale from APR-0001", got.Coverage, got.RecordID)
	}
}
