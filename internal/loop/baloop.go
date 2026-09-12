package loop

import (
	"context"
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

const (
	baGateID        = "product_spec"
	defaultBARoleID = "ba.primary"
)

// IntentOptions configures one BA loop invocation for a human product idea.
type IntentOptions struct {
	Root config.Root
	// Idea is the human's product objective. Required when opening a new run.
	Idea string
	// RunID resumes an existing intent run. When empty and Idea is set, opens a new run
	// unless ResumeOpen is true.
	RunID string
	// Answer records the human's response to a blocked question on resume.
	Answer string
	// ResumeOpen continues the newest open intent run when RunID and Idea are empty.
	ResumeOpen bool
	Provider   string
	RoleID     string
	Adapter    runtime.Adapter
	Now        func() time.Time
	Out        io.Writer
}

// RunIntent executes ADR-023's loop for the BA profile: conversational intent capture,
// specialist preflight, clarifying questions, and parking on the product_spec gate.
func RunIntent(ctx context.Context, o IntentOptions) (Outcome, error) {
	now := o.Now
	if now == nil {
		now = time.Now
	}
	nowStr := func() string { return now().UTC().Format(time.RFC3339) }

	if o.Root != "" {
		if abs, aerr := filepath.Abs(string(o.Root)); aerr == nil {
			o.Root = config.Root(abs)
		}
	}

	proj, err := config.LoadProject(o.Root)
	if err != nil {
		return Outcome{}, err
	}
	roleID := o.RoleID
	if roleID == "" {
		roleID = defaultBARoleID
	}
	role, err := config.LoadRole(o.Root, roleID)
	if err != nil {
		return Outcome{}, err
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

	var (
		rec       *Record
		runID     string
		startStep int
		answers   []string
		questions []string
	)

	switch {
	case o.RunID != "":
		runID = o.RunID
		rec, err = LoadRecord(o.Root, runID)
		if err != nil {
			return Outcome{}, err
		}
		if rec.Intent == nil {
			return Outcome{}, fmt.Errorf("run %q is not an intent run", runID)
		}
		if rec.State == "blocked" {
			if o.Answer == "" {
				return Outcome{RunID: runID, Record: rec, Exit: 3}, nil
			}
			rec.State = "working"
			rec.Question = ""
			answers = priorAnswers(rec)
			answers = append(answers, o.Answer)
			questions = priorQuestions(rec)
			startStep = 4
		} else {
			startStep = rec.Resume()
			if rec.State == "awaiting_human" || rec.State == "idle" {
				rec.State = "working"
			}
		}
	case o.ResumeOpen:
		found, ok, err := FindOpenIntentRun(o.Root)
		if err != nil {
			return Outcome{}, err
		}
		if !ok {
			return Outcome{}, fmt.Errorf("no open intent run to resume")
		}
		o.RunID = found
		return RunIntent(ctx, o)
	default:
		if strings.TrimSpace(o.Idea) == "" {
			return Outcome{}, fmt.Errorf("idea is required to open a new intent run")
		}
		runID, err = NewIntentRunID(o.Root)
		if err != nil {
			return Outcome{}, err
		}
		created := nowStr()
		rec = &Record{
			Schema:        "run-record/v1",
			RunID:         runID,
			AgentIdentity: role.ID,
			RoleVersion:   fmt.Sprintf("%s@%d", role.Role, role.RoleVersion),
			Runtime:       Runtime{Provider: provider, Model: "provider-current", Override: override},
			Work:          Work{Item: "INTENT", Title: "Product intent capture", WorkStateAtDispatch: "intake"},
			Intent:        &Intent{Idea: strings.TrimSpace(o.Idea)},
			State:         "working",
			Complete:      false,
			Created:       created,
			Pass:          1,
			Passes:        []Pass{{N: 1, Objective: o.Idea, Opened: created}},
			Stubbed: []string{
				"PRD/ADS projection (WI-0051): step 8 emits draft artifacts only after human approval",
				"human approval record: product_spec gate closes only on an explicit approval under .agentic/approvals/",
			},
		}
		startStep = 1
	}

	runDir := RunDir(o.Root, runID)
	var (
		result  runtime.Result
		sid     string
		waitErr error
	)

	step := startStep
	for {
		switch step {
		case 1:
			ts := nowStr()
			rec.Steps = append(rec.Steps, Step{
				N: 1, Name: StepNames[0], Pass: rec.Pass, Entered: ts, Exited: ts,
				Note: "human idea accepted for product analysis",
			})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 2

		case 2:
			entered := nowStr()
			inputs, herr := hydrateBAInputs(o.Root, roleID, rec.Intent.Idea)
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

		case 3:
			entered := nowStr()
			reports, rerr := runSpecialistPreflight(runDir, rec.Intent.Idea)
			if rerr != nil {
				return Outcome{}, rerr
			}
			var note strings.Builder
			for _, r := range reports {
				fmt.Fprintf(&note, "- %s (%s)\n", r.Role, r.Path)
			}
			rec.Steps = append(rec.Steps, Step{
				N: 3, Name: StepNames[2], Pass: rec.Pass, Entered: entered, Exited: nowStr(),
				Note: strings.TrimSpace(note.String()),
			})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 4

		case 4:
			entered := nowStr()
			attempt := countStepEntries(rec, rec.Pass, 4) + 1
			rec.Steps = append(rec.Steps, Step{N: 4, Name: StepNames[3], Pass: rec.Pass, Entered: entered})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}

			specNotes, _ := os.ReadFile(filepath.Join(runDir, "specialists", "compliance.md"))
			legalNotes, _ := os.ReadFile(filepath.Join(runDir, "specialists", "legal.md"))
			specialistNotes := strings.TrimSpace(string(specNotes) + "\n\n" + string(legalNotes))

			prompt, perr := RenderBAPacket(BAPacket{
				AgentIdentity:   role.ID,
				RoleTitle:       role.Role,
				Project:         projectLabel(proj),
				Idea:            rec.Intent.Idea,
				SpecialistNotes: specialistNotes,
				PriorQuestions:  questions,
				HumanAnswers:    answers,
			})
			if perr != nil {
				return Outcome{}, perr
			}

			transcriptRel := filepath.Join("sessions", fmt.Sprintf("%d-%d.jsonl", rec.Pass, attempt))
			transcriptAbs := filepath.Join(runDir, transcriptRel)
			if err := os.MkdirAll(filepath.Dir(transcriptAbs), 0o755); err != nil {
				return Outcome{}, fmt.Errorf("baloop: creating sessions directory: %w", err)
			}

			sess, serr := adapter.Start(ctx, runtime.WorkPacket{
				RunID:          runID,
				Objective:      rec.Intent.Idea,
				Prompt:         prompt,
				Worktree:       string(o.Root),
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
					if ev.Kind == runtime.EventText || ev.Kind == runtime.EventResult {
						texts = append(texts, ev.Text)
					}
					mu.Unlock()
				}
			}()

			result, waitErr = sess.Wait()
			<-drainDone
			if ctx.Err() != nil {
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
			hasSpecOutline := false
			for _, t := range stepTexts {
				for _, line := range strings.Split(t, "\n") {
					if strings.HasPrefix(line, "QUESTION: ") {
						questions = append(questions, strings.TrimPrefix(line, "QUESTION: "))
					}
					if strings.HasPrefix(line, "SPEC OUTLINE:") {
						hasSpecOutline = true
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
			if !hasSpecOutline && len(questions) == 0 && result.Status == runtime.StatusSucceeded {
				questions = []string{"What actors and success criteria should this specification assume?"}
			}
			step = 5

		case 5:
			entered := nowStr()
			var transcriptRel string
			if s4, ok := rec.Step(rec.Pass, 4); ok {
				transcriptRel = s4.Transcript
			}
			exitStatus := result.ExitCode
			rec.Steps = append(rec.Steps, Step{
				N: 5, Name: StepNames[4], Pass: rec.Pass, Entered: entered, Exited: nowStr(),
				SessionID: sid, ExitStatus: &exitStatus,
			})
			rec.AddEvidence("transcript", sid, transcriptRel, now())
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			step = 6

		case 6:
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
				N: 6, Name: StepNames[5], Pass: rec.Pass, Entered: entered, Exited: nowStr(),
			})
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			if result.Status == runtime.StatusFailed {
				rec.State = "blocked"
				rec.Question = "runtime session failed; inspect the transcript"
				if err := SaveRecord(o.Root, rec, now()); err != nil {
					return Outcome{}, err
				}
				return Outcome{RunID: runID, Record: rec, Exit: 3}, nil
			}
			step = 7

		case 7:
			entered := nowStr()
			gate := &Gate{ID: baGateID, HumanOwned: true, Surface: "human_approval"}
			body := strings.Join(collectSessionText(rec, rec.Pass), "\n\n")
			specRel := filepath.Join("handoff", fmt.Sprintf("pass-%d-spec-review.md", rec.Pass))
			specAbs := filepath.Join(runDir, specRel)
			if err := os.MkdirAll(filepath.Dir(specAbs), 0o755); err != nil {
				return Outcome{}, err
			}
			content := fmt.Sprintf("# Product specification review\n\nGate: %s (human-owned)\n\n## Intent\n\n%s\n\n## BA session output\n\n%s\n\n## Status\n\nAwaiting explicit human approval. Withholding approval leaves this run open.\n",
				baGateID, rec.Intent.Idea, body)
			if err := os.WriteFile(specAbs, []byte(content), 0o644); err != nil {
				return Outcome{}, err
			}
			rec.Steps = append(rec.Steps, Step{
				N: 7, Name: StepNames[6], Pass: rec.Pass, Entered: entered, Exited: nowStr(),
				Gate: gate, Artifacts: []string{specRel},
			})
			rec.AddEvidence("spec_review", sid, specRel, now())
			rec.State = "awaiting_human"
			closePassOutcome(rec, "parked", nowStr())
			if err := SaveRecord(o.Root, rec, now()); err != nil {
				return Outcome{}, err
			}
			return Outcome{RunID: runID, Record: rec, Exit: 0}, nil

		default:
			return Outcome{}, fmt.Errorf("baloop: step %d not implemented for BA profile", step)
		}
	}
}

func projectLabel(proj *config.Project) string {
	if proj != nil && proj.Project != "" {
		return proj.Project
	}
	return "project"
}

func hydrateBAInputs(root config.Root, roleID, idea string) ([]Input, error) {
	rels := []string{
		filepath.Join(".agentic", "roles", roleID+".yaml"),
		filepath.Join("docs", "context", "CONTEXT_PACKET.md"),
	}
	inputs := make([]Input, 0, len(rels)+1)
	ideaBytes := []byte(idea)
	inputs = append(inputs, Input{Path: "intent.idea", SHA256: sha256Hex(ideaBytes)})
	for _, rel := range rels {
		b, err := os.ReadFile(root.Path(rel))
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, Input{Path: rel, SHA256: sha256Hex(b)})
	}
	return inputs, nil
}

func priorQuestions(rec *Record) []string {
	var qs []string
	for _, s := range rec.Steps {
		if s.N == 6 && len(s.Questions) > 0 {
			qs = append(qs, s.Questions...)
		}
	}
	return qs
}

func priorAnswers(rec *Record) []string {
	// Answers are not yet persisted as structured fields; future slice may add them.
	return nil
}

func collectSessionText(rec *Record, pass int) []string {
	var out []string
	s4, ok := rec.Step(pass, 4)
	if !ok || s4.Transcript == "" {
		return out
	}
	// Transcript is JSONL; return path reference for the handoff doc body above uses step notes.
	return []string{fmt.Sprintf("(see transcript at %s)", s4.Transcript)}
}
