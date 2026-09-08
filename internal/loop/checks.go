package loop

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LocalCheck is one command the loop runs in the worktree at step 7.
type LocalCheck struct {
	Name    string
	Command []string
}

// Checker runs a caller-supplied list of local checks. An interface so a test can
// fail one on demand: C25 and C29 are both claims about what the record says when a
// check fails, and making a real `go test` fail on cue is not a fixture anyone can
// maintain.
//
// Run takes the checks to execute as an explicit argument rather than something a
// Checker value carries: step 7 calls the same Checker twice in one pass — once
// against DefaultChecks before the commit, once against PostCommitChecks after it —
// and a single stateless method keeps both calls going through one seam instead of
// needing two differently-configured Checker values.
type Checker interface {
	// Run executes checks with dir as the working directory, writing each one's
	// combined output under outDir and returning one Check per command. A non-zero rc
	// is data: Run returns an error only when a check could not be started.
	Run(ctx context.Context, dir, outDir, prefix string, checks []LocalCheck) ([]Check, error)
}

// DefaultChecks run against the working tree, before the commit: go build, go test,
// and the docs validator all examine what the session actually changed, and any of
// the three failing is grounds to re-enter step 3 in the same pass.
var DefaultChecks = []LocalCheck{
	{"go-build", []string{"go", "build", "./..."}},
	{"go-test", []string{"go", "test", "./..."}},
	{"validate-docs", []string{"python3", "scripts/validate_docs.py"}},
}

// PostCommitChecks run after control.Commit and before control.Push. check_gates.py's
// changed_paths() diffs COMMITS against origin/main (scripts/check_gates.py:58-68), so
// run before the commit it sees nothing to evaluate — a real pass-1 run once logged
// "OK: no changes against origin/main" for exactly that reason, on a session that had
// written under a gate-triggering path. Run WITHOUT --strict: the loop records what
// the check said, and deciding what an open gate means is the human's job at the gate
// (POL-001 M5 blocks the merge on it), never the loop's inside step 7. A post-commit
// check's result is recorded as data only — see the step-7 case in loop.go, which
// never re-enters or blocks on it.
var PostCommitChecks = []LocalCheck{
	{"check-gates", []string{"python3", "scripts/check_gates.py", "--base", "origin/main"}},
}

// CommandChecker runs LocalChecks as subprocesses. Stateless: which checks to run
// comes from Run's argument, so the same CommandChecker value serves both the
// pre-commit and the post-commit call.
type CommandChecker struct{}

// Run's Output on each returned Check is checks/<prefix>-<name>.log, relative to
// outDir (the run directory).
func (c CommandChecker) Run(ctx context.Context, dir, outDir, prefix string, checks []LocalCheck) ([]Check, error) {
	checksDir := filepath.Join(outDir, "checks")
	if err := os.MkdirAll(checksDir, 0o755); err != nil {
		return nil, err
	}

	out := make([]Check, 0, len(checks))
	for _, lc := range checks {
		rel := filepath.Join("checks", fmt.Sprintf("%s-%s.log", prefix, lc.Name))
		abs := filepath.Join(outDir, rel)

		cmd := exec.CommandContext(ctx, lc.Command[0], lc.Command[1:]...)
		cmd.Dir = dir
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		runErr := cmd.Run()

		if err := os.WriteFile(abs, buf.Bytes(), 0o644); err != nil {
			return nil, err
		}

		rc := 0
		if runErr != nil {
			var exitErr *exec.ExitError
			if errors.As(runErr, &exitErr) {
				rc = exitErr.ExitCode()
			} else {
				return nil, fmt.Errorf("check %s could not start: %w", lc.Name, runErr)
			}
		}

		out = append(out, Check{
			Name:    lc.Name,
			Command: strings.Join(lc.Command, " "),
			RC:      rc,
			Output:  rel,
		})
	}
	return out, nil
}
