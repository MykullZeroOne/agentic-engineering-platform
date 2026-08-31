package work

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
)

// A store with one item per interesting state, deliberately written to disk in an order
// that is neither the flow order nor alphabetical, so a passing ordering test cannot be
// an accident of how the filesystem happened to return them.
func store(t *testing.T) config.Root {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		".agentic/project.yaml": "registries:\n  states: .agentic/registries/states.yaml\nwork_store: local\n",
		".agentic/registries/states.yaml": `
axes:
  work_state:
    normal_flow: [intake, ready, in_progress, review, done]
    values:
      - {token: intake}
      - {token: ready}
      - {token: in_progress}
      - {token: review}
      - {token: done}
      - {token: blocked}
      - {token: awaiting_human}
`,
		".agentic/work/WI-0003.yaml": item("WI-0003", "done", "feat", "normal", "Third"),
		".agentic/work/WI-0001.yaml": item("WI-0001", "awaiting_human", "spec", "high", "First"),
		".agentic/work/WI-0002.yaml": item("WI-0002", "intake", "feat", "low", "Second"),
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
	return config.Root(root)
}

func item(id, state, typ, prio, title string) string {
	return "id: " + id + "\ntype: " + typ + "\nwork_state: " + state +
		"\npriority: " + prio + "\ntitle: " + title + "\n"
}

func ids(items []config.WorkItem) []string {
	out := make([]string, len(items))
	for i, w := range items {
		out[i] = w.ID
	}
	return out
}

func TestListOrdersByNormalFlow(t *testing.T) {
	got, err := List(store(t), Filter{})
	if err != nil {
		t.Fatal(err)
	}
	// intake -> done along the flow, then awaiting_human, which the flow omits.
	want := []string{"WI-0002", "WI-0003", "WI-0001"}
	if strings.Join(ids(got), ",") != strings.Join(want, ",") {
		t.Fatalf("want %v, got %v -- a listing that is not in flow order reads as a dump", want, ids(got))
	}
}

func TestOffFlowStatesSortLast(t *testing.T) {
	got, _ := List(store(t), Filter{})
	if got[len(got)-1].WorkState != "awaiting_human" {
		t.Fatalf("awaiting_human must sort last so it is noticed, got %v", ids(got))
	}
}

func TestFilters(t *testing.T) {
	root := store(t)
	for _, tc := range []struct {
		name string
		f    Filter
		want []string
	}{
		{"state", Filter{State: "intake"}, []string{"WI-0002"}},
		{"type", Filter{Type: "spec"}, []string{"WI-0001"}},
		{"priority", Filter{Priority: "high"}, []string{"WI-0001"}},
		{"combined", Filter{Type: "feat", Priority: "low"}, []string{"WI-0002"}},
		{"no match", Filter{State: "blocked"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := List(root, tc.f)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(ids(got), ",") != strings.Join(tc.want, ",") {
				t.Errorf("want %v, got %v", tc.want, ids(got))
			}
		})
	}
}

func TestFindAndMissing(t *testing.T) {
	root := store(t)
	w, err := Find(root, "WI-0002")
	if err != nil || w.Title != "Second" {
		t.Fatalf("Find(WI-0002) = %v, %v", w, err)
	}
	if _, err := Find(root, "WI-9999"); err == nil {
		t.Fatal("a missing id must be an error, not an empty item")
	}
}

func TestFormatListMarksStatesNeedingAttention(t *testing.T) {
	got, _ := List(store(t), Filter{})
	out := FormatList(got)
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "WI-0001") && !strings.HasPrefix(line, "! ") {
			t.Errorf("awaiting_human must be marked, got %q", line)
		}
		if strings.Contains(line, "WI-0002") && strings.HasPrefix(line, "! ") {
			t.Errorf("intake must not be marked, got %q", line)
		}
	}
	if !strings.Contains(out, "3 item(s).") {
		t.Errorf("the count is what says the listing looked at everything, got:\n%s", out)
	}
}

func TestFormatListEmpty(t *testing.T) {
	if out := FormatList(nil); !strings.Contains(out, "No work items match") {
		t.Errorf("an empty listing must say so rather than print nothing, got %q", out)
	}
}

func TestFormatItemShowsGatesWithoutClaimingCoverage(t *testing.T) {
	w := &config.WorkItem{
		ID: "WI-0001", Title: "T", WorkState: "review", Type: "spec",
		RequiredGates: []string{"platform_config"},
		Description:   "line one\nline two\n",
	}
	out := FormatItem(w)
	if !strings.Contains(out, "platform_config") {
		t.Error("required gates must be shown")
	}
	// Naming a gate is not saying whether it is closed. check_gates.py answers that, and
	// a second implementation here would be the same rule in two languages.
	if !strings.Contains(out, "check_gates") {
		t.Error("must point at what actually evaluates coverage rather than imply it")
	}
	if !strings.Contains(out, "  line two") {
		t.Error("description should be indented as a block, got:\n" + out)
	}
}

// The real store must be readable, or `devctl work` is tested only against fixtures.
func TestThisRepositoryStoreLists(t *testing.T) {
	got, err := List(config.Root("../.."), Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("this repository's work store is not empty")
	}
}
