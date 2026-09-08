package loop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/runtime"
)

// argvLog collects every command every fake in this package was asked to run, for the
// whole suite. It exists for C2: "zero merge invocations" is a claim about the suite,
// not about one test, and a per-test assertion would leave the one path nobody
// covered free to merge.
var argvLog struct {
	mu    sync.Mutex
	lines [][]string
}

func logArgv(argv ...string) {
	argvLog.mu.Lock()
	defer argvLog.mu.Unlock()
	cp := append([]string(nil), argv...)
	argvLog.lines = append(argvLog.lines, cp)
}

// TestMain inspects argvLog after every test has finished.
//
// It fails the package on two conditions: any recorded argv containing `merge` as a
// standalone argument, and an EMPTY log. The second is the anti-vacuity guard -- a
// probe that passes because nothing was recorded proves nothing, and this suite has
// to be able to tell those two apart.
func TestMain(m *testing.M) {
	code := m.Run()

	argvLog.mu.Lock()
	lines := append([][]string(nil), argvLog.lines...)
	argvLog.mu.Unlock()

	if len(lines) == 0 {
		fmt.Println("TestMain: anti-vacuity guard FAILED: argvLog is empty; no command was ever recorded, so the C2 no-merge check proves nothing.")
		if code == 0 {
			code = 1
		}
	} else {
		foundMerge := false
		for _, argv := range lines {
			for _, a := range argv {
				if a == "merge" {
					fmt.Printf("TestMain: FOUND `merge` as a standalone argv element: %v\n", argv)
					foundMerge = true
				}
			}
		}
		if foundMerge {
			code = 1
		} else {
			fmt.Printf("TestMain: anti-vacuity guard held: argvLog recorded %d command(s) across the suite, none containing `merge` as a standalone argument.\n", len(lines))
		}
	}

	os.Exit(code)
}

// files is a fixture's contents, relative path to body, matching internal/doctor's
// fixture style. There is no testdata/ in this repository and none is introduced
// here.
type files map[string]string

func writeFiles(t *testing.T, root string, f files) {
	t.Helper()
	for rel, body := range f {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func baseFixtureFiles() files {
	return files{
		".agentic/project.yaml": `
registries:
  gates: .agentic/registries/gates.yaml
  vocabularies: .agentic/registries/vocabularies.yaml
  states: .agentic/registries/states.yaml
work_store: local
human_gates:
  - platform_config
runtime_preferences:
  implementation: claude-subscription
`,
		".agentic/registries/gates.yaml": `
gates:
  - id: platform_config
    risk_tier: reversible
`,
		".agentic/registries/vocabularies.yaml": `
vocabularies:
  risk_tier:
    terms:
      - {token: reversible}
      - {token: irreversible}
`,
		".agentic/registries/states.yaml": `
axes:
  agent_state:
    values:
      - {token: working}
      - {token: waiting}
      - {token: blocked}
      - {token: awaiting_human}
      - {token: idle}
      - {token: failed}
  work_state:
    values:
      - {token: intake}
      - {token: ready}
      - {token: done}
`,
		".agentic/roles/engineer.primary.yaml": `
id: engineer.primary
role: engineer
role_version: 1
function: implementation
parent: null
specialists: []
capabilities:
  - implementation
tools:
  allow:
    - repository.read
    - repository.write
    - shell.exec
  deny:
    - github.merge
memory:
  namespace: agent/engineer.primary
  inherit:
    - project
completion:
  human_owned: true
  criteria:
    - local_checks_pass
returns_from:
  - human
human_gates: []
`,
		".agentic/work/WI-0001.yaml": `
id: WI-0001
store: local
project: PRJ-test/repo
type: feat
work_state: ready
priority: normal
title: A test work item
description: |
  A work item used only by internal/loop's own tests.
`,
		"docs/context/CONTEXT_PACKET.md": "# Context packet\n\nFixture content for internal/loop's tests.\n",
	}
}

// fixture builds a temp project root holding .agentic/project.yaml, the three
// registries, .agentic/roles/engineer.primary.yaml, .agentic/work/WI-0001.yaml and
// docs/context/CONTEXT_PACKET.md, initialises it as a git repository with one commit,
// and returns the root.
func fixture(t *testing.T, extra files) config.Root {
	t.Helper()
	root := t.TempDir()
	writeFiles(t, root, baseFixtureFiles())
	writeFiles(t, root, extra)

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	runGit("init", "-q")
	runGit("config", "user.email", "loop-tests@example.com")
	runGit("config", "user.name", "loop tests")
	runGit("add", "-A")
	runGit("commit", "-q", "-m", "initial")

	return config.Root(root)
}

// treeHash returns a sha256 over every file under dir, path and content both, so a
// probe can assert a directory is byte-identical rather than merely still present.
func treeHash(t *testing.T, dir string) string {
	t.Helper()
	h := sha256.New()
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		fmt.Fprintln(h, rel)
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		h.Write(b)
		return nil
	})
	return hex.EncodeToString(h.Sum(nil))
}

// fixedNow returns a Now func for Options that advances one second per call, so a
// record's timestamps are strictly increasing and assertable.
func fixedNow() func() time.Time {
	base := time.Date(2026, 9, 8, 15, 11, 5, 0, time.UTC)
	n := 0
	return func() time.Time {
		n++
		return base.Add(time.Duration(n) * time.Second)
	}
}

// fakeControl is a Control that records every call as an argv-shaped line and answers
// from a script the test sets.
type fakeControl struct {
	PR          PR   // what ViewPR returns once PRSeq is exhausted (or always, if PRSeq is empty)
	PRSeq       []PR // successive ViewPR answers, when a test needs the state to change
	Calls       []string
	CreatedBody string
	StatusFiles []string // what Status returns

	viewCalls int
}

func (f *fakeControl) Worktree(repo, dir, branch, base string) error {
	f.Calls = append(f.Calls, "Worktree")
	logArgv("git", "-C", repo, "worktree", "add", "-b", branch, dir, base)
	return os.MkdirAll(dir, 0o755)
}

func (f *fakeControl) RemoveWorktree(repo, dir string) error {
	f.Calls = append(f.Calls, "RemoveWorktree")
	logArgv("git", "-C", repo, "worktree", "remove", "--force", dir)
	return os.RemoveAll(dir)
}

func (f *fakeControl) Status(dir string) ([]string, error) {
	f.Calls = append(f.Calls, "Status")
	logArgv("git", "-C", dir, "status", "--porcelain")
	return f.StatusFiles, nil
}

func (f *fakeControl) Commit(dir, msg string) (string, error) {
	f.Calls = append(f.Calls, "Commit")
	logArgv("git", "-C", dir, "commit", "-m", msg)
	return "abc1234", nil
}

func (f *fakeControl) Push(dir, branch string) error {
	f.Calls = append(f.Calls, "Push")
	logArgv("git", "-C", dir, "push", "-u", "origin", branch)
	return nil
}

func (f *fakeControl) CreatePR(dir, branch, title, body string) (string, error) {
	f.Calls = append(f.Calls, "CreatePR")
	logArgv("gh", "pr", "create", "--head", branch, "--title", title)
	f.CreatedBody = body
	return "https://github.com/o/r/pull/1", nil
}

func (f *fakeControl) ViewPR(dir, url string) (PR, error) {
	f.Calls = append(f.Calls, "ViewPR")
	logArgv("gh", "pr", "view", url)
	if f.viewCalls < len(f.PRSeq) {
		pr := f.PRSeq[f.viewCalls]
		f.viewCalls++
		return pr, nil
	}
	if len(f.PRSeq) > 0 {
		return f.PRSeq[len(f.PRSeq)-1], nil
	}
	return f.PR, nil
}

// fakeChecker returns a scripted sequence of check results, one entry per
// invocation, repeating the last once the script runs out. It writes a real log file
// for every check so a probe asserting the Output path resolves to a file on disk
// (C25, C5) has something to find.
type fakeChecker struct {
	Script [][]Check
	calls  int
}

func (f *fakeChecker) Run(ctx context.Context, dir, outDir, prefix string) ([]Check, error) {
	idx := f.calls
	if idx >= len(f.Script) {
		idx = len(f.Script) - 1
	}
	f.calls++
	logArgv("fakeChecker", prefix)

	checksDir := filepath.Join(outDir, "checks")
	if err := os.MkdirAll(checksDir, 0o755); err != nil {
		return nil, err
	}

	out := make([]Check, 0, len(f.Script[idx]))
	for _, c := range f.Script[idx] {
		rel := filepath.Join("checks", fmt.Sprintf("%s-%s.log", prefix, c.Name))
		abs := filepath.Join(outDir, rel)
		if err := os.WriteFile(abs, []byte(fmt.Sprintf("rc=%d\n", c.RC)), 0o644); err != nil {
			return nil, err
		}
		cc := c
		cc.Output = rel
		out = append(out, cc)
	}
	return out, nil
}

// fakeSession is the scripted behaviour of one runtime session: the events it emits,
// the Result it returns, and an optional block so a test can cancel mid-session.
type fakeSession struct {
	SessionID string
	Texts     []string // becomes one EventText each; a "QUESTION: …" line goes here
	Result    runtime.Result
	Block     chan struct{} // when non-nil, Wait blocks on it or on ctx
}

// fakeAdapter is a runtime.Adapter returning a scripted session per call to Start,
// clamped to the last entry once the script runs out.
type fakeAdapter struct {
	Sessions []fakeSession
	// Packets records every WorkPacket a call to Start was given, in call order, so a
	// test can inspect what a delegate attempt was actually told (e.g. the rendered
	// prompt on a re-entry).
	Packets []runtime.WorkPacket
	started int
}

func (a *fakeAdapter) Start(ctx context.Context, p runtime.WorkPacket) (runtime.Session, error) {
	idx := a.started
	if idx >= len(a.Sessions) {
		idx = len(a.Sessions) - 1
	}
	a.started++
	a.Packets = append(a.Packets, p)
	spec := a.Sessions[idx]
	logArgv("claude", "-p", "<work packet>", "--allowedTools", fmt.Sprintf("%v", p.Tools))

	if p.TranscriptPath != "" {
		if err := os.MkdirAll(filepath.Dir(p.TranscriptPath), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(p.TranscriptPath, []byte("fake transcript\n"), 0o644); err != nil {
			return nil, err
		}
	}

	s := &fakeRunningSession{
		id:     spec.SessionID,
		result: spec.Result,
		ctx:    ctx,
		events: make(chan runtime.Event),
		done:   make(chan struct{}),
	}
	go s.run(spec.Texts, spec.Block)
	return s, nil
}

type fakeRunningSession struct {
	id     string
	result runtime.Result
	ctx    context.Context
	events chan runtime.Event
	done   chan struct{}
}

func (s *fakeRunningSession) run(texts []string, block chan struct{}) {
	defer close(s.events)
	defer close(s.done)

	select {
	case s.events <- runtime.Event{Kind: runtime.EventStarted, SessionID: s.id}:
	case <-s.ctx.Done():
		return
	}
	for _, t := range texts {
		select {
		case s.events <- runtime.Event{Kind: runtime.EventText, SessionID: s.id, Text: t}:
		case <-s.ctx.Done():
			return
		}
	}
	if block != nil {
		select {
		case <-block:
		case <-s.ctx.Done():
		}
	}
}

func (s *fakeRunningSession) ID() string                   { return s.id }
func (s *fakeRunningSession) Status() runtime.Status       { return s.result.Status }
func (s *fakeRunningSession) Events() <-chan runtime.Event { return s.events }
func (s *fakeRunningSession) Wait() (runtime.Result, error) {
	<-s.done
	if s.ctx.Err() != nil {
		return runtime.Result{SessionID: s.id, Status: runtime.StatusFailed}, nil
	}
	return s.result, nil
}

// fakeGH installs a fake `gh` first on PATH, recording every argv into recDir/argv.N.
// Its answers are controlled by env vars the test writes: `pr view` returns them as
// the state and review, `pr create` prints a fixed URL.
func fakeGH(t *testing.T, state PRState, review string) (recDir string) {
	t.Helper()
	dir := t.TempDir()
	recDir = filepath.Join(dir, "rec")
	if err := os.MkdirAll(recDir, 0o755); err != nil {
		t.Fatal(err)
	}

	script := `#!/bin/sh
d="$FAKE_GH_DIR"; mkdir -p "$d"
i=$(cat "$d/n" 2>/dev/null || echo 0); i=$((i + 1)); echo "$i" > "$d/n"
for a in "$@"; do printf '%s\n' "$a"; done > "$d/argv.$i"
case "$2" in
  create) echo "https://github.com/o/r/pull/1" ;;
  view)   printf '{"state":"%s","url":"https://github.com/o/r/pull/1","mergeCommit":{"oid":"abc1234"},"reviews":[{"body":"%s"}],"comments":[]}\n' \
            "$(cat "$FAKE_GH_STATE")" "$FAKE_GH_REVIEW" ;;
esac
exit 0
`
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	statePath := filepath.Join(dir, "state")
	if err := os.WriteFile(statePath, []byte(state), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH_DIR", recDir)
	t.Setenv("FAKE_GH_STATE", statePath)
	t.Setenv("FAKE_GH_REVIEW", review)

	return recDir
}

// readGHArgv reads every argv file fakeGH recorded, in call order.
func readGHArgv(t *testing.T, recDir string) [][]string {
	t.Helper()
	entries, err := os.ReadDir(recDir)
	if err != nil {
		t.Fatal(err)
	}
	var calls [][]string
	for _, e := range entries {
		if e.Name() == "n" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(recDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var argv []string
		for _, line := range splitLines(string(b)) {
			argv = append(argv, line)
		}
		calls = append(calls, argv)
	}
	return calls
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
