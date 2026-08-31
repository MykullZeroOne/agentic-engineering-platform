// Package work reads and presents the local work store.
//
// Read-only, like doctor. Writing work state is what `work.advance_state` is bound to do
// under `after_merge`, and it is bound to a hook engine that does not exist. A `devctl
// work set-state` would be the third mechanism for the same job -- prose asking an agent
// to remember, a human editing YAML by hand, and now a command -- when the point of the
// binding is that none of those should be needed.
//
// So this shows what is there. Every bookkeeping pull request today came from not being
// able to see the store without opening nineteen files.
package work

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
)

// Filter narrows a listing. An empty field matches everything.
type Filter struct {
	State    string
	Type     string
	Priority string
}

func (f Filter) matches(w config.WorkItem) bool {
	return (f.State == "" || w.WorkState == f.State) &&
		(f.Type == "" || w.Type == f.Type) &&
		(f.Priority == "" || w.Priority == f.Priority)
}

// List returns the items matching f, ordered by the normal flow of the work_state axis so
// the listing reads as a pipeline rather than as an alphabetical dump. States off that
// flow -- blocked, awaiting_human, failed, cancelled -- sort last, because they are the
// ones a reader needs to notice rather than scroll past.
func List(root config.Root, f Filter) ([]config.WorkItem, error) {
	items, err := config.LoadWorkItems(root)
	if err != nil {
		return nil, err
	}
	order, err := stateOrder(root)
	if err != nil {
		return nil, err
	}
	out := items[:0:0]
	for _, w := range items {
		if f.matches(w) {
			out = append(out, w)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		oi, oj := order[out[i].WorkState], order[out[j].WorkState]
		if oi != oj {
			return oi < oj
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// stateOrder ranks work states by the registry's normal_flow. Read from states.yaml
// rather than hardcoded: the flow is the registry's to define, and a second copy here
// would drift the moment a state is added.
func stateOrder(root config.Root) (map[string]int, error) {
	proj, err := config.LoadProject(root)
	if err != nil {
		return nil, err
	}
	rel, ok := proj.Registries["states"]
	if !ok {
		return map[string]int{}, nil
	}
	st, err := config.LoadStates(root, rel)
	if err != nil {
		return nil, err
	}
	order := map[string]int{}
	flow, values := st.Flow("work_state")
	for i, tok := range flow {
		order[tok] = i
	}
	// Anything the flow omits sorts after everything on it, in registry order.
	for i, tok := range values {
		if _, on := order[tok]; !on {
			order[tok] = len(flow) + i
		}
	}
	return order, nil
}

// Find returns one item by id.
func Find(root config.Root, id string) (*config.WorkItem, error) {
	items, err := config.LoadWorkItems(root)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == id {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("no work item %q in the local store", id)
}

// FormatList renders a listing as a fixed-width table.
func FormatList(items []config.WorkItem) string {
	if len(items) == 0 {
		return "No work items match.\n"
	}
	wid, wst, wty := 2, 5, 4
	for _, w := range items {
		wid = max(wid, len(w.ID))
		wst = max(wst, len(w.WorkState))
		wty = max(wty, len(w.Type))
	}
	var b strings.Builder
	for _, w := range items {
		// A marker beats a colour: this output gets pasted into pull request bodies and
		// issue comments, where an escape sequence is noise rather than emphasis.
		mark := "  "
		if w.WorkState == "blocked" || w.WorkState == "awaiting_human" || w.WorkState == "failed" {
			mark = "! "
		}
		fmt.Fprintf(&b, "%s%-*s  %-*s  %-*s  %s\n",
			mark, wid, w.ID, wst, w.WorkState, wty, w.Type, w.Title)
	}
	fmt.Fprintf(&b, "\n%d item(s).\n", len(items))
	return b.String()
}

// FormatItem renders one item in full.
func FormatItem(w *config.WorkItem) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s\n\n", w.ID, w.Title)
	row := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "  %-15s %s\n", k+":", v)
		}
	}
	row("state", w.WorkState)
	row("type", w.Type)
	row("priority", w.Priority)
	row("branch", w.Branch)
	row("parent", w.Parent)
	row("store", w.Store)
	if len(w.RequiredGates) > 0 {
		// Named, never evaluated. Whether each is closed is check_gates.py's answer, and
		// duplicating that here would put the same rule in two places in two languages.
		row("required gates", strings.Join(w.RequiredGates, ", ")+"  (coverage: see check_gates)")
	}
	if len(w.Dependencies) > 0 {
		row("depends on", strings.Join(w.Dependencies, ", "))
	}
	if d := strings.TrimRight(w.Description, "\n"); d != "" {
		b.WriteString("\n")
		for _, line := range strings.Split(d, "\n") {
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
