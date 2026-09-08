package loop

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// evidencePackage is docs/spec/EVIDENCE_PACKAGE.md's evidence/v1 schema: identity,
// claims, proofs, and gates as the four required blocks, plus open, required to be
// present even when empty. RenderPRBody renders it as the pull request body's
// leading fenced block, per ADR-017 (the body is a projection of the package).
type evidencePackage struct {
	Identity evidenceIdentity `yaml:"identity"`
	Claims   []evidenceClaim  `yaml:"claims"`
	Proofs   []evidenceProof  `yaml:"proofs"`
	Gates    []evidenceGate   `yaml:"gates"`
	Open     []string         `yaml:"open"`
}

type evidenceIdentity struct {
	EvidenceID string `yaml:"evidence_id"`
	Schema     string `yaml:"schema"`
	WorkItem   string `yaml:"work_item"`
	HeadCommit string `yaml:"head_commit"`
	RecordedOn string `yaml:"recorded_on"`
}

// evidenceClaim is one material change. Rule 4 (one proof per claim) is trivially
// true here: the loop has exactly one primary proof, the local-check run, so every
// claim cites it as P-checks.
type evidenceClaim struct {
	Change    string `yaml:"change"`
	Satisfies string `yaml:"satisfies"`
	Proof     string `yaml:"proof"`
}

// evidenceProof is one automated proof. The loop never produces a manual proof, so
// Kind is always "automated".
type evidenceProof struct {
	ID      string `yaml:"id"`
	Kind    string `yaml:"kind"`
	Command string `yaml:"command"`
	Expect  string `yaml:"expect"`
	RanOn   string `yaml:"ran_on"`
	Result  string `yaml:"result"`
}

// evidenceGate is one human gate the check-gates check reported. ApprovalRecord is
// nil, not "", when the line names none: ADR-019 makes `state: crossed` with a null
// approval_record a valid, blocking entry rather than an omission.
type evidenceGate struct {
	ID             string  `yaml:"id"`
	State          string  `yaml:"state"`
	ApprovalRecord *string `yaml:"approval_record"`
}

// gateLineRE matches one gate line from scripts/check_gates.py's non-quiet output:
//
//	"  <gate id>  [OPEN|closed]  approver: <name>"
//
// Deliberately tolerant: SPEC-EVIDENCE-PACKAGE says nothing about a schema for that
// output, so this keys only on a token immediately before a bracketed OPEN/closed and
// does not otherwise constrain the line.
var gateLineRE = regexp.MustCompile(`(?m)^\s*(\w+)\s+\[(OPEN|closed)\].*$`)

// approvalRecordRE finds an approval record id named anywhere on a matched gate line.
var approvalRecordRE = regexp.MustCompile(`APR-\d+`)

// suspendedReviewNote records POL-001 M4 (independent review) as open on every
// package, per docs/policies/POL-001-BRANCHING-AND-MERGING.md: one human identity
// means GitHub cannot tell author from reviewer, so the requirement is suspended
// rather than met and must never read as satisfied.
const suspendedReviewNote = "independent review (POL-001 M4) suspended repository-wide"

// buildEvidencePackage assembles the current pass's evidence/v1 package from the
// record. It is best-effort about the check-gates check's output: a run whose checks
// never included one (every fixture in this package, and any first mile that has not
// yet wired check-gates into its own checker) still renders a package, with
// gates: [].
func buildEvidencePackage(rec *Record, w config.WorkItem, headCommit string, root config.Root) evidencePackage {
	head := headCommit
	if rec.Handoff != nil && rec.Handoff.HeadCommit != "" {
		head = rec.Handoff.HeadCommit
	}

	pkg := evidencePackage{
		Identity: evidenceIdentity{
			EvidenceID: fmt.Sprintf("EV-%s-pass-%d", rec.RunID, rec.Pass),
			Schema:     "evidence/v1",
			WorkItem:   w.ID,
			HeadCommit: head,
			RecordedOn: rec.Updated,
		},
		Claims: []evidenceClaim{},
		Proofs: []evidenceProof{},
		Gates:  []evidenceGate{},
		Open:   append([]string(nil), rec.Stubbed...),
	}

	if step5, ok := rec.Step(rec.Pass, 5); ok {
		for _, f := range step5.FilesChanged {
			pkg.Claims = append(pkg.Claims, evidenceClaim{
				Change:    fmt.Sprintf("%s changed", f),
				Satisfies: w.ID,
				Proof:     "P-checks",
			})
		}
	}

	if step7, ok := rec.Step(rec.Pass, 7); ok {
		allPass := true
		for _, c := range step7.Checks {
			result := "pass"
			if c.RC != 0 {
				result = "fail"
				allPass = false
			}
			pkg.Proofs = append(pkg.Proofs, evidenceProof{
				ID:      "P-" + c.Name,
				Kind:    "automated",
				Command: c.Command,
				Expect:  "exit 0",
				RanOn:   head,
				Result:  result,
			})
			if c.Name == "check-gates" {
				pkg.Gates = append(pkg.Gates, parseGateLines(readCheckOutput(root, rec.RunID, c.Output))...)
			}
		}
		aggResult := "pass"
		if !allPass {
			aggResult = "fail"
		}
		pkg.Proofs = append(pkg.Proofs, evidenceProof{
			ID:      "P-checks",
			Kind:    "automated",
			Command: "all local checks",
			Expect:  "every rc 0",
			RanOn:   head,
			Result:  aggResult,
		})
	}

	if !anyContains(pkg.Open, "POL-001 M4") {
		pkg.Open = append(pkg.Open, suspendedReviewNote)
	}

	return pkg
}

// readCheckOutput returns a check's captured output, or "" if it cannot be read.
// rel is relative to the run directory (Check.Output's documented contract).
func readCheckOutput(root config.Root, runID, rel string) string {
	if rel == "" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(RunDir(root, runID), rel))
	if err != nil {
		return ""
	}
	return string(b)
}

func parseGateLines(output string) []evidenceGate {
	var gates []evidenceGate
	for _, m := range gateLineRE.FindAllStringSubmatch(output, -1) {
		line, id := m[0], m[1]
		var approvalRecord *string
		if apr := approvalRecordRE.FindString(line); apr != "" {
			approvalRecord = &apr
		}
		gates = append(gates, evidenceGate{ID: id, State: "crossed", ApprovalRecord: approvalRecord})
	}
	return gates
}

// anyContains reports whether any entry in ss contains substr. Used to decide
// whether the POL-001 M4 note is "already in it" (buildEvidencePackage's `open`):
// rec.Stubbed's own wording of that note ("independent review (POL-001 M4): ...")
// differs from suspendedReviewNote's, so an exact-match check would append a second,
// redundant entry every time.
func anyContains(ss []string, substr string) bool {
	for _, x := range ss {
		if strings.Contains(x, substr) {
			return true
		}
	}
	return false
}

// RenderPRBody renders the pull request body: the evidence/v1 package, the stubbed
// list, and the run record itself. ADR-017 makes the body a projection of the
// package; .agentic/runs/ is gitignored, so this projection is also the record's
// durable copy. headCommit is the sha control.Commit just produced; rec.Handoff is
// not set yet at the point step 7 calls this (Handoff.HeadCommit takes precedence
// once it exists, e.g. if this is ever called again after handoff). root locates the
// run directory so the check-gates check's captured output can be read back.
func RenderPRBody(rec *Record, w config.WorkItem, headCommit string, root config.Root) (string, error) {
	pkg := buildEvidencePackage(rec, w, headCommit, root)
	pkgYAML, err := yaml.Marshal(pkg)
	if err != nil {
		return "", err
	}

	b, err := yaml.Marshal(rec)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "# %s — %s\n\n", w.ID, w.Title)
	fmt.Fprintf(&buf, "Evidence package: evidence/v1\n\n")
	fmt.Fprintf(&buf, "```yaml\n%s```\n\n", string(pkgYAML))

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
