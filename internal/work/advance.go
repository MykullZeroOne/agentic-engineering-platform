package work

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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

// MergedItems maps a work item ID to the commit on ref that closed it.
//
// git, not the GitHub API. The fact is already in the repository: the squash commit body
// carries the trailer, so this needs no token, no network, and can be tested against a
// fixture repository. CLAUDE.md's "GitHub's merge record is the only authority" is about
// identifying a merged BRANCH -- squash means the branch tip is never an ancestor of main
// -- which is a different question from which items a commit closed.
func MergedItems(root, ref string) (map[string]string, error) {
	cmd := exec.Command("git", "log", ref, "--format=%H%x1f%B%x1e")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log %s: %w", ref, err)
	}

	merged := map[string]string{}
	for _, entry := range strings.Split(string(out), "\x1e") {
		sha, body, ok := strings.Cut(strings.TrimSpace(entry), "\x1f")
		if !ok {
			continue
		}
		for _, m := range closes.FindAllStringSubmatch(body, -1) {
			// First writer wins: git log is newest-first, and if an ID somehow appears
			// twice the later merge is the one that closed it.
			if _, seen := merged[m[1]]; !seen {
				merged[m[1]] = sha[:min(12, len(sha))]
			}
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
func Plan(root string, items []config.WorkItem, merged map[string]string, records []approvals.Record) []Change {
	var changes []Change
	for _, item := range items {
		commit, wasMerged := merged[item.ID]
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
			changes = append(changes, Change{item.ID, item.WorkState, target, commit, why})
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
		fmt.Fprintf(&b, "  %s  %s -> %s   (closed by %s)\n", ch.ID, ch.From, ch.To, ch.Commit)
		for _, why := range ch.Why {
			fmt.Fprintf(&b, "      %s\n", why)
		}
	}
	return b.String()
}
