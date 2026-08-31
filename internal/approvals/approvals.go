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
}

// Coverage is how well a gate is closed. The three states are deliberately distinct:
// collapsing Stale into Open would lose the difference between "nobody approved this"
// and "somebody approved an earlier version of it", which are different problems with
// different fixes.
type Coverage string

const (
	// Covered: a record closes the gate and every artifact it names is unchanged.
	Covered Coverage = "covered"
	// Stale: a record closes the gate but an artifact has changed since.
	Stale Coverage = "stale"
	// Open: no record closes the gate at all.
	Open Coverage = "open"
)

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

// Cover reports the best coverage the records give gateID.
//
// Best, not first: a gate re-approved after each material change accumulates records, and
// platform_config already carries four. Reporting the first match would let an early stale
// record mask the current one that actually closes the gate.
func Cover(root, gateID string, records []Record) Result {
	var stale *Result
	for _, rec := range records {
		if rec.Gate != gateID {
			continue
		}
		if len(rec.Artifacts) == 0 {
			// A record naming no artifact pins nothing, so nothing can verify it. Say so
			// rather than counting it either way.
			return Result{Covered, rec.ID, "record names no artifact; currency unverifiable"}
		}
		fresh, reason := allCurrent(root, rec)
		if fresh {
			return Result{Covered, rec.ID, reason}
		}
		if stale == nil {
			stale = &Result{Stale, rec.ID, reason}
		}
	}
	if stale != nil {
		return *stale
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
