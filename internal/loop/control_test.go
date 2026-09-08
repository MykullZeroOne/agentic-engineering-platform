package loop

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
)

func TestC2_ControlInterfaceHasNoMergeMethod(t *testing.T) {
	typ := reflect.TypeOf((*Control)(nil)).Elem()

	want := map[string]bool{
		"Worktree":       true,
		"RemoveWorktree": true,
		"Status":         true,
		"Commit":         true,
		"Push":           true,
		"CreatePR":       true,
		"ViewPR":         true,
	}
	if typ.NumMethod() != len(want) {
		names := make([]string, typ.NumMethod())
		for i := range names {
			names[i] = typ.Method(i).Name
		}
		t.Fatalf("Control has %d methods %v, want exactly %d: %v", typ.NumMethod(), names, len(want), want)
	}
	for i := 0; i < typ.NumMethod(); i++ {
		name := typ.Method(i).Name
		if strings.Contains(strings.ToLower(name), "merge") {
			t.Fatalf("Control has a method whose name contains 'merge': %s", name)
		}
		if !want[name] {
			t.Fatalf("Control has an unexpected method: %s", name)
		}
	}
}

// TestC2_NoMergeArgvAcrossTheSuite is a fast local signal: it inspects argvLog as it
// stands when this test runs. TestMain re-checks after every test in the package has
// finished, including the anti-vacuity guard; this test only needs the merge-free
// property, on however much has been recorded so far.
func TestC2_NoMergeArgvAcrossTheSuite(t *testing.T) {
	argvLog.mu.Lock()
	defer argvLog.mu.Unlock()
	for _, argv := range argvLog.lines {
		for _, a := range argv {
			if a == "merge" {
				t.Fatalf("found `merge` as a standalone argv element: %v", argv)
			}
		}
	}
}

func TestC26_GHControlBuildsThePRArgv(t *testing.T) {
	recDir := fakeGH(t, PROpen, "")

	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-q", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	g := GHControl{}
	url, err := g.CreatePR(dir, "feat/1-a-test-item", "feat: a test item", "hello body")
	if err != nil {
		t.Fatalf("CreatePR: %v", err)
	}
	if url != "https://github.com/o/r/pull/1" {
		t.Fatalf("url = %q", url)
	}

	calls := readGHArgv(t, recDir)
	if len(calls) != 1 {
		t.Fatalf("expected exactly one gh invocation, got %d: %v", len(calls), calls)
	}
	argv := calls[0]
	logArgv(append([]string{"gh"}, argv...)...)

	for _, a := range argv {
		if a == "merge" {
			t.Fatalf("gh invoked with merge as an argument: %v", argv)
		}
	}

	if len(argv) < 6 || argv[0] != "pr" || argv[1] != "create" {
		t.Fatalf("argv does not open with `pr create`: %v", argv)
	}
	if argv[2] != "--head" || argv[3] != "feat/1-a-test-item" {
		t.Fatalf("argv missing --head <branch>: %v", argv)
	}
	if argv[4] != "--title" || argv[5] != "feat: a test item" {
		t.Fatalf("argv missing --title <title>: %v", argv)
	}

	bodyFileIdx := -1
	for i, a := range argv {
		if a == "--body-file" && i+1 < len(argv) {
			bodyFileIdx = i + 1
		}
	}
	if bodyFileIdx == -1 {
		t.Fatalf("argv missing --body-file <path>: %v", argv)
	}
	body, err := os.ReadFile(argv[bodyFileIdx])
	if err != nil {
		t.Fatalf("reading body file %s: %v", argv[bodyFileIdx], err)
	}
	if string(body) != "hello body" {
		t.Fatalf("body file content = %q, want %q", body, "hello body")
	}
	if !filepath.IsAbs(argv[bodyFileIdx]) {
		t.Fatalf("body file path %q is not absolute", argv[bodyFileIdx])
	}
}

// TestC26_CommitSubjectKeepsCaseAndCutsAtAWordBoundary pins amendment (4): the
// commit/PR-title subject preserves the title's original case (it is not
// lowercased) and, when it must be cut, cuts at the last word boundary before
// byte 72 with no ellipsis appended.
func TestC26_CommitSubjectKeepsCaseAndCutsAtAWordBoundary(t *testing.T) {
	w := config.WorkItem{Type: "docs", Title: "Mark ISA claim polarity so anti-claims partition out of orphan lists"}
	subj := commitSubject(w)
	if !strings.HasPrefix(subj, "docs: Mark ISA claim polarity") {
		t.Errorf("subj = %q, want prefix %q", subj, "docs: Mark ISA claim polarity")
	}
	if strings.Contains(subj, "...") {
		t.Errorf("subj = %q, contains an ellipsis", subj)
	}
	if len(subj) > 72 {
		t.Errorf("subj = %q, len %d > 72", subj, len(subj))
	}
	if strings.HasSuffix(subj, " ") {
		t.Errorf("subj = %q, ends with a space", subj)
	}

	short := config.WorkItem{Type: "fix", Title: "Short title"}
	got := commitSubject(short)
	want := "fix: Short title"
	if got != want {
		t.Errorf("commitSubject(short) = %q, want %q", got, want)
	}
}

// TestC23_ControlStatusWorksWithARelativeDir pins the fix for a real `devctl run`
// failure: `git -C .agentic/runs/RUN-.../worktree status --porcelain` reported
// "fatal: cannot change to '.../worktree': No such file or directory" when the loop
// was invoked with a relative --root. GHControl.run set cmd.Dir to the relative dir
// AND the caller's argv carried the same relative dir as `-C dir`; git resolved the
// `-C` argument against its own (already-changed) working directory, doubling the
// path onto itself. Before the fix, Status of a relative dir fails with exactly that
// "cannot change to" error; after it, an absolute dir is resolved once and used
// consistently for both cmd.Dir and `-C`.
func TestC23_ControlStatusWorksWithARelativeDir(t *testing.T) {
	parent := t.TempDir()
	const repoName = "repo"
	repoDir := filepath.Join(parent, repoName)
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")

	filePath := filepath.Join(repoDir, "a.txt")
	if err := os.WriteFile(filePath, []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "initial")

	if err := os.WriteFile(filePath, []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(parent)

	g := GHControl{}
	paths, err := g.Status(repoName)
	if err != nil {
		t.Fatalf("Status(%q) with a relative dir: %v", repoName, err)
	}
	found := false
	for _, p := range paths {
		if p == "a.txt" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Status(%q) = %v, want a.txt among the changed paths", repoName, paths)
	}
}
