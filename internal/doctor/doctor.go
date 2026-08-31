// Package doctor answers one question: is this project's .agentic/ configuration
// coherent with itself.
//
// Deliberately not the same question as scripts/validate_docs.py, which validates the
// documentation corpus -- front matter, tiers, supersession, the index. The split is a
// real distinction rather than a staging accident: a project could have perfect
// configuration and a broken corpus, or the reverse, and conflating them would mean
// neither answer is trustworthy on its own.
//
// Read-only. No network, no writes, no state. A doctor that repairs is a doctor whose
// report you cannot trust, because you can no longer tell what was already true.
package doctor

import (
	"fmt"
	"os"
	"sort"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
)

// Severity distinguishes what breaks a runtime from what merely reads badly.
type Severity string

const (
	// Error means a runtime reading this configuration would fail or act wrongly.
	Error Severity = "error"
	// Warning means the configuration is usable but something is off.
	Warning Severity = "warning"
)

// Finding is one thing wrong, named where it is wrong.
type Finding struct {
	Severity Severity
	Where    string // the file or key, so the reader knows what to open
	Message  string
}

func (f Finding) String() string { return fmt.Sprintf("%s: %s: %s", f.Severity, f.Where, f.Message) }

// Report is the outcome of a run.
type Report struct {
	Findings []Finding
	Checked  map[string]int // what was examined, so a clean report is not silent
}

// Errors reports whether anything found would break a runtime.
func (r *Report) Errors() int {
	n := 0
	for _, f := range r.Findings {
		if f.Severity == Error {
			n++
		}
	}
	return n
}

type run struct {
	root config.Root
	rep  *Report
}

func (r *run) err(where, format string, a ...any) {
	r.rep.Findings = append(r.rep.Findings, Finding{Error, where, fmt.Sprintf(format, a...)})
}

func (r *run) warn(where, format string, a ...any) {
	r.rep.Findings = append(r.rep.Findings, Finding{Warning, where, fmt.Sprintf(format, a...)})
}

// Run examines the project rooted at root.
//
// It stops early only when it cannot read project.yaml, because every later check
// resolves paths that file names. Everything else accumulates: a report listing one
// problem when there are six wastes a round trip.
func Run(root config.Root) (*Report, error) {
	r := &run{root: root, rep: &Report{Checked: map[string]int{}}}

	proj, err := config.LoadProject(root)
	if err != nil {
		if os.IsNotExist(err) {
			r.err(".agentic/project.yaml", "not found; this is not an AEP project root")
			return r.rep, nil
		}
		return nil, err
	}

	gates := r.checkRegistries(proj)
	r.checkHumanGates(proj, gates)
	r.checkWorkStore(proj)
	return r.rep, nil
}

// checkRegistries verifies every registry project.yaml points at exists and parses,
// and returns the gate registry if it loaded, since later checks resolve against it.
func (r *run) checkRegistries(proj *config.Project) *config.Gates {
	const where = ".agentic/project.yaml"
	if len(proj.Registries) == 0 {
		r.err(where, "declares no registries; gate, state and vocabulary tokens cannot be resolved")
		return nil
	}

	names := make([]string, 0, len(proj.Registries))
	for n := range proj.Registries {
		names = append(names, n)
	}
	sort.Strings(names)

	var gates *config.Gates
	for _, name := range names {
		rel := proj.Registries[name]
		if _, err := os.Stat(r.root.Path(rel)); err != nil {
			r.err(where, "registry %q points at %s, which does not exist", name, rel)
			continue
		}
		r.rep.Checked["registries"]++

		// Parse the ones devctl understands. A registry it does not know is counted as
		// present but not read, rather than assumed fine -- silence is not verification.
		var perr error
		switch name {
		case "gates":
			gates, perr = config.LoadGates(r.root, rel)
		case "vocabularies":
			_, perr = config.LoadVocabularies(r.root, rel)
		case "states":
			_, perr = config.LoadStates(r.root, rel)
		}
		if perr != nil {
			r.err(rel, "%v", perr)
		}
	}

	if gates != nil {
		r.checkGateRegistry(proj, gates)
	}
	return gates
}

// checkGateRegistry validates the gates themselves. risk_tier is required by ADR-014
// clause 7: a gate without one is indistinguishable from a reversible one, and treating
// every gate alike is what the clause exists to prevent.
func (r *run) checkGateRegistry(proj *config.Project, gates *config.Gates) {
	rel := proj.Registries["gates"]
	vocRel, hasVoc := proj.Registries["vocabularies"]
	tiers, scopes := map[string]bool{}, map[string]bool{}
	if hasVoc {
		if v, err := config.LoadVocabularies(r.root, vocRel); err == nil {
			if t, ok := v.Tokens("risk_tier"); ok {
				tiers = t
			}
			if s, ok := v.Tokens("gate_scope"); ok {
				scopes = s
			}
		}
	}

	seen := map[string]bool{}
	for _, g := range gates.Gates {
		r.rep.Checked["gates"]++
		if g.ID == "" {
			r.err(rel, "a gate has no id")
			continue
		}
		if seen[g.ID] {
			r.err(rel, "duplicate gate id %q", g.ID)
		}
		seen[g.ID] = true

		if len(tiers) > 0 {
			if g.RiskTier == "" {
				r.err(rel, "gate %q has no risk_tier (ADR-014 clause 7)", g.ID)
			} else if !tiers[g.RiskTier] {
				r.err(rel, "gate %q risk_tier %q is not in the risk_tier vocabulary", g.ID, g.RiskTier)
			}
		}
		// applies_when decides whether a path trigger fires on a draft artifact (WI-0018).
		// An unrecognised value would silently widen or narrow the gate.
		if len(scopes) > 0 && g.AppliesWhen != "" && !scopes[g.AppliesWhen] {
			r.err(rel, "gate %q applies_when %q is not in the gate_scope vocabulary", g.ID, g.AppliesWhen)
		}
	}
}

// checkHumanGates verifies every gate the project puts in force actually exists.
// A human_gates entry naming nothing is a gate that silently never fires.
func (r *run) checkHumanGates(proj *config.Project, gates *config.Gates) {
	if gates == nil {
		return
	}
	known := map[string]bool{}
	for _, g := range gates.Gates {
		known[g.ID] = true
	}
	for _, id := range proj.HumanGates {
		r.rep.Checked["human_gates"]++
		if !known[id] {
			r.err(".agentic/project.yaml", "human_gates names unknown gate %q", id)
		}
	}
	if len(proj.HumanGates) == 0 {
		r.warn(".agentic/project.yaml", "no human gates are in force")
	}
}

// checkWorkStore validates the store named by work_store. Only `local` is readable
// here; `github` lives behind an adapter that does not exist yet, and reporting it as
// broken would be wrong -- it is unimplemented, which is a different thing.
func (r *run) checkWorkStore(proj *config.Project) {
	const where = ".agentic/project.yaml"
	switch proj.WorkStore {
	case "":
		r.err(where, "work_store is unset; ADR-012 requires exactly one authoritative store")
		return
	case "local":
	case "github":
		r.warn(where, "work_store is %q, which devctl cannot read yet", proj.WorkStore)
		return
	default:
		r.err(where, "work_store %q is not a known store", proj.WorkStore)
		return
	}

	items, err := config.LoadWorkItems(r.root)
	if err != nil {
		r.err(".agentic/work/", "%v", err)
		return
	}

	types, states, prios := r.workVocabularies(proj)
	ids := map[string]bool{}
	for _, w := range items {
		ids[w.ID] = true
	}

	for _, w := range items {
		r.rep.Checked["work_items"]++
		at := ".agentic/work/" + w.ID + ".yaml"
		if w.ID == "" {
			r.err(".agentic/work/", "a work item has no id")
			continue
		}
		check := func(field, val string, set map[string]bool) {
			if len(set) == 0 {
				return
			}
			if !set[val] {
				r.err(at, "%s %q is not a known token", field, val)
			}
		}
		check("type", w.Type, types)
		check("work_state", w.WorkState, states)
		check("priority", w.Priority, prios)

		// A dependency naming nothing is a decorative edge: it reads as a constraint
		// and constrains nothing, which is worse than having no edge at all.
		for _, dep := range w.Dependencies {
			if !ids[dep] {
				r.err(at, "depends on unknown work item %q", dep)
			}
		}
	}
}

func (r *run) workVocabularies(proj *config.Project) (types, states, prios map[string]bool) {
	if rel, ok := proj.Registries["vocabularies"]; ok {
		if v, err := config.LoadVocabularies(r.root, rel); err == nil {
			types, _ = v.Tokens("work_item_type")
			prios, _ = v.Tokens("priority")
		}
	}
	if rel, ok := proj.Registries["states"]; ok {
		if s, err := config.LoadStates(r.root, rel); err == nil {
			states, _ = s.Tokens("work_state")
		}
	}
	return
}
