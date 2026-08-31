package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
)

// The fixtures are built in a temp directory rather than committed under testdata/.
// A committed fixture of .agentic/ would be a second copy of the schema, and this
// repository keeps finding that two copies means one of them is quietly wrong.

type files map[string]string

func write(t *testing.T, root string, f files) {
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

// good is the smallest configuration that should produce no findings.
func good() files {
	return files{
		".agentic/project.yaml": `
registries:
  gates: .agentic/registries/gates.yaml
  vocabularies: .agentic/registries/vocabularies.yaml
  states: .agentic/registries/states.yaml
work_store: local
human_gates:
  - platform_config
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
  work_item_type:
    terms:
      - {token: feat}
  priority:
    terms:
      - {token: normal}
`,
		".agentic/registries/states.yaml": `
axes:
  work_state:
    values:
      - {token: intake}
      - {token: done}
`,
		".agentic/work/WI-0001.yaml": `
id: WI-0001
type: feat
work_state: intake
priority: normal
title: A work item
`,
	}
}

func check(t *testing.T, f files) *Report {
	t.Helper()
	root := t.TempDir()
	write(t, root, f)
	rep, err := Run(config.Root(root))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return rep
}

// findings returns every message, so a test can assert on what was said rather than
// only on how many things were said.
func findings(rep *Report) string {
	var b strings.Builder
	for _, f := range rep.Findings {
		b.WriteString(f.String())
		b.WriteString("\n")
	}
	return b.String()
}

func TestCleanProject(t *testing.T) {
	rep := check(t, good())
	if len(rep.Findings) != 0 {
		t.Fatalf("expected no findings, got:\n%s", findings(rep))
	}
	// A clean run that examined nothing would also report no findings, which is the
	// failure this guards: the counts are what distinguish passing from not looking.
	for _, k := range []string{"gates", "registries", "work_items", "human_gates"} {
		if rep.Checked[k] == 0 {
			t.Errorf("checked nothing for %q; a clean report must say what it examined", k)
		}
	}
}

func TestNotAProjectRoot(t *testing.T) {
	rep := check(t, files{"README.md": "not a project"})
	if rep.Errors() == 0 {
		t.Fatal("a directory with no .agentic/ must be an error, not a pass")
	}
	if !strings.Contains(findings(rep), "not an AEP project root") {
		t.Errorf("finding should name the cause plainly, got:\n%s", findings(rep))
	}
}

func TestRegistryPointerDangling(t *testing.T) {
	f := good()
	f[".agentic/project.yaml"] = strings.Replace(f[".agentic/project.yaml"],
		"gates: .agentic/registries/gates.yaml", "gates: .agentic/registries/missing.yaml", 1)
	rep := check(t, f)
	if rep.Errors() == 0 {
		t.Fatal("a registry pointer at a missing file must be an error")
	}
	if !strings.Contains(findings(rep), "which does not exist") {
		t.Errorf("finding should name the dangling path, got:\n%s", findings(rep))
	}
}

func TestGateMissingRiskTier(t *testing.T) {
	f := good()
	f[".agentic/registries/gates.yaml"] = "gates:\n  - id: platform_config\n"
	rep := check(t, f)
	if !strings.Contains(findings(rep), "no risk_tier") {
		t.Fatalf("a gate without risk_tier must be reported (ADR-014 clause 7), got:\n%s", findings(rep))
	}
}

func TestGateRiskTierNotInVocabulary(t *testing.T) {
	f := good()
	f[".agentic/registries/gates.yaml"] = "gates:\n  - id: platform_config\n    risk_tier: catastrophic\n"
	rep := check(t, f)
	if !strings.Contains(findings(rep), "not in the risk_tier vocabulary") {
		t.Fatalf("an invented tier must be rejected, got:\n%s", findings(rep))
	}
}

func TestHumanGateNamesUnknownGate(t *testing.T) {
	f := good()
	f[".agentic/project.yaml"] = strings.Replace(f[".agentic/project.yaml"],
		"- platform_config", "- no_such_gate", 1)
	rep := check(t, f)
	if !strings.Contains(findings(rep), `unknown gate "no_such_gate"`) {
		t.Fatalf("a human gate naming nothing never fires and must be reported, got:\n%s", findings(rep))
	}
}

func TestWorkStoreUnset(t *testing.T) {
	f := good()
	f[".agentic/project.yaml"] = strings.Replace(f[".agentic/project.yaml"], "work_store: local", "", 1)
	rep := check(t, f)
	if !strings.Contains(findings(rep), "ADR-012") {
		t.Fatalf("an unset work_store must cite why exactly one is required, got:\n%s", findings(rep))
	}
}

func TestWorkStoreGithubIsUnimplementedNotBroken(t *testing.T) {
	f := good()
	f[".agentic/project.yaml"] = strings.Replace(f[".agentic/project.yaml"], "work_store: local", "work_store: github", 1)
	rep := check(t, f)
	if rep.Errors() != 0 {
		t.Fatalf("github is unimplemented, which is not the same as broken; got:\n%s", findings(rep))
	}
	if !strings.Contains(findings(rep), "cannot read yet") {
		t.Errorf("it should still say so, got:\n%s", findings(rep))
	}
}

func TestWorkItemTokensCheckedAgainstRegistries(t *testing.T) {
	f := good()
	f[".agentic/work/WI-0001.yaml"] = `
id: WI-0001
type: invented
work_state: nowhere
priority: screaming
title: A work item
`
	rep := check(t, f)
	out := findings(rep)
	for _, want := range []string{`type "invented"`, `work_state "nowhere"`, `priority "screaming"`} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %s to be rejected, got:\n%s", want, out)
		}
	}
}

func TestDependencyOnUnknownItem(t *testing.T) {
	f := good()
	f[".agentic/work/WI-0001.yaml"] += "dependencies:\n- WI-9999\n"
	rep := check(t, f)
	if !strings.Contains(findings(rep), `unknown work item "WI-9999"`) {
		t.Fatalf("a decorative dependency edge must be reported, got:\n%s", findings(rep))
	}
}

func TestFindingsAccumulate(t *testing.T) {
	// One report per run, not one problem per run: a check that stops at the first
	// finding costs a round trip for every additional one.
	f := good()
	f[".agentic/registries/gates.yaml"] = "gates:\n  - id: platform_config\n"
	f[".agentic/work/WI-0001.yaml"] += "dependencies:\n- WI-9999\n"
	rep := check(t, f)
	if len(rep.Findings) < 2 {
		t.Fatalf("expected findings from both checks, got:\n%s", findings(rep))
	}
}

// The real thing: this repository's own configuration must pass, or devctl is
// asserting a schema its reference instance does not satisfy.
func TestThisRepositoryIsHealthy(t *testing.T) {
	rep, err := Run(config.Root("../.."))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if rep.Errors() != 0 {
		t.Fatalf("this repository's own .agentic/ must be coherent, got:\n%s", findings(rep))
	}
}
