// Package loop implements ADR-023's universal agent loop: ten observable steps on one
// durable run record, with re-entry and return as edges in a state machine rather than
// paragraphs. A step that cannot be seen in the record did not happen.
package loop

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
	"gopkg.in/yaml.v3"
)

// sha256Hex is the hex-encoded sha256 of b.
func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// StepNames are ADR-023's ten steps in its order. The index is the step number minus
// one; the numbering is the ADR's and this slice is the only place it lives.
var StepNames = [10]string{
	"receive", "hydrate", "analyze", "delegate", "collect",
	"resolve", "validate", "produce", "handoff", "consolidate",
}

// Record is one run's durable state: everything needed to resume it, hand it off, or
// read its evidence with no session alive.
//
// There is no context_manifest key and there must never be one (ADR-021: it is a
// Kairo concept).
type Record struct {
	Schema        string     `yaml:"schema"`
	RunID         string     `yaml:"run_id"`
	AgentIdentity string     `yaml:"agent_identity"`
	RoleVersion   string     `yaml:"role_version"`
	Runtime       Runtime    `yaml:"runtime"`
	Work          Work       `yaml:"work"`
	State         string     `yaml:"state"`
	Complete      bool       `yaml:"complete"`
	Created       string     `yaml:"created"`
	Updated       string     `yaml:"updated"`
	Pass          int        `yaml:"pass"`
	Passes        []Pass     `yaml:"passes"`
	Steps         []Step     `yaml:"steps"`
	Evidence      []Evidence `yaml:"evidence"`
	Handoff       *Handoff   `yaml:"handoff"`
	Stubbed       []string   `yaml:"stubbed"`
	Question      string     `yaml:"question"`
}

// Runtime records which provider and model executed the session, and whether that
// resolution was overridden rather than resolved from project.yaml.
type Runtime struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	Override bool   `yaml:"override"`
}

// Work identifies the work item this run serves and the state it was in when
// dispatched. WorkStateAtDispatch is recorded, never enforced (decision 5).
type Work struct {
	Item                string `yaml:"item"`
	Title               string `yaml:"title"`
	WorkStateAtDispatch string `yaml:"work_state_at_dispatch"`
}

// Pass is one objective/close cycle of the loop. A return is a new Pass on the same
// run, never a new run id.
type Pass struct {
	N         int    `yaml:"pass"`
	Objective string `yaml:"objective"`
	Opened    string `yaml:"opened"`
	Closed    string `yaml:"closed"`
	Outcome   string `yaml:"outcome"`
	PRURL     string `yaml:"pr_url"`
}

// Input is one file step 2 (hydrate) hashed.
type Input struct {
	Path   string `yaml:"path"`
	SHA256 string `yaml:"sha256"`
}

// Delegate is one thing step 4 (delegate) handed work to.
type Delegate struct {
	Kind      string `yaml:"kind"`
	Provider  string `yaml:"provider"`
	SessionID string `yaml:"session_id"`
}

// Check is one local check and its result. RC is recorded, never inferred, and Output
// is a path relative to the run directory rather than the output itself: a record
// that inlines a failing test suite stops being readable.
type Check struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	RC      int    `yaml:"rc"`
	Output  string `yaml:"output"`
}

// Gate is the human-owned gate step 7 records having reached.
type Gate struct {
	ID         string `yaml:"id"`
	HumanOwned bool   `yaml:"human_owned"`
	Surface    string `yaml:"surface"`
}

// Step is one recorded entry for one of ADR-023's ten steps. Resume and re-entry both
// append a second entry for the same step number; Record.Step returns the newest one.
type Step struct {
	N            int        `yaml:"n"`
	Name         string     `yaml:"name"`
	Pass         int        `yaml:"pass"`
	Entered      string     `yaml:"entered"`
	Exited       string     `yaml:"exited"`
	Note         string     `yaml:"note,omitempty"`
	Inputs       []Input    `yaml:"inputs,omitempty"`
	Delegates    []Delegate `yaml:"delegates,omitempty"`
	SessionID    string     `yaml:"session_id,omitempty"`
	ExitStatus   *int       `yaml:"exit_status,omitempty"`
	FilesChanged []string   `yaml:"files_changed,omitempty"`
	Questions    []string   `yaml:"questions,omitempty"`
	Checks       []Check    `yaml:"checks,omitempty"`
	Gate         *Gate      `yaml:"gate,omitempty"`
	Artifacts    []string   `yaml:"artifacts,omitempty"`
}

// Evidence is one piece of proof the run produced, keyed to the run rather than the
// session (ISC-26): it must read from disk with no session alive.
type Evidence struct {
	ID        string `yaml:"id"`
	RunID     string `yaml:"run_id"`
	Pass      int    `yaml:"pass"`
	Kind      string `yaml:"kind"`
	SessionID string `yaml:"session_id,omitempty"`
	Path      string `yaml:"path"`
	Recorded  string `yaml:"recorded"`
}

// Handoff is the pull request a pass opened. Nil until step 7 opens one.
type Handoff struct {
	PRURL      string `yaml:"pr_url"`
	Branch     string `yaml:"branch"`
	HeadCommit string `yaml:"head_commit"`
	Opened     string `yaml:"opened"`
	State      string `yaml:"state"`
}

// RunsDir returns .agentic/runs/.
func RunsDir(r config.Root) string {
	return r.Agentic("runs")
}

// RunDir returns .agentic/runs/<runID>.
func RunDir(r config.Root, runID string) string {
	return filepath.Join(RunsDir(r), runID)
}

// RecordPath returns .agentic/runs/<runID>/record.yaml.
func RecordPath(r config.Root, runID string) string {
	return filepath.Join(RunDir(r, runID), "record.yaml")
}

// WorktreePath returns .agentic/runs/<runID>/worktree.
func WorktreePath(r config.Root, runID string) string {
	return filepath.Join(RunDir(r, runID), "worktree")
}

// LoadRecord reads a run record from disk. A missing record wraps os.ErrNotExist.
func LoadRecord(r config.Root, runID string) (*Record, error) {
	b, err := os.ReadFile(RecordPath(r, runID))
	if err != nil {
		return nil, fmt.Errorf("run %q: %w", runID, err)
	}
	var rec Record
	if err := yaml.Unmarshal(b, &rec); err != nil {
		return nil, fmt.Errorf("parse %s: %w", RecordPath(r, runID), err)
	}
	return &rec, nil
}

// SaveRecord stamps Updated and writes atomically: a temp file in the run directory,
// then a rename. A record half-written by an interrupted run is a run that cannot
// resume, and C28 is the claim that says it must.
func SaveRecord(r config.Root, rec *Record, now time.Time) error {
	rec.Updated = now.UTC().Format(time.RFC3339)

	dir := RunDir(r, rec.RunID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	b, err := yaml.Marshal(rec)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".record-*.yaml.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, RecordPath(r, rec.RunID))
}

// FindRun returns the run id currently open for a work item. One run per item at a
// time (decision 1), so at most one is returned: the newest un-complete run whose
// Work.Item matches.
func FindRun(r config.Root, item string) (string, bool, error) {
	entries, err := os.ReadDir(RunsDir(r))
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rec, err := LoadRecord(r, e.Name())
		if err != nil {
			// A run directory with no readable record is not a candidate; it is not
			// this loop's job to repair a corrupt run.
			continue
		}
		if rec.Work.Item == item && !rec.Complete {
			return e.Name(), true, nil
		}
	}
	return "", false, nil
}

var digitsRE = regexp.MustCompile(`\d+`)

// NewRunID returns the next run id for a work item: RUN-<digits>-<seq>, seq from 1.
func NewRunID(r config.Root, item string) (string, error) {
	digits := digitsRE.FindString(item)
	if digits == "" {
		digits = item
	}
	prefix := "RUN-" + digits + "-"

	entries, err := os.ReadDir(RunsDir(r))
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if n, err := strconv.Atoi(strings.TrimPrefix(name, prefix)); err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("%s%d", prefix, max+1), nil
}

// Step returns the LAST recorded entry for step n in pass p, and whether one exists.
// Last, not first: re-entry and resume both append a second entry for the same step,
// and the current one is always the newest.
func (rec *Record) Step(pass, n int) (*Step, bool) {
	var found *Step
	for i := range rec.Steps {
		s := &rec.Steps[i]
		if s.Pass == pass && s.N == n {
			found = s
		}
	}
	if found == nil {
		return nil, false
	}
	return found, true
}

// Resume returns the step number the loop should execute next in the current pass:
// the lowest-numbered step with no recorded entry, or the number of the newest entry
// that was entered and never exited.
func (rec *Record) Resume() int {
	for n := 1; n <= len(StepNames); n++ {
		s, ok := rec.Step(rec.Pass, n)
		if !ok {
			return n
		}
		if s.Exited == "" {
			return n
		}
	}
	return len(StepNames)
}

// currentPass returns the Pass entry matching rec.Pass, or nil if none exists yet.
func (rec *Record) currentPass() *Pass {
	for i := range rec.Passes {
		if rec.Passes[i].N == rec.Pass {
			return &rec.Passes[i]
		}
	}
	return nil
}

// AddEvidence appends an entry, keyed to the run rather than the session (ISC-26).
func (rec *Record) AddEvidence(kind, sessionID, path string, now time.Time) {
	rec.Evidence = append(rec.Evidence, Evidence{
		ID:        fmt.Sprintf("EV-%d", len(rec.Evidence)+1),
		RunID:     rec.RunID,
		Pass:      rec.Pass,
		Kind:      kind,
		SessionID: sessionID,
		Path:      path,
		Recorded:  now.UTC().Format(time.RFC3339),
	})
}
