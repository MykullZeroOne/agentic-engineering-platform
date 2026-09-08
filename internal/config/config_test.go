package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// The fixtures are built in a temp directory rather than committed under testdata/.
// A committed fixture of .agentic/ would be a second copy of the schema, and this
// repository keeps finding that two copies means one of them is quietly wrong.
//
// This helper style is copied from internal/doctor/doctor_test.go rather than shared:
// the two packages do not share a test package today, and this leg is not the place
// to introduce one.

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

// engineerPrimaryYAML is the real role file content this leg ships at
// .agentic/roles/engineer.primary.yaml, duplicated here so the loader probe fixture
// and the shipped file can drift apart loudly (a test failure) rather than silently.
const engineerPrimaryYAML = `
id: engineer.primary
role: engineer
role_version: 1

function: implementation

parent: null
specialists: []

capabilities:
  - implementation
  - test-authoring

skills: []

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
    - function/engineering
    - project

completion:
  human_owned: true
  criteria:
    - local_checks_pass
    - scope_respected
    - evidence_package_complete

returns_from:
  - review
  - qa
  - human

human_gates: []
`

func TestC9_RoleDeclaresFunctionCompletionAndTools(t *testing.T) {
	root := t.TempDir()
	write(t, root, files{
		".agentic/roles/engineer.primary.yaml": engineerPrimaryYAML,
	})

	role, err := LoadRole(Root(root), "engineer.primary")
	if err != nil {
		t.Fatalf("LoadRole: %v", err)
	}

	if role.Function != "implementation" {
		t.Errorf("Function = %q, want %q", role.Function, "implementation")
	}
	if !role.Completion.HumanOwned {
		t.Errorf("Completion.HumanOwned = false, want true")
	}
	if len(role.ReturnsFrom) == 0 {
		t.Errorf("ReturnsFrom is empty, want at least one entry")
	}
	wantAllow := []string{"repository.read", "repository.write", "shell.exec"}
	if !reflect.DeepEqual(role.Tools.Allow, wantAllow) {
		t.Errorf("Tools.Allow = %v, want %v", role.Tools.Allow, wantAllow)
	}
}

func TestC9_LoadRoleNamesTheIDItCouldNotFind(t *testing.T) {
	root := t.TempDir()
	// No .agentic/roles/ at all: a project that has defined no roles.

	_, err := LoadRole(Root(root), "engineer.primary")
	if err == nil {
		t.Fatal("LoadRole: expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "engineer.primary") {
		t.Errorf("error should name the id it could not find, got: %v", err)
	}
}
