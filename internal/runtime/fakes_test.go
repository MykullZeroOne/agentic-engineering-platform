package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// fakeOpts configures the fake claude binary.
type fakeOpts struct {
	Stream []string // stream-json lines to emit on stdout, one per element
	Exit   int      // process exit code
	Stderr string   // text written to stderr
	Sleep  string   // seconds to sleep before emitting anything, for cancel tests
}

// fakeClaude installs a fake `claude` first on PATH for the duration of t and returns
// the directory it records into. Recordings are numbered per invocation: argv.1,
// cwd.1, stdin.1, argv.2, ... so a test that starts two sessions can tell them apart.
func fakeClaude(t *testing.T, o fakeOpts) (recDir string) {
	t.Helper()

	binDir := t.TempDir()
	recDir = t.TempDir()

	script := "#!/bin/sh\n" +
		"# Fake `claude`. Written by fakeClaude; never committed as a fixture.\n" +
		"d=\"$FAKE_CLAUDE_DIR\"\n" +
		"mkdir -p \"$d\"\n" +
		"i=$(cat \"$d/n\" 2>/dev/null || echo 0); i=$((i + 1)); echo \"$i\" > \"$d/n\"\n" +
		"for a in \"$@\"; do printf '%s\\n' \"$a\"; done > \"$d/argv.$i\"\n" +
		"pwd > \"$d/cwd.$i\"\n" +
		"cat > \"$d/stdin.$i\"\n" +
		"if [ -n \"$FAKE_CLAUDE_SLEEP\" ]; then sleep \"$FAKE_CLAUDE_SLEEP\"; fi\n" +
		"if [ -n \"$FAKE_CLAUDE_STDERR\" ]; then printf '%s' \"$FAKE_CLAUDE_STDERR\" >&2; fi\n" +
		"if [ -n \"$FAKE_CLAUDE_STREAM\" ]; then cat \"$FAKE_CLAUDE_STREAM\"; fi\n" +
		"exit \"${FAKE_CLAUDE_EXIT:-0}\"\n"

	scriptPath := filepath.Join(binDir, "claude")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("writing fake claude script: %v", err)
	}

	var streamPath string
	if len(o.Stream) > 0 {
		streamPath = filepath.Join(recDir, "stream.jsonl")
		body := strings.Join(o.Stream, "\n") + "\n"
		if err := os.WriteFile(streamPath, []byte(body), 0o644); err != nil {
			t.Fatalf("writing fake stream fixture: %v", err)
		}
	}

	t.Setenv("FAKE_CLAUDE_DIR", recDir)
	t.Setenv("FAKE_CLAUDE_STREAM", streamPath)
	t.Setenv("FAKE_CLAUDE_EXIT", strconv.Itoa(o.Exit))
	t.Setenv("FAKE_CLAUDE_STDERR", o.Stderr)
	t.Setenv("FAKE_CLAUDE_SLEEP", o.Sleep)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return recDir
}

// readRecFile reads one numbered recording file, failing the test if it is missing.
func readRecFile(t *testing.T, recDir, prefix string, n int) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(recDir, fmt.Sprintf("%s.%d", prefix, n)))
	if err != nil {
		t.Fatalf("reading %s.%d: %v", prefix, n, err)
	}
	return string(body)
}

// fakeArgv reads the argv of the n'th invocation, one element per line.
func fakeArgv(t *testing.T, recDir string, n int) []string {
	t.Helper()
	body := readRecFile(t, recDir, "argv", n)
	if body == "" {
		return nil
	}
	lines := strings.Split(body, "\n")
	// The script writes one trailing newline per element; drop the empty element the
	// final newline produces, but only that one, so a genuinely empty argument in the
	// middle of argv is preserved.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// fakeCwd reads the working directory of the n'th invocation.
func fakeCwd(t *testing.T, recDir string, n int) string {
	t.Helper()
	return strings.TrimRight(readRecFile(t, recDir, "cwd", n), "\n")
}

// fakeStdin reads what the n'th invocation was given on stdin.
func fakeStdin(t *testing.T, recDir string, n int) string {
	t.Helper()
	return readRecFile(t, recDir, "stdin", n)
}

// fakeCount reports how many times the fake ran.
func fakeCount(t *testing.T, recDir string) int {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(recDir, "n"))
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("reading fake claude invocation count: %v", err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		t.Fatalf("parsing fake claude invocation count %q: %v", body, err)
	}
	return n
}

// streamLines returns a well-formed stream-json transcript: init, a tool_use, a text
// block, and a result, all carrying sessionID.
func streamLines(sessionID string, texts ...string) []string {
	text := "hello"
	if len(texts) > 0 {
		text = texts[0]
	}
	init := fmt.Sprintf(`{"type":"system","subtype":"init","session_id":%q}`, sessionID)
	assistant := fmt.Sprintf(
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"},{"type":"text","text":%q}]}}`,
		text,
	)
	result := fmt.Sprintf(`{"type":"result","session_id":%q,"is_error":false,"result":%q}`, sessionID, text)
	return []string{init, assistant, result}
}

// drainEvents reads every event off a session until the channel closes.
func drainEvents(sess Session) []Event {
	var events []Event
	for e := range sess.Events() {
		events = append(events, e)
	}
	return events
}
