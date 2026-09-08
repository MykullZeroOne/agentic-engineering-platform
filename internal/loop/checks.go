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

// Checker runs the local checks. An interface so a test can fail one on demand: C25
// and C29 are both claims about what the record says when a check fails, and making a
// real `go test` fail on cue is not a fixture anyone can maintain.
type Checker interface {
	// Run executes the checks with dir as the working directory, writing each one's
	// combined output under outDir and returning one Check per command. A non-zero rc
	// is data: Run returns an error only when a check could not be started.
	Run(ctx context.Context, dir, outDir, prefix string) ([]Check, error)
}

// DefaultChecks are the checks the first mile runs. check_gates runs WITHOUT
// --strict: the loop records what a gate check said, and deciding what an open gate
// means is the human's job at the gate, not the loop's inside it.
var DefaultChecks = []LocalCheck{
	{"go-build", []string{"go", "build", "./..."}},
	{"go-test", []string{"go", "test", "./..."}},
	{"validate-docs", []string{"python3", "scripts/validate_docs.py"}},
	{"check-gates", []string{"python3", "scripts/check_gates.py", "--base", "origin/main"}},
}

// CommandChecker runs LocalChecks as subprocesses.
type CommandChecker struct{ Checks []LocalCheck }

// Run's Output on each returned Check is checks/<prefix>-<name>.log, relative to
// outDir (the run directory).
func (c CommandChecker) Run(ctx context.Context, dir, outDir, prefix string) ([]Check, error) {
	checksDir := filepath.Join(outDir, "checks")
	if err := os.MkdirAll(checksDir, 0o755); err != nil {
		return nil, err
	}

	out := make([]Check, 0, len(c.Checks))
	for _, lc := range c.Checks {
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
