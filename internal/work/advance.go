package work

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/approvals"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
)

// closes matches the trailer that carries a work item onto the trunk.
//
// CLAUDE.md requires `Closes WI-NNNN` in the BRANCH commit message rather than only in
// the pull request body, because GitHub's squash body defaults to the branch's commits.
// PRs #8 and #9 both merged without it, which is why this returns "no merge record" for
// their items instead of inventing one.
var closes = regexp.MustCompile(`(?m)\bCloses\s+(WI-\d{4})\b`)

// pullRequest matches the number GitHub appends to a squash commit subject.
var pullRequest = regexp.MustCompile(`\(#(\d+)\)\s*$`)

// Merge is how a work item was found to have merged.
type Merge struct {
	Commit string
	Source string // "trailer" or "pull-request"
}

// PRResolver maps pull request numbers to the branch each merged from.
//
// An interface so the fallback is testable without a network. The production implementation
// shells out to `gh`, which is already authenticated; nothing here handles a token.
type PRResolver interface {
	MergedBranches() (map[int]string, error)
}

// GH resolves through the gh CLI, in one call rather than one per pull request.
type GH struct{ Repo string }

func (g GH) MergedBranches() (map[int]string, error) {
	args := []string{"pr", "list", "--state", "merged", "--limit", "200",
		"--json", "number,headRefName"}
	if g.Repo != "" {
		args = append(args, "--repo", g.Repo)
	}
	out, err := exec.Command("gh", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}
	var rows []struct {
		Number      int    `json:"number"`
		HeadRefName string `json:"headRefName"`
	}
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}
	branches := make(map[int]string, len(rows))
	for _, r := range rows {
		branches[r.Number] = r.HeadRefName
	}
	return branches, nil
}

// MergedItems maps a work item ID to the merge that closed it, from two sources.
//
// The `Closes WI-NNNN` trailer is preferred: it needs no token, no network, and is testable
// against a fixture repository. It is also unreliable in practice. Of twenty merges to this
// trunk, eight arrived with no body at all -- six of them consecutively -- even though every
// branch commit carried the trailer and the repository's squash setting is COMMIT_MESSAGES.
// The convention CLAUDE.md hardened after PRs #8 and #9 has now failed eight times.
//
// So when a commit carries no trailer, the pull request number GitHub appends to every squash
// subject is used instead: resolve it to the branch that merged, and match that against the
// work item's own `branch` field. That path needs the network and is therefore second, not
// first -- but it works on every merge, because GitHub writes the number itself rather than
// relying on anyone to remember a convention.
//
// resolver may be nil, which disables the fallback and restores the previous behaviour.
func MergedItems(root, ref string, items []config.WorkItem, resolver PRResolver) (map[string]Merge, error) {
	cmd := exec.Command("git", "log", ref, "--format=%H%x1f%B%x1e")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log %s: %w", ref, err)
	}

	type commit struct{ sha, subject, body string }
	var commits []commit
	merged := map[string]Merge{}

	for _, entry := range strings.Split(string(out), "\x1e") {
		sha, body, ok := strings.Cut(strings.TrimSpace(entry), "\x1f")
		if !ok {
			continue
		}
		short := sha[:min(12, len(sha))]
		subject, _, _ := strings.Cut(body, "\n")
		found := false
		for _, m := range closes.FindAllStringSubmatch(body, -1) {
			found = true
			// First writer wins: git log is newest-first, and if an ID somehow appears
			// twice the later merge is the one that closed it.
			if _, seen := merged[m[1]]; !seen {
				merged[m[1]] = Merge{short, "trailer"}
			}
		}
		if !found {
			commits = append(commits, commit{short, subject, body})
		}
	}

	if resolver == nil || len(commits) == 0 {
		return merged, nil
	}

	branches, err := resolver.MergedBranches()
	if err != nil {
		// A resolver failure degrades to the trailer-only answer rather than failing the
		// run. The caller already reports which items were not evaluated, so the loss is
		// visible -- and an offline machine should still be able to see obvious drift.
		return merged, nil
	}

	byBranch := make(map[string]string, len(items))
	for _, it := range items {
		if it.Branch != "" {
			byBranch[it.Branch] = it.ID
		}
	}
	for _, c := range commits {
		m := pullRequest.FindStringSubmatch(c.subject)
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		id, ok := byBranch[branches[n]]
		if !ok {
			continue
		}
		if _, seen := merged[id]; !seen {
			merged[id] = Merge{c.sha, "pull-request"}
		}
	}
	return merged, nil
}

// Change is one work item whose recorded state disagrees with the merge record.
type Change struct {
	ID     string
	From   string
	To     string
	Commit string   // the squash commit that closed it
	Source string   // how it was found: "trailer" or "pull-request"
	Why    []string // per-gate coverage, in the order required_gates lists them
}

// terminal states are never advanced out of. `cancelled` means withdrawn; a stray trailer
// naming a cancelled item must not resurrect it, and `done` is already the destination.
var terminal = map[string]bool{"done": true, "cancelled": true}

// Plan derives the state each merged work item should be in, and returns only the items
// whose recorded state disagrees.
//
// The rule is work.advance_state's own, from .agentic/hooks/hooks.yaml: `done` when every
// entry in required_gates is covered by an approval record, `awaiting_human` when any is
// not. Coverage is verified by hash, so a record that approved an earlier version of its
// artifact does not advance anything -- which is the point of binding an approval to a
// content hash in the first place (ADR-019).
func Plan(root string, items []config.WorkItem, merged map[string]Merge, records []approvals.Record) []Change {
	var changes []Change
	for _, item := range items {
		m, wasMerged := merged[item.ID]
		if !wasMerged || terminal[item.WorkState] {
			continue
		}

		target := "done"
		var why []string
		for _, gate := range item.RequiredGates {
			res := approvals.Cover(root, gate, records)
			if res.Coverage != approvals.Covered {
				target = "awaiting_human"
			}
			why = append(why, fmt.Sprintf("%s: %s (%s)", gate, res.Coverage, res.Detail))
		}
		if len(item.RequiredGates) == 0 {
			why = append(why, "no required gates")
		}

		if item.WorkState != target {
			changes = append(changes, Change{item.ID, item.WorkState, target, m.Commit, m.Source, why})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].ID < changes[j].ID })
	return changes
}

// stateLine matches the work_state key at the top level of a work item file.
var stateLine = regexp.MustCompile(`^work_state:\s*\S+\s*$`)

// Apply writes each change to its work item file.
//
// A targeted line rewrite, never a YAML round-trip. Re-marshalling would drop every
// comment and reflow the `description: |` block, turning a one-token state change into a
// whole-file diff that no reviewer can read -- and the descriptions in this store carry
// the reasoning that makes the items worth having.
func Apply(root string, changes []Change) error {
	for _, ch := range changes {
		path := filepath.Join(root, ".agentic", "work", ch.ID+".yaml")
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var out bytes.Buffer
		replaced := 0
		scanner := bufio.NewScanner(bytes.NewReader(raw))
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if stateLine.MatchString(line) {
				line = "work_state: " + ch.To
				replaced++
			}
			out.WriteString(line)
			out.WriteByte('\n')
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("%s: %w", ch.ID, err)
		}
		if replaced != 1 {
			// Zero means the key is missing or indented; more than one means the file has
			// a shape this rewrite does not understand. Either way, guessing which line to
			// change would corrupt the item silently.
			return fmt.Errorf("%s: found %d top-level work_state lines, want exactly 1", ch.ID, replaced)
		}
		if err := os.WriteFile(path, out.Bytes(), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// FormatPlan renders changes for a terminal.
func FormatPlan(changes []Change, applied bool) string {
	if len(changes) == 0 {
		return "Every merged work item already carries the state the merge record implies.\n"
	}
	var b strings.Builder
	verb := "would change"
	if applied {
		verb = "changed"
	}
	fmt.Fprintf(&b, "%d work item(s) %s:\n\n", len(changes), verb)
	for _, ch := range changes {
		via := ""
		if ch.Source == "pull-request" {
			// Named rather than silent: this item was found only because GitHub writes the
			// pull request number itself. Its commit carries no Closes trailer, which is a
			// convention failure worth seeing rather than papering over.
			via = "  [via pull request; no Closes trailer]"
		}
		fmt.Fprintf(&b, "  %s  %s -> %s   (closed by %s)%s\n", ch.ID, ch.From, ch.To, ch.Commit, via)
		for _, why := range ch.Why {
			fmt.Fprintf(&b, "      %s\n", why)
		}
	}
	return b.String()
}
