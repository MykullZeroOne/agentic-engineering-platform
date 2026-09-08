package loop

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/runtime"
)

var (
	// ErrUnknownItem reports a work item the store does not hold.
	ErrUnknownItem = errors.New("no such work item")
	// ErrNoAdapter reports a resolved provider this build cannot execute.
	ErrNoAdapter = errors.New("no runtime adapter for the resolved provider")
)

// maxReentries bounds ADR-023's re-entry edge. Step 3 may be entered once plus this
// many times in one pass; the next failure parks the run blocked. The edge is what
// makes the loop recursive, and an unbounded recursion is what makes it a hang.
const maxReentries = 2

// gateID is the step-7 gate this first mile evaluates against. Named here because
// .agentic/registries/gates.yaml has no first-mile-specific engineer gate yet
// (decision 2, gate ownership): the role file says the gate is human-owned, and this
// loop records that against a stable id.
const gateID = "engineer_completion"

// defaultRoleID is used when Options.RoleID is empty.
const defaultRoleID = "engineer.primary"

// Options configures one invocation of the loop.
type Options struct {
	Root config.Root
	// Item is the work item id, WI-NNNN.
	Item string
	// Provider is the --runtime override. Empty resolves from project.yaml's
	// runtime_preferences against the role's function, and Runtime.Override records
	// which of the two happened.
	Provider string
	// RoleID defaults to "engineer.primary".
	RoleID string
	// Adapter, Control and Checker are injected by tests. Nil means the production
	// implementation: runtime.New(provider), GHControl{}, CommandChecker{DefaultChecks}.
	Adapter runtime.Adapter
	Control Control
	Checker Checker
	// Now defaults to time.Now. Injected so a record's timestamps are assertable.
	Now func() time.Time
	// Out receives the human-readable progress lines. Nil means io.Discard.
	Out io.Writer
}

// Outcome is what one invocation did.
type Outcome struct {
	RunID  string
	Record *Record
	// Exit is the process exit code devctl should use: 0 parked or complete,
	// 1 unknown work item, 2 a configuration error, 3 blocked on a question.
	Exit int
}

// Run executes ADR-023's ten-step loop for one work item, resuming an open run when
// one exists.
//
// It never merges, it writes nothing outside .agentic/runs/, and it parks at step 7 on
// a human-owned gate. Steps 8, 9 and 10 run on a later invocation, after the human has
// merged the pull request step 7 opened.
func Run(ctx context.Context, o Options) (Outcome, error) {
	now := o.Now
	if now == nil {
		now = time.Now
	}
	nowStr := func() string { return now().UTC().Format(time.RFC3339) }

	// --- Resolve, before writing anything. ---

	proj, err := config.LoadProject(o.Root)
	if err != nil {
		return Outcome{}, err
	}
	roleID := o.RoleID
	if roleID == "" {
		roleID = defaultRoleID
	}
	role, err := config.LoadRole(o.Root, roleID)
	if err != nil {
		return Outcome{}, err
	}
	items, err := config.LoadWorkItems(o.Root)
	if err != nil {
		return Outcome{}, err
	}
	var item *config.WorkItem
	for i := range items {
		if items[i].ID == o.Item {
			item = &items[i]
			break
		}
	}
	if item == nil {
		return Outcome{}, fmt.Errorf("%w: %s", ErrUnknownItem, o.Item)
	}

	provider := o.Provider
	override := true
	if provider == "" {
		provider = proj.RuntimePreferences[role.Function]
		override = false
	}

	adapter := o.Adapter
	if adapter == nil {
		a, aerr := runtime.New(provider)
		if aerr != nil {
			return Outcome{}, fmt.Errorf("%w %q; pass --runtime with a provider that has one (have: %s)",
				ErrNoAdapter, provider, strings.Join(runtime.Providers(), ", "))
		}
		adapter = a
	}

	control := o.Control
	if control == nil {
		control = GHControl{}
	}
	checker := o.Checker
	if checker == nil {
		checker = CommandChecker{Checks: DefaultChecks}
	}

	// --- Resume or open. ---

	runID, found, err := FindRun(o.Root, o.Item)
	if err != nil {
		return Outcome{}, err
	}

	var rec *Record
	startStep := 1
	var mergeCommit string

	if !found {
		newID, nerr := NewRunID(o.Root, o.Item)
		if nerr != nil {
			return Outcome{}, nerr
		}
		runID = newID
		created := nowStr()
		objective := fmt.Sprintf("Implement %s: %s", item.ID, item.Title)
		rec = &Record{
			Schema:        "run-record/v1",
			RunID:         runID,
			AgentIdentity: role.ID,
			RoleVersion:   fmt.Sprintf("%s@%d", role.Role, role.RoleVersion),
			Runtime:       Runtime{Provider: provider, Model: "provider-current", Override: override},
			Work:          Work{Item: item.ID, Title: item.Title, WorkStateAtDispatch: item.WorkState},
			State:         "working",
			Complete:      false,
			Created:       created,
			Pass:          1,
			Passes:        []Pass{{N: 1, Objective: objective, Opened: created}},
			Stubbed: []string{
				"hook engine (WI-0052): hooks.yaml bindings are declarative and none were executed",
				"work.validate_state: the run records the work state it found and enforces no dispatchability rule",
				"memory and retrieval: hydrate reads three files; consolidate writes one candidate stub",
				"Codex adapter: absent, and one entry in the adapter map away",
				"independent review (POL-001 M4): suspended repository-wide, not satisfied",
			},
		}
		if err := SaveRecord(o.Root, rec, now()); err != nil {
			return Outcome{}, err
		}
	} else {
		rec, err = LoadRecord(o.Root, runID)
		if err != nil {
			return Outcome{}, err
		}

		// C1.1: a blocked run re-invoked with no answer enters no step and starts no
		// session. Blocked is set only at step 6 (a question) or step 6/7 (the
		// re-entry limit), and in either case there is nothing to resume into short
		// of a human unblocking it — falling through to Resume() would pick up
		// wherever the record's last unexited/unentered step points, which is
		// exactly the advance this claim forbids.
		if rec.State == "blocked" {
			if o.Out != nil {
				fmt.Fprintln(o.Out, rec.Question)
			}
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			return Outcome{RunID: runID, Record: rec, Exit: 3}, nil
		}

		if rec.Handoff != nil && rec.Handoff.State == "open" {
			worktree := WorktreePath(o.Root, runID)
			pr, verr := control.ViewPR(worktree, rec.Handoff.PRURL)
			if verr != nil {
				return Outcome{}, verr
			}
			switch pr.State {
			case PROpen:
				if err := SaveRecord(o.Root, rec, now()); err != nil {
					return Outcome{}, err
				}
				return Outcome{RunID: runID, Record: rec, Exit: 0}, nil
			case PRMerged:
				rec.Handoff.State = "merged"
				mergeCommit = pr.MergeCommit
				closePassOutcome(rec, "accepted", nowStr())
				startStep = 8
			case PRClosed:
				closePassOutcome(rec, "returned", nowStr())
				if cp := rec.currentPass(); cp != nil {
					cp.PRURL = rec.Handoff.PRURL
				}
				rec.Handoff = nil
				rec.Pass++
				rec.Passes = append(rec.Passes, Pass{N: rec.Pass, Objective: "returned: " + pr.Review, Opened: nowStr()})
				startStep = 1
			default:
				return Outcome{}, fmt.Errorf("loop: unrecognized PR state %q", pr.State)
			}
		} else {
			startStep = rec.Resume()
		}
	}

	worktree := WorktreePath(o.Root, runID)
	branch := BranchName(*item)

	var (
		result       runtime.Result
		sid          string
		questions    []string
		failedChecks []Check
		checkOutput  map[string]string
	)

	step := startStep
	for {
		switch step {

		case 1: // receive
			ts := nowStr()
			rec.Steps = append(rec.Steps, Step{
				N: 1, Name: StepNames[0], Pass: rec.Pass, Entered: ts, Exited: ts,
				Note: "objective accepted; authority boundary is the role's tools.allow",
			})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 2

		case 2: // hydrate
			entered := nowStr()
			inputs, herr := hydrateInputs(o.Root, item, roleID)
			if herr != nil {
				return Outcome{}, herr
			}
			rec.Steps = append(rec.Steps, Step{
				N: 2, Name: StepNames[1], Pass: rec.Pass, Entered: entered, Exited: nowStr(), Inputs: inputs,
			})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 3

		case 3: // analyze
			entered := nowStr()
			if _, statErr := os.Stat(worktree); os.IsNotExist(statErr) {
				if werr := control.Worktree(string(o.Root), worktree, branch, "origin/main"); werr != nil {
					return Outcome{}, werr
				}
			}
			rec.Steps = append(rec.Steps, Step{
				N: 3, Name: StepNames[2], Pass: rec.Pass, Entered: entered, Exited: nowStr(),
				Note: "specialists empty; the runtime session is the sole delegate",
			})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 4

		case 4: // delegate
			entered := nowStr()
			attempt := countStepEntries(rec, rec.Pass, 4) + 1
			rec.Steps = append(rec.Steps, Step{N: 4, Name: StepNames[3], Pass: rec.Pass, Entered: entered})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}

			cp := rec.currentPass()
			objective := ""
			if cp != nil {
				objective = cp.Objective
			}
			prompt, perr := RenderPacket(Packet{
				AgentIdentity:   role.ID,
				RoleTitle:       role.Role,
				Project:         item.Project,
				Objective:       objective,
				ItemID:          item.ID,
				ItemTitle:       item.Title,
				ItemType:        item.Type,
				ItemState:       item.WorkState,
				ItemDescription: item.Description,
				Worktree:        worktree,
				Branch:          branch,
				FailedChecks:    failedChecks,
				CheckOutput:     checkOutput,
			})
			if perr != nil {
				return Outcome{}, perr
			}

			transcriptRel := filepath.Join("sessions", fmt.Sprintf("%d-%d.jsonl", rec.Pass, attempt))
			transcriptAbs := filepath.Join(RunDir(o.Root, runID), transcriptRel)

			// The loop owns the run directory layout (adapters never create
			// directories: a session that cannot start leaves no trace, C17), so the
			// sessions/ directory the adapter's TranscriptPath points into has to
			// exist before Start is called.
			if err := os.MkdirAll(filepath.Dir(transcriptAbs), 0o755); err != nil {
				return Outcome{}, fmt.Errorf("loop: creating sessions directory: %w", err)
			}

			sess, serr := adapter.Start(ctx, runtime.WorkPacket{
				RunID:          runID,
				Objective:      objective,
				Prompt:         prompt,
				Worktree:       worktree,
				Tools:          role.Tools.Allow,
				Deny:           role.Tools.Deny,
				TranscriptPath: transcriptAbs,
			})
			if serr != nil {
				return Outcome{}, serr
			}

			var (
				mu          sync.Mutex
				evSessionID string
				texts       []string
			)
			drainDone := make(chan struct{})
			go func() {
				defer close(drainDone)
				for ev := range sess.Events() {
					mu.Lock()
					if ev.SessionID != "" {
						evSessionID = ev.SessionID
					}
					if ev.Kind == runtime.EventText {
						texts = append(texts, ev.Text)
					}
					mu.Unlock()
				}
			}()

			var waitErr error
			result, waitErr = sess.Wait()
			<-drainDone

			if ctx.Err() != nil {
				// Exit only after Wait returns (it just did); the step stays entered
				// and unexited, so the next invocation resumes here (C28).
				return Outcome{RunID: runID, Record: rec}, ctx.Err()
			}
			if waitErr != nil {
				return Outcome{}, waitErr
			}

			mu.Lock()
			sid = evSessionID
			stepTexts := append([]string(nil), texts...)
			mu.Unlock()
			if sid == "" {
				sid = result.SessionID
			}

			questions = nil
			for _, t := range stepTexts {
				for _, line := range strings.Split(t, "\n") {
					if strings.HasPrefix(line, "QUESTION: ") {
						questions = append(questions, strings.TrimPrefix(line, "QUESTION: "))
					}
				}
			}

			last := &rec.Steps[len(rec.Steps)-1]
			last.Exited = nowStr()
			last.Delegates = []Delegate{{Kind: "runtime_session", Provider: provider, SessionID: sid}}
			last.Transcript = transcriptRel
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 5

		case 5: // collect
			entered := nowStr()
			var transcriptRel string
			if s4, ok := rec.Step(rec.Pass, 4); ok {
				if sid == "" && len(s4.Delegates) > 0 {
					sid = s4.Delegates[0].SessionID
				}
				transcriptRel = s4.Transcript
			}
			filesChanged, cerr := control.Status(worktree)
			if cerr != nil {
				return Outcome{}, cerr
			}
			exitStatus := result.ExitCode
			rec.Steps = append(rec.Steps, Step{
				N: 5, Name: StepNames[4], Pass: rec.Pass, Entered: entered, Exited: nowStr(),
				SessionID: sid, ExitStatus: &exitStatus, FilesChanged: filesChanged,
			})
			rec.AddEvidence("transcript", sid, transcriptRel, now())
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 6

		case 6: // resolve
			entered := nowStr()
			if len(questions) > 0 {
				rec.Steps = append(rec.Steps, Step{
					N: 6, Name: StepNames[5], Pass: rec.Pass, Entered: entered, Exited: nowStr(), Questions: questions,
				})
				rec.State = "blocked"
				rec.Question = questions[0]
				if err := SaveRecord(o.Root, rec, now()); err != nil {
					return Outcome{}, err
				}
				return Outcome{RunID: runID, Record: rec, Exit: 3}, nil
			}

			rec.Steps = append(rec.Steps, Step{
				N: 6, Name: StepNames[5], Pass: rec.Pass, Entered: entered, Exited: nowStr(), Questions: []string{},
			})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}

			if result.Status == runtime.StatusFailed {
				// Amendment (b): a session that exited non-zero with no QUESTION line
				// still needs a non-empty re-entry packet, so synthesize one check
				// naming the failure and run it through the same re-entry/block
				// decision as a failing local check.
				synth := synthesizeSessionCheck(o.Root, runID, result)
				gate := &Gate{ID: gateID, HumanOwned: role.Completion.HumanOwned, Surface: "github_review"}
				v7 := nowStr()
				rec.Steps = append(rec.Steps, Step{
					N: 7, Name: StepNames[6], Pass: rec.Pass, Entered: v7, Exited: nowStr(),
					Checks: []Check{synth}, Gate: gate,
				})
				rec.AddEvidence("check", sid, synth.Output, now())
				if err := SaveRecord(o.Root, rec, now()); err != nil {
					return Outcome{}, err
				}

				if overReentryLimit(rec) {
					rec.State = "blocked"
					rec.Question = fmt.Sprintf("local checks failing: %s", synth.Name)
					if err := SaveRecord(o.Root, rec, now()); err != nil {
						return Outcome{}, err
					}
					return Outcome{RunID: runID, Record: rec, Exit: 3}, nil
				}
				failedChecks = []Check{synth}
				checkOutput = map[string]string{synth.Name: result.Stderr}
				step = 3
				continue
			}
			step = 7

		case 7: // validate
			entered := nowStr()
			prefix := fmt.Sprintf("pass-%d-%d", rec.Pass, countStepEntries(rec, rec.Pass, 3))
			checks, verr := checker.Run(ctx, worktree, RunDir(o.Root, runID), prefix)
			if verr != nil {
				return Outcome{}, verr
			}
			gate := &Gate{ID: gateID, HumanOwned: role.Completion.HumanOwned, Surface: "github_review"}
			rec.Steps = append(rec.Steps, Step{
				N: 7, Name: StepNames[6], Pass: rec.Pass, Entered: entered, Exited: nowStr(), Checks: checks, Gate: gate,
			})
			for _, c := range checks {
				rec.AddEvidence("check", sid, c.Output, now())
			}
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}

			if !allChecksPassed(checks) {
				if overReentryLimit(rec) {
					rec.State = "blocked"
					rec.Question = fmt.Sprintf("local checks failing: %s", strings.Join(namesOf(failingChecks(checks)), ", "))
					if err := SaveRecord(o.Root, rec, now()); err != nil {
						return Outcome{}, err
					}
					return Outcome{RunID: runID, Record: rec, Exit: 3}, nil
				}
				failedChecks = failingChecks(checks)
				checkOutput = map[string]string{}
				for _, c := range failedChecks {
					content, _ := os.ReadFile(filepath.Join(RunDir(o.Root, runID), c.Output))
					checkOutput[c.Name] = string(content)
				}
				step = 3
				continue
			}

			cp := rec.currentPass()
			objective := ""
			if cp != nil {
				objective = cp.Objective
			}
			commitMsg := CommitMessage(*item, objective)
			headSHA, herr := control.Commit(worktree, commitMsg)
			if herr != nil {
				return Outcome{}, herr
			}
			if perr := control.Push(worktree, branch); perr != nil {
				return Outcome{}, perr
			}
			prBody, berr := RenderPRBody(rec, *item)
			if berr != nil {
				return Outcome{}, berr
			}
			title := strings.SplitN(commitMsg, "\n", 2)[0]
			prURL, cerr := control.CreatePR(worktree, branch, title, prBody)
			if cerr != nil {
				return Outcome{}, cerr
			}

			bodyRel := filepath.Join("handoff", "pr-body.md")
			bodyAbs := filepath.Join(RunDir(o.Root, runID), bodyRel)
			if err := os.MkdirAll(filepath.Dir(bodyAbs), 0o755); err != nil {
				return Outcome{}, err
			}
			if err := os.WriteFile(bodyAbs, []byte(prBody), 0o644); err != nil {
				return Outcome{}, err
			}

			rec.Handoff = &Handoff{PRURL: prURL, Branch: branch, HeadCommit: headSHA, Opened: nowStr(), State: "open"}
			if cp != nil {
				cp.PRURL = prURL
			}
			rec.AddEvidence("pull_request", sid, bodyRel, now())
			rec.State = "awaiting_human"
			closePassOutcome(rec, "parked", nowStr())
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			return Outcome{RunID: runID, Record: rec, Exit: 0}, nil

		case 8: // produce
			entered := nowStr()
			rec.Steps = append(rec.Steps, Step{
				N: 8, Name: StepNames[7], Pass: rec.Pass, Entered: entered, Exited: nowStr(),
				Artifacts: []string{"pull request", "evidence package", "run record"},
			})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 9

		case 9: // handoff
			entered := nowStr()
			note := "accepted by the human's merge; no downstream role exists in the first mile"
			if mergeCommit != "" {
				note = fmt.Sprintf("%s; merge commit %s", note, mergeCommit)
			}
			rec.Steps = append(rec.Steps, Step{
				N: 9, Name: StepNames[8], Pass: rec.Pass, Entered: entered, Exited: nowStr(), Note: note,
			})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 10

		case 10: // consolidate
			entered := nowStr()
			memDir := filepath.Join(RunDir(o.Root, runID), "memory-candidates")
			if err := os.MkdirAll(memDir, 0o755); err != nil {
				return Outcome{}, err
			}
			relPath := filepath.Join("memory-candidates", fmt.Sprintf("pass-%d.md", rec.Pass))
			absPath := filepath.Join(RunDir(o.Root, runID), relPath)
			cp := rec.currentPass()
			objective := ""
			prURL := ""
			if cp != nil {
				objective = cp.Objective
				prURL = cp.PRURL
			}
			var checkSummary string
			if s7, ok := rec.Step(rec.Pass, 7); ok {
				var b strings.Builder
				for _, c := range s7.Checks {
					fmt.Fprintf(&b, "- %s: rc %d\n", c.Name, c.RC)
				}
				checkSummary = b.String()
			}
			content := fmt.Sprintf("# Pass %d memory candidate\n\n## Objective\n\n%s\n\n## What changed\n\n%s\n\n## Checks\n\n%s\n## Pull request\n\n%s\n",
				rec.Pass, objective, strings.Join(filesChangedForPass(rec, rec.Pass), "\n"), checkSummary, prURL)
			if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
				return Outcome{}, err
			}

			rec.Steps = append(rec.Steps, Step{
				N: 10, Name: StepNames[9], Pass: rec.Pass, Entered: entered, Exited: nowStr(),
				Artifacts: []string{relPath},
			})
			rec.State = "idle"
			rec.Complete = true
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			return Outcome{RunID: runID, Record: rec, Exit: 0}, nil

		default:
			return Outcome{}, fmt.Errorf("loop: unknown step %d", step)
		}
	}
}

// hydrateInputs reads and hashes ADR-023 step 2's exactly three inputs, in order.
func hydrateInputs(root config.Root, item *config.WorkItem, roleID string) ([]Input, error) {
	rels := []string{
		filepath.Join(".agentic", "work", item.ID+".yaml"),
		filepath.Join(".agentic", "roles", roleID+".yaml"),
		filepath.Join("docs", "context", "CONTEXT_PACKET.md"),
	}
	inputs := make([]Input, 0, len(rels))
	for _, rel := range rels {
		b, err := os.ReadFile(root.Path(rel))
		if err != nil {
			return nil, fmt.Errorf("hydrate: %s: %w", rel, err)
		}
		inputs = append(inputs, Input{Path: rel, SHA256: sha256Hex(b)})
	}
	return inputs, nil
}

// countStepEntries returns how many entries step n has in the given pass.
func countStepEntries(rec *Record, pass, n int) int {
	count := 0
	for _, s := range rec.Steps {
		if s.Pass == pass && s.N == n {
			count++
		}
	}
	return count
}

// overReentryLimit reports whether step 3 (analyze) has already been entered the
// maximum number of times this pass allows: once plus maxReentries.
func overReentryLimit(rec *Record) bool {
	return countStepEntries(rec, rec.Pass, 3) >= 1+maxReentries
}

func allChecksPassed(checks []Check) bool {
	if len(checks) == 0 {
		return false
	}
	for _, c := range checks {
		if c.RC != 0 {
			return false
		}
	}
	return true
}

func failingChecks(checks []Check) []Check {
	var out []Check
	for _, c := range checks {
		if c.RC != 0 {
			out = append(out, c)
		}
	}
	return out
}

func namesOf(checks []Check) []string {
	names := make([]string, len(checks))
	for i, c := range checks {
		names[i] = c.Name
	}
	return names
}

func filesChangedForPass(rec *Record, pass int) []string {
	if s5, ok := rec.Step(pass, 5); ok {
		return s5.FilesChanged
	}
	return nil
}

// closePassOutcome closes the current pass with an outcome and timestamp.
func closePassOutcome(rec *Record, outcome, closed string) {
	if cp := rec.currentPass(); cp != nil {
		cp.Outcome = outcome
		cp.Closed = closed
	}
}

// synthesizeSessionCheck builds the Check amendment (b) requires when a session exits
// non-zero with no QUESTION line: the re-entry packet must not be empty, so the
// session's own stderr becomes a check named "session" that failed.
func synthesizeSessionCheck(root config.Root, runID string, result runtime.Result) Check {
	rel := filepath.Join("checks", fmt.Sprintf("session-%s.log", strings.TrimSpace(result.SessionID)))
	if result.SessionID == "" {
		rel = filepath.Join("checks", "session.log")
	}
	abs := filepath.Join(RunDir(root, runID), rel)
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	_ = os.WriteFile(abs, []byte(result.Stderr), 0o644)
	return Check{Name: "session", Command: "claude -p", RC: result.ExitCode, Output: rel}
}
