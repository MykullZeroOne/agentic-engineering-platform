package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runFixture builds a minimal project root with a valid engineer.primary role and
// project.yaml, but no work items, so a caller can add exactly the one it needs.
func runFixture(t *testing.T, projectYAML string) string {
	t.Helper()
	root := t.TempDir()

	files := map[string]string{
		".agentic/project.yaml": projectYAML,
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
`,
		".agentic/registries/states.yaml": `
axes:
  agent_state:
    values:
      - {token: working}
      - {token: awaiting_human}
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
		"docs/context/CONTEXT_PACKET.md": "# Context packet\n\nFixture content.\n",
	}
	for rel, body := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cmd := exec.Command("git", "init", "-q", root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return root
}

// captureStderr redirects os.Stderr to a pipe for the duration of fn, and returns
// everything written to it.
func captureStderr(t *testing.T, fn func() int) (int, string) {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	rc := fn()

	w.Close()
	os.Stderr = orig

	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	return rc, string(buf[:n])
}

func TestC18_RunRefusesAnUnknownWorkItem(t *testing.T) {
	root := runFixture(t, `
registries:
  gates: .agentic/registries/gates.yaml
  vocabularies: .agentic/registries/vocabularies.yaml
  states: .agentic/registries/states.yaml
work_store: local
human_gates:
  - platform_config
runtime_preferences:
  implementation: claude-subscription
`)

	rc, stderr := captureStderr(t, func() int {
		return runRun([]string{"WI-9999", "--root", root})
	})

	if rc != 1 {
		t.Fatalf("runRun exit = %d, want 1 (stderr: %s)", rc, stderr)
	}
	if !strings.Contains(stderr, "WI-9999") {
		t.Errorf("stderr = %q, missing WI-9999", stderr)
	}
	if _, err := os.Stat(filepath.Join(root, ".agentic", "runs")); !os.IsNotExist(err) {
		t.Errorf(".agentic/runs/ exists after an unknown-item refusal: %v", err)
	}
}

func TestC32_NoAdapterForTheResolvedProviderExitsTwo(t *testing.T) {
	root := runFixture(t, `
registries:
  gates: .agentic/registries/gates.yaml
  vocabularies: .agentic/registries/vocabularies.yaml
  states: .agentic/registries/states.yaml
work_store: local
human_gates:
  - platform_config
runtime_preferences:
  implementation: codex-subscription
`)
	itemPath := filepath.Join(root, ".agentic", "work", "WI-0001.yaml")
	if err := os.MkdirAll(filepath.Dir(itemPath), 0o755); err != nil {
		t.Fatal(err)
	}
	item := `
id: WI-0001
store: local
project: PRJ-test/repo
type: feat
work_state: ready
priority: normal
title: A test work item
description: A work item used only by devctl's own tests.
`
	if err := os.WriteFile(itemPath, []byte(item), 0o644); err != nil {
		t.Fatal(err)
	}

	rc, stderr := captureStderr(t, func() int {
		return runRun([]string{"WI-0001", "--root", root})
	})

	if rc != 2 {
		t.Fatalf("runRun exit = %d, want 2 (stderr: %s)", rc, stderr)
	}
	if !strings.Contains(stderr, "codex-subscription") {
		t.Errorf("stderr = %q, missing codex-subscription", stderr)
	}
	if !strings.Contains(stderr, "--runtime") {
		t.Errorf("stderr = %q, missing --runtime", stderr)
	}
	if _, err := os.Stat(filepath.Join(root, ".agentic", "runs")); !os.IsNotExist(err) {
		t.Errorf(".agentic/runs/ exists after a no-adapter refusal: %v", err)
	}
}
