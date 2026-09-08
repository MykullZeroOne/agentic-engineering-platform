package runtime

import (
	"context"
	"errors"
	"sort"
)

// claudeBinary is the executable this adapter resolves on PATH.
const claudeBinary = "claude"

// permissionMode is passed to the CLI so a session started with no terminal can
// actually edit files in its worktree.
const permissionMode = "acceptEdits"

// claudeTools maps a role's capability tokens to this CLI's tool names. The mapping
// lives here rather than in the role file or the loop, because it is a fact about one
// CLI and nothing above this package may learn a CLI flag.
var claudeTools = map[string][]string{
	"repository.read":   {"Read", "Glob", "Grep"},
	"repository.write":  {"Edit", "Write"},
	"shell.exec":        {"Bash"},
	"knowledge.search":  {"Grep", "Glob"},
	"knowledge.context": {"Read"},
}

// claudeDenied maps a role's tools.deny tokens to this CLI's disallow patterns.
var claudeDenied = map[string][]string{
	"github.merge": {"Bash(gh pr merge*)", "Bash(gh api *)"},
}

// mapTools translates role capability tokens into CLI tool names, deduplicated and
// sorted. Tokens with no mapping come back in unknown rather than being dropped: a
// silently ignored capability is a permission nobody granted and nobody noticed.
func mapTools(tokens []string) (names, unknown []string) {
	return mapTokens(tokens, claudeTools)
}

// mapDeny translates role deny tokens into CLI disallow patterns, the same way
// mapTools translates allow tokens: deduplicated, sorted, and an unmappable token is
// never silently dropped.
func mapDeny(tokens []string) (names, unknown []string) {
	return mapTokens(tokens, claudeDenied)
}

// mapTokens is the shared engine behind mapTools and mapDeny: look each token up in
// table, collect the mapped CLI names deduplicated and sorted, and collect every
// token with no entry in unknown rather than dropping it silently.
func mapTokens(tokens []string, table map[string][]string) (names, unknown []string) {
	seen := map[string]bool{}
	for _, tok := range tokens {
		mapped, ok := table[tok]
		if !ok {
			unknown = append(unknown, tok)
			continue
		}
		for _, m := range mapped {
			if !seen[m] {
				seen[m] = true
				names = append(names, m)
			}
		}
	}
	sort.Strings(names)
	return names, unknown
}

// claudeAdapter runs the `claude` CLI in print mode (ADR-008: a subscription CLI,
// shelled out to; no API client and no token handling anywhere in this file).
type claudeAdapter struct {
	// binary overrides PATH resolution. Empty means look up claudeBinary on PATH.
	binary string
}

func (a *claudeAdapter) Start(ctx context.Context, p WorkPacket) (Session, error) {
	return nil, errors.New("not implemented")
}

// claudeSession is one `claude -p` process.
type claudeSession struct{}

func (s *claudeSession) ID() string            { return "" }
func (s *claudeSession) Status() Status        { return "" }
func (s *claudeSession) Events() <-chan Event  { return nil }
func (s *claudeSession) Wait() (Result, error) { return Result{}, errors.New("not implemented") }
