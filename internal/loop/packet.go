package loop

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
	"gopkg.in/yaml.v3"
)

// Packet is the template data for the work packet handed to a runtime session.
type Packet struct {
	AgentIdentity   string
	RoleTitle       string
	Project         string
	Objective       string
	ItemID          string
	ItemTitle       string
	ItemType        string
	ItemState       string
	ItemDescription string
	Worktree        string
	Branch          string
	// FailedChecks is non-empty on a re-entry: the session is told what failed rather
	// than left to rediscover it.
	FailedChecks []Check
	// CheckOutput maps a check name to its captured output, for FailedChecks.
	CheckOutput map[string]string
}

const packetTemplateText = `You are {{.AgentIdentity}}, the {{.RoleTitle}} on {{.Project}}.

# Objective

{{.Objective}}

# Work item

{{.ItemID}} — {{.ItemTitle}}
type: {{.ItemType}}    work_state: {{.ItemState}}

{{.ItemDescription}}

# Where you are

A git worktree at {{.Worktree}}, on branch {{.Branch}}, checked out from this
repository's main. Read CLAUDE.md before you change anything.

# Instructions

1. Implement only {{.ItemID}}. Anything else you notice is a finding for your closing
   summary, never a change to make.
2. Run this repository's own checks before you finish:
     go build ./... && go test ./...     (whenever you changed cmd/ or internal/)
     python3 scripts/build_index.py      (whenever you added or removed a document)
     python3 scripts/validate_docs.py
3. Do not commit. Do not push. Do not open, review, or merge a pull request. Do not run
   git commit, git push, or gh. The loop that started you does all of that, and a commit
   you make is one it did not expect.
4. Do not write under .agentic/work/, .agentic/approvals/, or .agentic/registries/.
5. If a judgment call comes up — something this packet does not decide, a requirement you
   would have to invent, a check you would have to weaken, an approval you would have to
   infer — stop. Emit one line beginning "QUESTION: " followed by the question, then end
   your turn. Do not guess, and do not proceed with a guess flagged as a guess.
6. End with a short summary: what you changed, which files, and what you deliberately
   did not do.
{{if .FailedChecks}}
# The previous attempt failed these checks
{{range .FailedChecks}}
{{.Name}} — {{.Command}} — rc {{.RC}}
{{index $.CheckOutput .Name}}
{{end}}
Fix them. The instructions above still hold.
{{end}}`

var packetTemplate = template.Must(template.New("packet").Parse(packetTemplateText))

// RenderPacket renders the prompt a runtime session is started with.
func RenderPacket(p Packet) (string, error) {
	var b strings.Builder
	if err := packetTemplate.Execute(&b, p); err != nil {
		return "", err
	}
	return b.String(), nil
}

// RenderPRBody renders the pull request body: the evidence/v1 package, the stubbed
// list, and the run record itself. ADR-017 makes the body a projection of the
// package; .agentic/runs/ is gitignored, so this projection is also the record's
// durable copy.
func RenderPRBody(rec *Record, w config.WorkItem) (string, error) {
	b, err := yaml.Marshal(rec)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "# %s — %s\n\n", w.ID, w.Title)
	fmt.Fprintf(&buf, "Evidence package: evidence/v1\n\n")

	if len(rec.Stubbed) > 0 {
		buf.WriteString("## Stubbed / out of scope\n\n")
		for _, s := range rec.Stubbed {
			fmt.Fprintf(&buf, "- %s\n", s)
		}
		buf.WriteString("\n")
	}

	fmt.Fprintf(&buf, "## Run record\n\n```yaml\n%s```\n", string(b))
	return buf.String(), nil
}
