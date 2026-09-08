package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}

// assertFollowedBy fails unless flag appears in argv immediately followed by value.
func assertFollowedBy(t *testing.T, argv []string, flag, value string) {
	t.Helper()
	i := indexOf(argv, flag)
	if i < 0 {
		t.Fatalf("argv %v: missing %q", argv, flag)
	}
	if i+1 >= len(argv) || argv[i+1] != value {
		t.Fatalf("argv %v: %q not followed by %q", argv, flag, value)
	}
}

func assertContains(t *testing.T, argv []string, value string) {
	t.Helper()
	if indexOf(argv, value) < 0 {
		t.Fatalf("argv %v: missing %q", argv, value)
	}
}

// TestC13_StartInvokesPrintModeInTheWorktree asserts Start runs `claude -p` with the
// prompt as a positional argument, stream-json output, and the session's cwd set to
// the worktree with no stdin handed to the child.
func TestC13_StartInvokesPrintModeInTheWorktree(t *testing.T) {
	recDir := fakeClaude(t, fakeOpts{Stream: streamLines("sess-1", "hello")})
	worktree := t.TempDir()

	a := &claudeAdapter{}
	sess, err := a.Start(context.Background(), WorkPacket{
		Prompt:   "do the thing",
		Worktree: worktree,
	})
	if err != nil {
		t.Fatalf("Start: unexpected error %v", err)
	}
	if sess == nil {
		t.Fatal("Start: got nil Session, want non-nil")
	}
	drainEvents(sess)
	if _, err := sess.Wait(); err != nil {
		t.Fatalf("Wait: unexpected error %v", err)
	}

	argv := fakeArgv(t, recDir, 1)
	assertContains(t, argv, "-p")
	assertContains(t, argv, "do the thing")
	assertFollowedBy(t, argv, "--output-format", "stream-json")
	assertContains(t, argv, "--verbose")

	if got := fakeCwd(t, recDir, 1); got != worktree {
		t.Errorf("cwd = %q, want %q", got, worktree)
	}
	if got := fakeStdin(t, recDir, 1); got != "" {
		t.Errorf("stdin = %q, want empty", got)
	}
}

// TestC13_ToolsAllowMapsToAllowedTools asserts the role's tools.allow tokens are
// mapped into the CLI's --allowedTools flag, deduplicated and sorted.
func TestC13_ToolsAllowMapsToAllowedTools(t *testing.T) {
	recDir := fakeClaude(t, fakeOpts{Stream: streamLines("sess-1", "hello")})
	worktree := t.TempDir()

	a := &claudeAdapter{}
	sess, err := a.Start(context.Background(), WorkPacket{
		Prompt:   "do the thing",
		Worktree: worktree,
		Tools:    []string{"repository.read", "repository.write", "shell.exec"},
	})
	if err != nil {
		t.Fatalf("Start: unexpected error %v", err)
	}
	drainEvents(sess)
	if _, err := sess.Wait(); err != nil {
		t.Fatalf("Wait: unexpected error %v", err)
	}

	argv := fakeArgv(t, recDir, 1)
	assertFollowedBy(t, argv, "--allowedTools", "Bash,Edit,Glob,Grep,Read,Write")
}

// TestC13_MapToolsReportsAnUnknownToken asserts an unmappable capability token is
// never silently dropped.
func TestC13_MapToolsReportsAnUnknownToken(t *testing.T) {
	_, unknown := mapTools([]string{"repository.read", "fly.to.mars"})
	if len(unknown) != 1 || unknown[0] != "fly.to.mars" {
		t.Fatalf("unknown = %v, want [\"fly.to.mars\"]", unknown)
	}
}

// TestC13_1_ToolsDenyMapsToDisallowedTools asserts C13.1: the role's tools.deny
// tokens map to --disallowedTools the same way tools.allow maps to --allowedTools,
// and an unmappable deny token is an error naming the token, never dropped.
func TestC13_1_ToolsDenyMapsToDisallowedTools(t *testing.T) {
	t.Run("mapped", func(t *testing.T) {
		recDir := fakeClaude(t, fakeOpts{Stream: streamLines("sess-1", "hello")})
		worktree := t.TempDir()

		a := &claudeAdapter{}
		sess, err := a.Start(context.Background(), WorkPacket{
			Prompt:   "do the thing",
			Worktree: worktree,
			Deny:     []string{"github.merge"},
		})
		if err != nil {
			t.Fatalf("Start: unexpected error %v", err)
		}
		drainEvents(sess)
		if _, err := sess.Wait(); err != nil {
			t.Fatalf("Wait: unexpected error %v", err)
		}

		argv := fakeArgv(t, recDir, 1)
		assertFollowedBy(t, argv, "--disallowedTools", "Bash(gh api *),Bash(gh pr merge*)")
	})

	t.Run("unknown", func(t *testing.T) {
		fakeClaude(t, fakeOpts{Stream: streamLines("sess-1", "hello")})
		worktree := t.TempDir()

		a := &claudeAdapter{}
		sess, err := a.Start(context.Background(), WorkPacket{
			Prompt:   "do the thing",
			Worktree: worktree,
			Deny:     []string{"nope"},
		})
		if err == nil {
			t.Fatal("Start: want error for unmappable deny token, got nil")
		}
		if !strings.Contains(err.Error(), "nope") {
			t.Errorf("Start error = %v, want it to name %q", err, "nope")
		}
		if sess != nil {
			t.Errorf("Start: got non-nil Session on error, want nil")
		}
	})
}

// TestC14_StreamJSONBecomesTypedEvents asserts the CLI's stream-json output is
// decoded into the four typed events, in order, all carrying the session id.
func TestC14_StreamJSONBecomesTypedEvents(t *testing.T) {
	fakeClaude(t, fakeOpts{Stream: streamLines("sess-1", "hello")})
	worktree := t.TempDir()

	a := &claudeAdapter{}
	sess, err := a.Start(context.Background(), WorkPacket{Prompt: "x", Worktree: worktree})
	if err != nil {
		t.Fatalf("Start: unexpected error %v", err)
	}

	events := drainEvents(sess)
	if _, err := sess.Wait(); err != nil {
		t.Fatalf("Wait: unexpected error %v", err)
	}

	wantKinds := []EventKind{EventStarted, EventToolUse, EventText, EventResult}
	if len(events) != len(wantKinds) {
		t.Fatalf("got %d events, want %d: %+v", len(events), len(wantKinds), events)
	}
	for i, e := range events {
		if e.Kind != wantKinds[i] {
			t.Errorf("event %d: Kind = %q, want %q", i, e.Kind, wantKinds[i])
		}
		if e.SessionID != "sess-1" {
			t.Errorf("event %d: SessionID = %q, want %q", i, e.SessionID, "sess-1")
		}
	}
	if events[1].Tool != "Read" {
		t.Errorf("tool_use event Tool = %q, want %q", events[1].Tool, "Read")
	}
}

// TestC14_WaitReturnsTheSessionIDAndExitStatus asserts Wait resolves the CLI's
// session id and exit status, and is idempotent.
func TestC14_WaitReturnsTheSessionIDAndExitStatus(t *testing.T) {
	fakeClaude(t, fakeOpts{Stream: streamLines("sess-1", "done")})
	worktree := t.TempDir()

	a := &claudeAdapter{}
	sess, err := a.Start(context.Background(), WorkPacket{Prompt: "x", Worktree: worktree})
	if err != nil {
		t.Fatalf("Start: unexpected error %v", err)
	}
	go drainEvents(sess)

	res, err := sess.Wait()
	if err != nil {
		t.Fatalf("Wait: unexpected error %v", err)
	}
	if res.SessionID != "sess-1" {
		t.Errorf("SessionID = %q, want %q", res.SessionID, "sess-1")
	}
	if res.Status != StatusSucceeded {
		t.Errorf("Status = %q, want %q", res.Status, StatusSucceeded)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", res.ExitCode)
	}
	if res.Text != "done" {
		t.Errorf("Text = %q, want %q", res.Text, "done")
	}

	res2, err2 := sess.Wait()
	if err2 != nil {
		t.Fatalf("second Wait: unexpected error %v", err2)
	}
	if res2 != res {
		t.Errorf("second Wait() = %+v, want the same Result %+v", res2, res)
	}
}

// TestC15_NonZeroExitIsAFailedResultNotAGoError asserts a non-zero CLI exit is data
// on the Result, not a Go error hiding the output.
func TestC15_NonZeroExitIsAFailedResultNotAGoError(t *testing.T) {
	fakeClaude(t, fakeOpts{Exit: 2, Stderr: "boom: could not read settings"})
	worktree := t.TempDir()

	a := &claudeAdapter{}
	sess, err := a.Start(context.Background(), WorkPacket{Prompt: "x", Worktree: worktree})
	if err != nil {
		t.Fatalf("Start: unexpected error %v", err)
	}
	go drainEvents(sess)

	res, err := sess.Wait()
	if err != nil {
		t.Fatalf("Wait: want nil error for a non-zero exit, got %v", err)
	}
	if res.Status != StatusFailed {
		t.Errorf("Status = %q, want %q", res.Status, StatusFailed)
	}
	if res.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", res.ExitCode)
	}
	if !strings.Contains(res.Stderr, "boom") {
		t.Errorf("Stderr = %q, want it to contain %q", res.Stderr, "boom")
	}
}

// TestC17_MissingBinaryFailsStartNamingItAndCreatesNoWorktree asserts Start fails
// before touching the filesystem when the claude binary cannot be resolved.
func TestC17_MissingBinaryFailsStartNamingItAndCreatesNoWorktree(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	worktree := filepath.Join(t.TempDir(), "nonexistent-worktree")

	a := &claudeAdapter{}
	sess, err := a.Start(context.Background(), WorkPacket{Prompt: "x", Worktree: worktree})
	if err == nil {
		t.Fatal("Start: want error for a missing binary, got nil")
	}
	if !strings.Contains(err.Error(), "claude") {
		t.Errorf("Start error = %v, want it to name %q", err, "claude")
	}
	if sess != nil {
		t.Errorf("Start: got non-nil Session on error, want nil")
	}
	if _, statErr := os.Stat(worktree); !os.IsNotExist(statErr) {
		t.Errorf("Start must create no worktree; os.Stat(%q) = %v", worktree, statErr)
	}
}
