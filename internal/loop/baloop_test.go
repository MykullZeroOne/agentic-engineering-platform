package loop

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/runtime"
)

func baFixture(t *testing.T) config.Root {
	root := fixture(t, map[string]string{
		".agentic/project.yaml": `
version: 1
project: test-project
control_plane:
  provider: github
  repository: test/repo
registries:
  gates: .agentic/registries/gates.yaml
  hook_points: .agentic/registries/hook-points.yaml
  states: .agentic/registries/states.yaml
  vocabularies: .agentic/registries/vocabularies.yaml
work_store: local
runtime_preferences:
  planning: claude-subscription
  implementation: codex-subscription
human_gates:
  - product_spec
`,
		".agentic/roles/ba.primary.yaml": `
id: ba.primary
role: ba
role_version: 1
function: planning
parent: null
specialists:
  - compliance.primary
  - legal.preflight
capabilities:
  - product-analysis
tools:
  allow:
    - repository.read
  deny:
    - repository.write
    - shell.exec
    - github.merge
memory:
  namespace: agent/ba.primary
  inherit:
    - project
completion:
  human_owned: true
  criteria:
    - readiness_gate_satisfied
human_gates:
  - product_spec
returns_from:
  - human
`,
	})
	return root
}

func questionAdapter(sessionID string) *fakeAdapter {
	return &fakeAdapter{Sessions: []fakeSession{{
		SessionID: sessionID,
		Texts:     []string{"QUESTION: Who is the primary user for this feature?"},
		Result:    runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0},
	}}}
}

func specOutlineAdapter(sessionID string) *fakeAdapter {
	return &fakeAdapter{Sessions: []fakeSession{{
		SessionID: sessionID,
		Texts:     []string{"SPEC OUTLINE: objective clear; actors: engineers; success: faster onboarding"},
		Result:    runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0},
	}}}
}

func TestBA_IdeaYieldsAQuestionBeforeSpecification(t *testing.T) {
	root := baFixture(t)
	out, err := RunIntent(context.Background(), IntentOptions{
		Root:    root,
		Idea:    "Build a better onboarding flow for new contributors",
		Adapter: questionAdapter("ba-sess-1"),
		Now:     fixedNow(),
	})
	if err != nil {
		t.Fatalf("RunIntent: %v", err)
	}
	if out.Exit != 3 {
		t.Fatalf("exit = %d, want 3 (blocked on question)", out.Exit)
	}
	if out.Record.State != "blocked" {
		t.Fatalf("state = %q, want blocked", out.Record.State)
	}
	if out.Record.Question == "" {
		t.Fatal("expected a recorded question")
	}
	s3, ok := out.Record.Step(1, 3)
	if !ok {
		t.Fatal("missing step 3")
	}
	if !strings.Contains(s3.Note, "compliance.primary") {
		t.Errorf("step 3 note = %q, want specialist routing", s3.Note)
	}
	compliance := filepath.Join(RunDir(root, out.RunID), "specialists", "compliance.md")
	if _, err := os.Stat(compliance); err != nil {
		t.Fatalf("compliance report missing: %v", err)
	}
}

func TestBA_WithholdingApprovalLeavesRunOpen(t *testing.T) {
	root := baFixture(t)
	out, err := RunIntent(context.Background(), IntentOptions{
		Root:    root,
		Idea:    "Add intent capture to devctl",
		Adapter: specOutlineAdapter("ba-sess-2"),
		Now:     fixedNow(),
	})
	if err != nil {
		t.Fatalf("RunIntent: %v", err)
	}
	if out.Exit != 0 {
		t.Fatalf("exit = %d, want 0 (parked awaiting human)", out.Exit)
	}
	if out.Record.Complete {
		t.Fatal("run must not be complete without human approval")
	}
	if out.Record.State != "awaiting_human" {
		t.Fatalf("state = %q, want awaiting_human", out.Record.State)
	}
	s7, ok := out.Record.Step(1, 7)
	if !ok || s7.Gate == nil || s7.Gate.ID != "product_spec" {
		t.Fatalf("step 7 gate = %v, want product_spec", s7.Gate)
	}
	specPath := filepath.Join(RunDir(root, out.RunID), "handoff", "pass-1-spec-review.md")
	if _, err := os.Stat(specPath); err != nil {
		t.Fatalf("spec review artifact missing: %v", err)
	}
}

func TestBA_ResumeAfterAnswerContinues(t *testing.T) {
	root := baFixture(t)
	adapter := &fakeAdapter{Sessions: []fakeSession{
		{
			SessionID: "ba-1",
			Texts:     []string{"QUESTION: What is the primary success metric?"},
			Result:    runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0},
		},
		{
			SessionID: "ba-2",
			Texts:     []string{"SPEC OUTLINE: metric is time-to-first-PR under one hour"},
			Result:    runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0},
		},
	}}
	first, err := RunIntent(context.Background(), IntentOptions{
		Root: root, Idea: "Speed up contributor onboarding", Adapter: adapter, Now: fixedNow(),
	})
	if err != nil || first.Exit != 3 {
		t.Fatalf("first RunIntent: err=%v exit=%d", err, first.Exit)
	}
	second, err := RunIntent(context.Background(), IntentOptions{
		Root: root, RunID: first.RunID, Answer: "Time to first merged PR under 24 hours",
		Adapter: adapter, Now: func() time.Time { t := fixedNow(); return t().Add(time.Minute) },
	})
	if err != nil {
		t.Fatalf("resume RunIntent: %v", err)
	}
	if second.Exit != 0 || second.Record.State != "awaiting_human" {
		t.Fatalf("resume exit=%d state=%q, want 0 awaiting_human", second.Exit, second.Record.State)
	}
}
