// Package config reads a project's .agentic/ configuration.
//
// This is the first code to treat .agentic/ as a schema rather than as this
// repository's private files. CLAUDE.md calls it "the reference instance of the
// schema the first runtime must read unchanged", and nothing had ever tested that
// claim: the Python validators were written against this repository and import its
// layout directly. Reading it from a separate module is what makes the claim falsifiable.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Project is .agentic/project.yaml. Only the fields devctl acts on are modelled;
// unknown keys are ignored rather than rejected, because a project may legitimately
// carry configuration this version of devctl predates.
type Project struct {
	Registries map[string]string `yaml:"registries"`
	WorkStore  string            `yaml:"work_store"`
	HumanGates []string          `yaml:"human_gates"`
}

// Gate is one entry in the gate registry.
type Gate struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Blocking    bool   `yaml:"blocking"`
	RiskTier    string `yaml:"risk_tier"`
	Enforcement string `yaml:"enforcement"`
}

// Gates is the gate registry file.
type Gates struct {
	Gates []Gate `yaml:"gates"`
}

// Vocabularies is the controlled-vocabulary registry. Terms carry more fields than
// this, but membership is all devctl needs to check a token against a closed set.
type Vocabularies struct {
	Vocabularies map[string]struct {
		Terms []struct {
			Token string `yaml:"token"`
		} `yaml:"terms"`
	} `yaml:"vocabularies"`
}

// States is the state registry: named axes, each a closed set of tokens.
type States struct {
	Axes map[string]struct {
		Values []struct {
			Token string `yaml:"token"`
		} `yaml:"values"`
	} `yaml:"axes"`
}

// WorkItem is one file in the local work store.
type WorkItem struct {
	ID            string   `yaml:"id"`
	Type          string   `yaml:"type"`
	WorkState     string   `yaml:"work_state"`
	Priority      string   `yaml:"priority"`
	Title         string   `yaml:"title"`
	RequiredGates []string `yaml:"required_gates"`
	Dependencies  []string `yaml:"dependencies"`
}

// Root is a project directory: the one holding .agentic/.
type Root string

// Agentic returns a path inside .agentic/.
func (r Root) Agentic(parts ...string) string {
	return filepath.Join(append([]string{string(r), ".agentic"}, parts...)...)
}

// Path resolves a repository-relative path, which is how project.yaml names its
// registries -- ".agentic/registries/gates.yaml", not "registries/gates.yaml".
func (r Root) Path(rel string) string { return filepath.Join(string(r), rel) }

func read(path string, into any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(b, into); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

// LoadProject reads .agentic/project.yaml.
func LoadProject(r Root) (*Project, error) {
	var p Project
	return &p, read(r.Agentic("project.yaml"), &p)
}

// LoadGates reads the gate registry from the path project.yaml names for it,
// rather than a path devctl assumes. A project that moves its registries stays
// readable; one whose pointer is wrong produces an error naming the pointer.
func LoadGates(r Root, rel string) (*Gates, error) {
	var g Gates
	return &g, read(r.Path(rel), &g)
}

// LoadVocabularies reads the vocabulary registry.
func LoadVocabularies(r Root, rel string) (*Vocabularies, error) {
	var v Vocabularies
	return &v, read(r.Path(rel), &v)
}

// LoadStates reads the state registry.
func LoadStates(r Root, rel string) (*States, error) {
	var s States
	return &s, read(r.Path(rel), &s)
}

// Tokens returns the membership of one vocabulary, and whether it exists.
func (v *Vocabularies) Tokens(name string) (map[string]bool, bool) {
	voc, ok := v.Vocabularies[name]
	if !ok {
		return nil, false
	}
	out := make(map[string]bool, len(voc.Terms))
	for _, t := range voc.Terms {
		out[t.Token] = true
	}
	return out, true
}

// Tokens returns the membership of one state axis, and whether it exists.
func (s *States) Tokens(axis string) (map[string]bool, bool) {
	ax, ok := s.Axes[axis]
	if !ok {
		return nil, false
	}
	out := make(map[string]bool, len(ax.Values))
	for _, v := range ax.Values {
		out[v.Token] = true
	}
	return out, true
}

// LoadWorkItems reads every WI-*.yaml in the local work store, sorted by filename.
func LoadWorkItems(r Root) ([]WorkItem, error) {
	paths, err := filepath.Glob(r.Agentic("work", "WI-*.yaml"))
	if err != nil {
		return nil, err
	}
	items := make([]WorkItem, 0, len(paths))
	for _, p := range paths {
		var w WorkItem
		if err := read(p, &w); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, nil
}
