// Package approvals reads the approval records under .agentic/approvals/ and answers
// one question: is a gate closed by a record that is still current.
//
// "Still current" is the whole point. ADR-019 binds an approval to a content hash, so a
// record whose artifact has since been edited approved something that no longer exists.
// A reader that only checked "does a record for this gate exist" would report a gate
// closed by an approval of different content, which is the overclaim the hash is there
// to prevent.
//
// The two hash schemes below are a SECOND implementation of a rule that already lives in
// scripts/check_gates.py. That duplication is deliberate and is checked rather than
// trusted: TestGoAgreesWithRecordedHashes recomputes every artifact in every record on
// disk, and validate_docs.py asserts the same thing in Python over the same corpus. Two
// implementations pinned to one corpus cannot drift without a test going red -- which is
// a different risk from two implementations left to drift in silence.
package approvals

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Artifact is one thing a record approves, pinned at the content it was approved at.
type Artifact struct {
	ID          string `yaml:"id"`
	Kind        string `yaml:"kind"`
	Path        string `yaml:"path"`
	Title       string `yaml:"title"`
	ContentHash string `yaml:"content_hash"`
}

// Record is one approval record: a human closing one gate over some artifacts.
type Record struct {
	ID         string     `yaml:"id"`
	Gate       string     `yaml:"gate"`
	Approver   string     `yaml:"approver"`
	ApprovedOn string     `yaml:"approved_on"`
	HashScheme string     `yaml:"hash_scheme"`
	Artifacts  []Artifact `yaml:"artifacts"`
	// WorkItems names the work items this record approves. Optional, and absent from every
	// record written before WI-0040: a record without it can still close a gate, but only
	// by naming a path the change touched. It exists because six of the twelve gates trigger
	// on classification, finding, lifecycle or capability rather than on a path, so for those
	// there is no path to match and naming the item is the only way to tie an approval to
	// the change it approved.
	WorkItems []string `yaml:"work_items"`
}

// Coverage is how well a gate is closed. The four states are deliberately distinct:
// collapsing Stale into Open would lose the difference between "nobody approved this"
// and "somebody approved an earlier version of it", and collapsing Unrelated into Open
// would lose the difference between "nobody approved this gate" and "somebody approved
// this gate, for a different change". Different problems, different fixes.
type Coverage string

const (
	// Covered: a record relates to the change, closes the gate, and every artifact it
	// names is unchanged.
	Covered Coverage = "covered"
	// Stale: a record relates to the change and closes the gate, but an artifact has
	// changed since.
	Stale Coverage = "stale"
	// Unrelated: a current record closes the gate, but for some other change. It names
	// neither this work item nor any path this change touched. This is the state WI-0040
	// added, and the one the old gate-id-only matching reported as Covered.
	Unrelated Coverage = "unrelated"
	// Open: no record closes the gate at all.
	Open Coverage = "open"
)

// Change is what a record must relate to before it can close a gate for it.
//
// Either key is sufficient and both are meaningful. WorkItem is the honest one -- an
// approval is given for a change, and a work item is how this repository names a change --
// but records written before WI-0040 carry no work item, so Paths keeps them working:
// a record naming a file the change touched is plainly about that change.
type Change struct {
	WorkItem string
	Paths    []string
}

// relates reports whether rec was given for this change, and why.
func (c Change) relates(rec Record) (bool, string) {
	for _, id := range rec.WorkItems {
		if id == c.WorkItem {
			return true, "names " + c.WorkItem
		}
	}
	for _, art := range rec.Artifacts {
		for _, p := range c.Paths {
			if art.Path == p {
				return true, "names " + p
			}
		}
	}
	return false, ""
}

// Result is a coverage verdict with the reason a human needs to act on it.
type Result struct {
	Coverage Coverage
	RecordID string // the record consulted; empty when Open
	Detail   string
}

// Load reads every APR-*.yaml under root's .agentic/approvals/, sorted by id.
//
// ATT-*.yaml files are deliberately skipped. An attestation is not an approval -- ATT-0002
// establishes that content is unchanged so that records can be written, and treating one
// as a closure would let a migration note close a human gate.
func Load(root string) ([]Record, error) {
	paths, err := filepath.Glob(filepath.Join(root, ".agentic", "approvals", "APR-*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	records := make([]Record, 0, len(paths))
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var rec Record
		if err := yaml.Unmarshal(raw, &rec); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(p), err)
		}
		records = append(records, rec)
	}
	return records, nil
}

// Cover reports the best coverage the records give gateID for one change.
//
// Best, not first: a gate re-approved after each material change accumulates records, and
// platform_config already carries four. Reporting the first match would let an early stale
// record mask the current one that actually closes the gate.
//
// For one change, not in general. Matching on gate id alone -- what this did before
// WI-0040 -- meant any current record for a gate closed it for every later change, so a
// 2026-08-31 approval of two registry files closed platform_config for a 2026-09-01 change
// to CLAUDE.md. The record was internally current, and about something else entirely. A
// record must now relate to the change before its currency is even asked about.
func Cover(root, gateID string, records []Record, change Change) Result {
	var stale *Result
	var unrelated *Result
	for _, rec := range records {
		if rec.Gate != gateID {
			continue
		}
		related, why := change.relates(rec)
		if !related {
			// Keep the first one so the operator is told a record exists and why it did
			// not count. "No approval record closes this gate" would be a lie here, and
			// the lie points at the wrong fix.
			if unrelated == nil {
				unrelated = &Result{Unrelated, rec.ID, fmt.Sprintf(
					"%s closes this gate for another change; it names neither %s nor any path this change touched",
					rec.ID, change.WorkItem)}
			}
			continue
		}
		if len(rec.Artifacts) == 0 {
			// A record naming no artifact pins nothing, so nothing can verify it. Say so
			// rather than counting it either way. Reachable only by naming the work item,
			// which is itself a deliberate human act.
			return Result{Covered, rec.ID, fmt.Sprintf("%s %s; names no artifact, currency unverifiable", rec.ID, why)}
		}
		fresh, reason := allCurrent(root, rec)
		if fresh {
			return Result{Covered, rec.ID, fmt.Sprintf("%s (%s)", reason, why)}
		}
		if stale == nil {
			stale = &Result{Stale, rec.ID, reason}
		}
	}
	if stale != nil {
		return *stale
	}
	if unrelated != nil {
		return *unrelated
	}
	return Result{Open, "", "no approval record closes this gate"}
}

// allCurrent reports whether every artifact in rec still hashes to its recorded digest.
//
// Every, not any: a record approves its artifacts as a set. If one of the two registry
// files in APR-0010 changed, the approval of that pair no longer describes what is on
// disk, even though the other file is untouched.
func allCurrent(root string, rec Record) (bool, string) {
	unverifiable := 0
	for _, art := range rec.Artifacts {
		declared := recordedDigest(art.ContentHash)
		actual, ok := Digest(root, art)
		if !ok {
			// The scheme does not define a digest for this file -- a markdown-body hash
			// over a YAML file, among others. It is not evidence of currency and it is
			// not evidence of staleness, so it must not be counted as either.
			unverifiable++
			continue
		}
		if declared != actual {
			return false, fmt.Sprintf("%s approved an earlier version of %s", rec.ID, art.Path)
		}
	}
	if unverifiable == len(rec.Artifacts) {
		return true, fmt.Sprintf("%s (currency unverifiable)", rec.ID)
	}
	if unverifiable > 0 {
		return true, fmt.Sprintf("%s (%d of %d artifact(s) unverifiable)", rec.ID, unverifiable, len(rec.Artifacts))
	}
	return true, rec.ID
}

// recordedDigest pulls the hex digest out of a `scheme:algo:hex` content_hash.
func recordedDigest(h string) string {
	if i := strings.LastIndex(h, ":"); i >= 0 {
		return h[i+1:]
	}
	return h
}

// Digest recomputes an artifact's hash from the working tree, returning ok=false when
// the artifact's scheme defines no digest for that file.
//
// Which digest to compute comes off the record's `kind`, never off the file extension.
// Guessing from the extension would silently pick the wrong scheme the moment a config
// artifact is a markdown file, and the mismatch would read as staleness rather than as
// the bug it is.
func Digest(root string, art Artifact) (string, bool) {
	raw, err := os.ReadFile(filepath.Join(root, art.Path))
	if err != nil {
		return "", false
	}
	if art.Kind == "config" {
		return fileHash(raw), true // legacy/file-32
	}
	return bodyHash(art.Path, raw) // legacy/truncated-32
}

// fileHash is `legacy/file-32`: sha256 over the whole file, truncated to 32 hex chars.
// Raw bytes, because that is how the digest in a record was computed.
func fileHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])[:32]
}

// bodyHash is `legacy/truncated-32`: sha256 over the markdown body with front matter
// excluded, truncated to 32 hex chars.
//
// Front matter is excluded because approving a document is approving what it says, and
// its status and approval fields change as a consequence of the approval itself. A hash
// covering them could never match the moment it was recorded.
func bodyHash(path string, raw []byte) (string, bool) {
	if !strings.HasSuffix(path, ".md") {
		return "", false
	}
	s := string(raw)
	if !strings.HasPrefix(s, "---\n") {
		return "", false
	}
	end := strings.Index(s[4:], "\n---\n")
	if end == -1 {
		return "", false
	}
	body := s[4+end+len("\n---\n"):]
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])[:32], true
}
