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
	"sort"

	"gopkg.in/yaml.v3"
)

// Project is .agentic/project.yaml. Only the fields devctl acts on are modelled;
// unknown keys are ignored rather than rejected, because a project may legitimately
// carry configuration this version of devctl predates.
type Project struct {
	Project    string            `yaml:"project"`
	Registries map[string]string `yaml:"registries"`
	WorkStore  string            `yaml:"work_store"`
	HumanGates []string          `yaml:"human_gates"`

	// RuntimePreferences maps a role function to a provider token. It is the only place
	// in a project's configuration where a provider may be named; ADR-002 forbids it in a
	// role definition, and `doctor` enforces that by comparing role values against these.
	RuntimePreferences map[string]string `yaml:"runtime_preferences"`
}

// Gate is one entry in the gate registry.
type Gate struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Blocking    bool   `yaml:"blocking"`
	RiskTier    string `yaml:"risk_tier"`
	AppliesWhen string `yaml:"applies_when"`
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

// States is the state registry: named axes, each a closed set of tokens with an
// optional normal_flow giving the order they are expected to move through.
type States struct {
	Axes map[string]struct {
		NormalFlow []string `yaml:"normal_flow"`
		Values     []struct {
			Token string `yaml:"token"`
		} `yaml:"values"`
	} `yaml:"axes"`
}

// WorkItem is one file in the local work store.
type WorkItem struct {
	ID            string   `yaml:"id"`
	Store         string   `yaml:"store"`
	Project       string   `yaml:"project"`
	Type          string   `yaml:"type"`
	WorkState     string   `yaml:"work_state"`
	Priority      string   `yaml:"priority"`
	Title         string   `yaml:"title"`
	Description   string   `yaml:"description"`
	RequiredGates []string `yaml:"required_gates"`
	Dependencies  []string `yaml:"dependencies"`
	Parent        string   `yaml:"parent"`
	Branch        string   `yaml:"branch"`
}

// Role is one file in .agentic/roles/. A durable organizational identity (ADR-002).
//
// The struct deliberately has no runtime, model, or provider field. That is not an
// omission to be corrected later: adding one would make the thing ADR-002 forbids
// representable, and doctor's check would be guarding a shape the loader invites.
type Role struct {
	ID           string         `yaml:"id"`
	Role         string         `yaml:"role"`
	RoleVersion  int            `yaml:"role_version"`
	Function     string         `yaml:"function"`
	Parent       string         `yaml:"parent"`
	Specialists  []string       `yaml:"specialists"`
	Capabilities []string       `yaml:"capabilities"`
	Skills       []string       `yaml:"skills"`
	Tools        RoleTools      `yaml:"tools"`
	Memory       RoleMemory     `yaml:"memory"`
	Completion   RoleCompletion `yaml:"completion"`
	ReturnsFrom  []string       `yaml:"returns_from"`
	HumanGates   []string       `yaml:"human_gates"`
}

// RoleTools is the capability boundary a runtime session is started inside.
type RoleTools struct {
	Allow []string `yaml:"allow"`
	Deny  []string `yaml:"deny"`
}

// RoleMemory names the role's memory namespace and what it inherits.
type RoleMemory struct {
	Namespace string   `yaml:"namespace"`
	Inherit   []string `yaml:"inherit"`
}

// RoleCompletion is the role's step-7 gate: its definition of done, and who owns it.
// HumanOwned true means the gate passes on a human's explicit approval and on nothing
// the agent concludes about completeness (ADR-023).
type RoleCompletion struct {
	HumanOwned bool     `yaml:"human_owned"`
	Criteria   []string `yaml:"criteria"`
}

// RolesDir returns .agentic/roles/.
func RolesDir(r Root) string {
	return r.Agentic("roles")
}

// RolePath returns the file a role id resolves to: .agentic/roles/<id>.yaml.
func RolePath(r Root, id string) string {
	return filepath.Join(RolesDir(r), id+".yaml")
}

// LoadRoles reads every .agentic/roles/*.yaml, sorted by filename.
func LoadRoles(r Root) ([]Role, error) {
	paths, err := filepath.Glob(filepath.Join(RolesDir(r), "*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	roles := make([]Role, 0, len(paths))
	for _, p := range paths {
		var role Role
		if err := read(p, &role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// LoadRole reads one role by id. The error names the id and the path it looked at.
func LoadRole(r Root, id string) (*Role, error) {
	path := RolePath(r, id)
	var role Role
	if err := read(path, &role); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("role %q not found at %s", id, path)
		}
		return nil, err
	}
	return &role, nil
}

// LoadRoleTree reads a role file as an untyped YAML tree.
//
// The typed Role silently discards keys it does not model, which is exactly the wrong
// behaviour for a check whose whole subject is a key that must not be there. Structural
// checks read this; everything else reads Role.
func LoadRoleTree(path string) (map[string]any, error) {
	var tree map[string]any
	if err := read(path, &tree); err != nil {
		return nil, err
	}
	return tree, nil
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

// Flow returns an axis's normal_flow and its full token list, in registry order.
func (s *States) Flow(axis string) (flow []string, values []string) {
	ax, ok := s.Axes[axis]
	if !ok {
		return nil, nil
	}
	for _, v := range ax.Values {
		values = append(values, v.Token)
	}
	return ax.NormalFlow, values
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
