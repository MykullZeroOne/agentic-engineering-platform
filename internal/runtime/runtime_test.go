package runtime

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// TestC12_InterfaceFileNamesNoProvider guards runtime.go against ever learning which
// provider answered: this package's public interface must not name one.
func TestC12_InterfaceFileNamesNoProvider(t *testing.T) {
	body, err := os.ReadFile("runtime.go")
	if err != nil {
		t.Fatalf("reading runtime.go: %v", err)
	}
	lower := strings.ToLower(string(body))
	for _, needle := range []string{"claude", "codex"} {
		if strings.Contains(lower, needle) {
			t.Errorf("runtime.go must not mention %q, but does", needle)
		}
	}
}

// TestC16_RuntimeDoesNotImportLoop guards the dependency direction: internal/runtime
// must never import internal/loop.
func TestC16_RuntimeDoesNotImportLoop(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "./...").Output()
	if err != nil {
		t.Fatalf("go list -deps ./...: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "internal/loop" || strings.HasSuffix(line, "/internal/loop") {
			t.Fatalf("internal/runtime must not depend on internal/loop, found %q", line)
		}
	}
}

// TestC16_AdaptersAreSelectedByProviderToken asserts adding a second provider costs
// one map entry: New resolves a token to an adapter, an unknown token wraps
// ErrNoAdapter and names itself, and Providers lists what is registered.
func TestC16_AdaptersAreSelectedByProviderToken(t *testing.T) {
	a, err := New("claude-subscription")
	if err != nil {
		t.Fatalf("New(claude-subscription): unexpected error %v", err)
	}
	if a == nil {
		t.Fatal("New(claude-subscription): got nil Adapter, want non-nil")
	}

	codex, err := New("codex-subscription")
	if err != nil {
		t.Fatalf("New(codex-subscription): unexpected error %v", err)
	}
	if codex == nil {
		t.Fatal("New(codex-subscription): got nil Adapter, want non-nil")
	}

	_, err = New("unknown-subscription")
	if err == nil {
		t.Fatal("New(unknown-subscription): want error, got nil")
	}
	if !errors.Is(err, ErrNoAdapter) {
		t.Errorf("New(unknown-subscription) error = %v; want errors.Is(err, ErrNoAdapter)", err)
	}
	if !strings.Contains(err.Error(), "unknown-subscription") {
		t.Errorf("New(unknown-subscription) error = %v; want it to name the provider", err)
	}

	got := Providers()
	want := []string{"claude-subscription", "codex-subscription"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Providers() = %v, want %v", got, want)
	}
}
