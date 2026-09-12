package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestC24_StartInvokesExecModeInTheWorktree asserts Start runs `codex exec` with the
// prompt as a positional argument, JSONL output, and the session's cwd set to the
// worktree with no stdin handed to the child.
func TestC24_StartInvokesExecModeInTheWorktree(t *testing.T) {
	recDir := fakeCodex(t, fakeOpts{Stream: codexStreamLines("thread-1", "hello")})
	worktree := t.TempDir()

	a := &codexAdapter{}
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
	assertContains(t, argv, "exec")
	assertContains(t, argv, "do the thing")
	assertContains(t, argv, "--json")
	assertFollowedBy(t, argv, "-C", worktree)
	assertFollowedBy(t, argv, "--sandbox", "read-only")
	assertFollowedBy(t, argv, "--ask-for-approval", "never")

	if got := fakeCwd(t, recDir, 1); got != worktree {
		t.Errorf("cwd = %q, want %q", got, worktree)
	}
	if got := fakeStdin(t, recDir, 1); got != "" {
		t.Errorf("stdin = %q, want empty", got)
	}
}

// TestC24_ToolsAllowSelectsWorkspaceWriteSandbox asserts write and shell tokens
// select workspace-write; read-only tokens alone select read-only.
func TestC24_ToolsAllowSelectsWorkspaceWriteSandbox(t *testing.T) {
	recDir := fakeCodex(t, fakeOpts{Stream: codexStreamLines("thread-1", "hello")})
	worktree := t.TempDir()

	a := &codexAdapter{}
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
	assertFollowedBy(t, argv, "--sandbox", "workspace-write")
}

// TestC24_MapToolsReportsAnUnknownToken asserts an unmappable capability token is
// never silently dropped.
func TestC24_MapToolsReportsAnUnknownToken(t *testing.T) {
	_, unknown := codexNeedsWriteSandbox([]string{"repository.read", "fly.to.mars"})
	if len(unknown) != 1 || unknown[0] != "fly.to.mars" {
		t.Fatalf("unknown = %v, want [\"fly.to.mars\"]", unknown)
	}
}

// TestC24_1_ToolsDenyRecognizesKnownTokens asserts a known deny token is accepted
// and an unmappable deny token is an error naming the token, never dropped.
func TestC24_1_ToolsDenyRecognizesKnownTokens(t *testing.T) {
	t.Run("mapped", func(t *testing.T) {
		recDir := fakeCodex(t, fakeOpts{Stream: codexStreamLines("thread-1", "hello")})
		worktree := t.TempDir()

		a := &codexAdapter{}
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

		if fakeCount(t, recDir) != 1 {
			t.Fatalf("fake codex invocations = %d, want 1", fakeCount(t, recDir))
		}
	})

	t.Run("unknown", func(t *testing.T) {
		fakeCodex(t, fakeOpts{Stream: codexStreamLines("thread-1", "hello")})
		worktree := t.TempDir()

		a := &codexAdapter{}
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

// TestC24_JSONLBecomesTypedEvents asserts codex exec --json output is decoded into
// the four typed events, in order, all carrying the thread id as the session id.
func TestC24_JSONLBecomesTypedEvents(t *testing.T) {
	fakeCodex(t, fakeOpts{Stream: codexStreamLines("thread-1", "hello")})
	worktree := t.TempDir()

	a := &codexAdapter{}
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
		if e.SessionID != "thread-1" {
			t.Errorf("event %d: SessionID = %q, want %q", i, e.SessionID, "thread-1")
		}
	}
	if events[1].Tool != "bash -lc ls" {
		t.Errorf("tool_use event Tool = %q, want %q", events[1].Tool, "bash -lc ls")
	}
}

// TestC24_WaitReturnsTheThreadIDAndExitStatus asserts Wait resolves the CLI's thread
// id and exit status, and is idempotent.
func TestC24_WaitReturnsTheThreadIDAndExitStatus(t *testing.T) {
	fakeCodex(t, fakeOpts{Stream: codexStreamLines("thread-1", "done")})
	worktree := t.TempDir()

	a := &codexAdapter{}
	sess, err := a.Start(context.Background(), WorkPacket{Prompt: "x", Worktree: worktree})
	if err != nil {
		t.Fatalf("Start: unexpected error %v", err)
	}
	go drainEvents(sess)

	res, err := sess.Wait()
	if err != nil {
		t.Fatalf("Wait: unexpected error %v", err)
	}
	if res.SessionID != "thread-1" {
		t.Errorf("SessionID = %q, want %q", res.SessionID, "thread-1")
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

// TestC24_NonZeroExitIsAFailedResultNotAGoError asserts a non-zero CLI exit is data
// on the Result, not a Go error hiding the output.
func TestC24_NonZeroExitIsAFailedResultNotAGoError(t *testing.T) {
	fakeCodex(t, fakeOpts{Exit: 2, Stderr: "boom: could not read settings"})
	worktree := t.TempDir()

	a := &codexAdapter{}
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

// TestC24_MissingBinaryFailsStartNamingItAndCreatesNoWorktree asserts Start fails
// before touching the filesystem when the codex binary cannot be resolved.
func TestC24_MissingBinaryFailsStartNamingItAndCreatesNoWorktree(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	worktree := filepath.Join(t.TempDir(), "nonexistent-worktree")

	a := &codexAdapter{}
	sess, err := a.Start(context.Background(), WorkPacket{Prompt: "x", Worktree: worktree})
	if err == nil {
		t.Fatal("Start: want error for a missing binary, got nil")
	}
	if !strings.Contains(err.Error(), "codex") {
		t.Errorf("Start error = %v, want it to name %q", err, "codex")
	}
	if sess != nil {
		t.Errorf("Start: got non-nil Session on error, want nil")
	}
	if _, statErr := os.Stat(worktree); !os.IsNotExist(statErr) {
		t.Errorf("Start must create no worktree; os.Stat(%q) = %v", worktree, statErr)
	}
}

// TestC24_CodexAndClaudeProduceTheSameResultShape asserts ISC-24: both adapters
// return a Result with the same fields populated from an equivalent session.
func TestC24_CodexAndClaudeProduceTheSameResultShape(t *testing.T) {
	worktree := t.TempDir()

	fakeClaude(t, fakeOpts{Stream: streamLines("sess-1", "done")})
	claudeSess, err := (&claudeAdapter{}).Start(context.Background(), WorkPacket{
		Prompt: "x", Worktree: worktree,
	})
	if err != nil {
		t.Fatalf("claude Start: %v", err)
	}
	go drainEvents(claudeSess)
	claudeRes, err := claudeSess.Wait()
	if err != nil {
		t.Fatalf("claude Wait: %v", err)
	}

	fakeCodex(t, fakeOpts{Stream: codexStreamLines("thread-1", "done")})
	codexSess, err := (&codexAdapter{}).Start(context.Background(), WorkPacket{
		Prompt: "x", Worktree: worktree,
	})
	if err != nil {
		t.Fatalf("codex Start: %v", err)
	}
	go drainEvents(codexSess)
	codexRes, err := codexSess.Wait()
	if err != nil {
		t.Fatalf("codex Wait: %v", err)
	}

	if claudeRes.Status != codexRes.Status {
		t.Errorf("Status mismatch: claude=%q codex=%q", claudeRes.Status, codexRes.Status)
	}
	if claudeRes.ExitCode != codexRes.ExitCode {
		t.Errorf("ExitCode mismatch: claude=%d codex=%d", claudeRes.ExitCode, codexRes.ExitCode)
	}
	if claudeRes.Text != codexRes.Text {
		t.Errorf("Text mismatch: claude=%q codex=%q", claudeRes.Text, codexRes.Text)
	}
	if claudeRes.SessionID == "" || codexRes.SessionID == "" {
		t.Errorf("SessionID must be set: claude=%q codex=%q", claudeRes.SessionID, codexRes.SessionID)
	}
}
