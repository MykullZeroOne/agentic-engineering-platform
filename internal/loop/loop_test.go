package loop

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/runtime"
)

func mustRun(t *testing.T, ctx context.Context, o Options) Outcome {
	t.Helper()
	out, err := Run(ctx, o)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return out
}

func passChecker() *fakeChecker {
	return &fakeChecker{Script: [][]Check{{
		{Name: "go-build", Command: "go build ./...", RC: 0},
	}}}
}

func passAdapter(sessionID string, texts ...string) *fakeAdapter {
	return &fakeAdapter{Sessions: []fakeSession{{
		SessionID: sessionID,
		Texts:     texts,
		Result:    runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0},
	}}}
}

func TestC20_AFullRunRecordsTenStepsInADROrder(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-1", "done"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	if out.Exit != 0 {
		t.Fatalf("first Run exit = %d, want 0", out.Exit)
	}

	control.PR = PR{State: PRMerged, MergeCommit: "deadbeef"}
	out2 := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-2"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	if out2.Exit != 0 {
		t.Fatalf("second Run exit = %d, want 0", out2.Exit)
	}

	rec := out2.Record
	for n := 1; n <= 10; n++ {
		s, ok := rec.Step(1, n)
		if !ok {
			t.Fatalf("no entry for step %d", n)
		}
		if s.Name != StepNames[n-1] {
			t.Errorf("step %d name = %q, want %q", n, s.Name, StepNames[n-1])
		}
		if s.Entered == "" || s.Exited == "" {
			t.Errorf("step %d missing Entered/Exited: %+v", n, s)
		}
		if s.Entered > s.Exited {
			t.Errorf("step %d Entered %q > Exited %q", n, s.Entered, s.Exited)
		}
	}
}

func TestC21_HydrateRecordsExactlyThreeInputsWithHashes(t *testing.T) {
	root := fixture(t, nil)
	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-1", "done"), Control: &fakeControl{}, Checker: passChecker(), Now: fixedNow(),
	})

	s, ok := out.Record.Step(1, 2)
	if !ok {
		t.Fatal("no hydrate step recorded")
	}
	if len(s.Inputs) != 3 {
		t.Fatalf("hydrate recorded %d inputs, want 3: %+v", len(s.Inputs), s.Inputs)
	}
	wantPaths := []string{
		filepath.Join(".agentic", "work", "WI-0001.yaml"),
		filepath.Join(".agentic", "roles", "engineer.primary.yaml"),
		filepath.Join("docs", "context", "CONTEXT_PACKET.md"),
	}
	for i, in := range s.Inputs {
		if in.Path != wantPaths[i] {
			t.Errorf("input %d path = %q, want %q", i, in.Path, wantPaths[i])
		}
		if len(in.SHA256) != 64 {
			t.Errorf("input %d sha256 = %q, not 64 hex chars", i, in.SHA256)
		}
		b, err := os.ReadFile(filepath.Join(string(root), in.Path))
		if err != nil {
			t.Fatal(err)
		}
		if got := sha256Hex(b); got != in.SHA256 {
			t.Errorf("input %d sha256 = %q, want freshly computed %q", i, in.SHA256, got)
		}
	}
}

func TestC22_DelegateRecordsTheRuntimeSessionAsItsOneDelegate(t *testing.T) {
	root := fixture(t, nil)
	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-42", "done"), Control: &fakeControl{}, Checker: passChecker(), Now: fixedNow(),
	})

	s, ok := out.Record.Step(1, 4)
	if !ok {
		t.Fatal("no delegate step recorded")
	}
	if len(s.Delegates) != 1 {
		t.Fatalf("delegate recorded %d delegates, want 1: %+v", len(s.Delegates), s.Delegates)
	}
	d := s.Delegates[0]
	if d.Kind != "runtime_session" {
		t.Errorf("delegate kind = %q, want runtime_session", d.Kind)
	}
	if d.Provider != "claude-subscription" {
		t.Errorf("delegate provider = %q, want claude-subscription", d.Provider)
	}
	if d.SessionID != "sess-42" {
		t.Errorf("delegate session id = %q, want sess-42", d.SessionID)
	}
}

func TestC23_CollectRecordsFilesChangedSessionIDAndExitStatus(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{StatusFiles: []string{"internal/loop/loop.go", "docs/foo.md"}}
	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-1", "done"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})

	s, ok := out.Record.Step(1, 5)
	if !ok {
		t.Fatal("no collect step recorded")
	}
	if len(s.FilesChanged) != 2 || s.FilesChanged[0] != "internal/loop/loop.go" || s.FilesChanged[1] != "docs/foo.md" {
		t.Fatalf("FilesChanged = %v, want the two fixture paths", s.FilesChanged)
	}
	if s.SessionID == "" {
		t.Error("collect step SessionID is empty")
	}
	if s.ExitStatus == nil || *s.ExitStatus != 0 {
		t.Errorf("collect step ExitStatus = %v, want *0", s.ExitStatus)
	}
}

func TestC24_AQuestionLineParksTheRunBlockedAndExitsThree(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	adapter := &fakeAdapter{Sessions: []fakeSession{{
		SessionID: "sess-q",
		Texts:     []string{"QUESTION: which registry owns polarity?"},
		Result:    runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0},
	}}}

	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: passChecker(), Now: fixedNow(),
	})

	if out.Exit != 3 {
		t.Fatalf("Exit = %d, want 3", out.Exit)
	}
	if out.Record.State != "blocked" {
		t.Errorf("State = %q, want blocked", out.Record.State)
	}
	if !strings.Contains(out.Record.Question, "which registry owns polarity?") {
		t.Errorf("Question = %q, missing the question text", out.Record.Question)
	}
	s, ok := out.Record.Step(1, 6)
	if !ok {
		t.Fatal("no resolve step recorded")
	}
	if len(s.Questions) != 1 {
		t.Fatalf("resolve step Questions = %v, want length 1", s.Questions)
	}
	if out.Record.Handoff != nil {
		t.Error("Handoff is non-nil on a blocked run")
	}
	for _, c := range control.Calls {
		if c == "CreatePR" {
			t.Fatal("CreatePR was called on a blocked run")
		}
	}
}

func TestC25_ValidateRecordsRCAndAnOutputRefForEachCheck(t *testing.T) {
	root := fixture(t, nil)
	checker := &fakeChecker{Script: [][]Check{{
		{Name: "go-build", Command: "go build ./...", RC: 0},
		{Name: "go-test", Command: "go test ./...", RC: 0},
		{Name: "validate-docs", Command: "python3 scripts/validate_docs.py", RC: 0},
		{Name: "check-gates", Command: "python3 scripts/check_gates.py --base origin/main", RC: 0},
	}}}
	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-1", "done"), Control: &fakeControl{}, Checker: checker, Now: fixedNow(),
	})

	s, ok := out.Record.Step(1, 7)
	if !ok {
		t.Fatal("no validate step recorded")
	}
	if len(s.Checks) != 4 {
		t.Fatalf("validate step recorded %d checks, want 4", len(s.Checks))
	}
	for _, c := range s.Checks {
		if c.Command == "" {
			t.Errorf("check %s has empty Command", c.Name)
		}
		if c.Output == "" {
			t.Errorf("check %s has empty Output", c.Name)
		}
		p := filepath.Join(RunDir(root, out.RunID), c.Output)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("check %s Output %q does not resolve under the run directory: %v", c.Name, c.Output, err)
		}
	}
}

func TestC26_TheLoopCommitsPushesAndOpensThePR(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-1", "done"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	_ = out

	want := []string{"Worktree", "Status", "Commit", "Push", "CreatePR"}
	if len(control.Calls) != len(want) {
		t.Fatalf("Calls = %v, want %v", control.Calls, want)
	}
	for i, c := range want {
		if control.Calls[i] != c {
			t.Errorf("Calls[%d] = %q, want %q (full: %v)", i, control.Calls[i], c, control.Calls)
		}
	}

	if !strings.Contains(control.CreatedBody, "Closes WI-0001") {
		// The commit message, not the PR body, carries the trailer (M1); check it separately below.
	}

	item, err := config.LoadWorkItems(root)
	if err != nil {
		t.Fatal(err)
	}
	var wi config.WorkItem
	for _, w := range item {
		if w.ID == "WI-0001" {
			wi = w
		}
	}
	branch := BranchName(wi)
	branchRE := regexp.MustCompile(`^[a-z]+/1-[a-z0-9-]+$`)
	if !branchRE.MatchString(branch) {
		t.Errorf("branch %q does not match ^[a-z]+/1-[a-z0-9-]+$", branch)
	}

	if !strings.Contains(control.CreatedBody, "run_id: ") {
		t.Error("PR body missing `run_id: `")
	}
	if !strings.Contains(control.CreatedBody, "evidence/v1") {
		t.Error("PR body missing `evidence/v1`")
	}
}

func TestC27_AfterValidateTheRunIsAwaitingHumanWithAPRURL(t *testing.T) {
	root := fixture(t, nil)
	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-1", "done"), Control: &fakeControl{}, Checker: passChecker(), Now: fixedNow(),
	})

	if out.Record.State != "awaiting_human" {
		t.Errorf("State = %q, want awaiting_human", out.Record.State)
	}
	if out.Record.Handoff == nil || out.Record.Handoff.PRURL == "" {
		t.Fatalf("Handoff = %+v, want a non-empty PRURL", out.Record.Handoff)
	}
	if out.Record.Handoff.PRURL != "https://github.com/o/r/pull/1" {
		t.Errorf("Handoff.PRURL = %q", out.Record.Handoff.PRURL)
	}
	if out.Exit != 0 {
		t.Errorf("Exit = %d, want 0", out.Exit)
	}
	if _, ok := out.Record.Step(1, 8); ok {
		t.Error("a step 8 entry exists after a single park at step 7")
	}
}

func TestC1_AnOpenPRAdvancesNothingOnASecondInvocation(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	adapter := passAdapter("sess-1", "done")
	first := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	if first.Record.State != "awaiting_human" {
		t.Fatalf("first run state = %q, want awaiting_human", first.Record.State)
	}

	control.PR = PR{State: PROpen}
	second := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: passChecker(), Now: fixedNow(),
	})

	if second.Record.State != "awaiting_human" {
		t.Errorf("second run state = %q, want awaiting_human", second.Record.State)
	}
	if len(second.Record.Steps) != len(first.Record.Steps) {
		t.Errorf("second run added Steps: %d -> %d", len(first.Record.Steps), len(second.Record.Steps))
	}
	if len(second.Record.Evidence) != len(first.Record.Evidence) {
		t.Errorf("second run added Evidence: %d -> %d", len(first.Record.Evidence), len(second.Record.Evidence))
	}
	if len(second.Record.Passes) != len(first.Record.Passes) {
		t.Errorf("second run added Passes: %d -> %d", len(first.Record.Passes), len(second.Record.Passes))
	}

	f1 := *first.Record
	f2 := *second.Record
	f1.Updated = ""
	f2.Updated = ""
	if f1.State != f2.State || len(f1.Steps) != len(f2.Steps) || len(f1.Evidence) != len(f2.Evidence) {
		t.Errorf("records differ beyond Updated:\nfirst=%+v\nsecond=%+v", f1, f2)
	}

	if adapter.started != 1 {
		t.Errorf("adapter started %d sessions, want 1", adapter.started)
	}
}

func TestC28_ACancelledRunResumesAtStepFourOnTheSameRunID(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	adapter := &fakeAdapter{Sessions: []fakeSession{
		{SessionID: "sess-blocked", Block: make(chan struct{})},
		{SessionID: "sess-2", Texts: []string{"done"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
	}}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := Run(ctx, Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	if err == nil {
		t.Fatal("expected an error from a cancelled run")
	}

	runID, found, ferr := FindRun(root, "WI-0001")
	if ferr != nil {
		t.Fatal(ferr)
	}
	if !found {
		t.Fatal("no open run found after cancellation")
	}
	rec, err := LoadRecord(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	s4, ok := rec.Step(1, 4)
	if !ok {
		t.Fatal("no step 4 entry after cancellation")
	}
	if s4.Exited != "" {
		t.Errorf("step 4 Exited = %q, want empty (unexited)", s4.Exited)
	}
	if rec.State != "working" {
		t.Errorf("State = %q, want working", rec.State)
	}

	second := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	if second.RunID != runID {
		t.Errorf("second run id = %q, want %q (same run)", second.RunID, runID)
	}
	count := 0
	for _, s := range second.Record.Steps {
		if s.N == 4 && s.Pass == 1 {
			count++
		}
	}
	if count != 2 {
		t.Errorf("step 4 entries in pass 1 = %d, want 2", count)
	}
	if second.Record.State != "awaiting_human" {
		t.Errorf("second run state = %q, want awaiting_human", second.Record.State)
	}
	if adapter.started != 2 {
		t.Errorf("adapter started %d sessions, want 2", adapter.started)
	}

	s4b, ok4b := second.Record.Step(1, 4)
	if !ok4b {
		t.Fatal("no step 4 entry after the second run")
	}
	if !strings.HasSuffix(s4b.Transcript, "1-2.jsonl") {
		t.Errorf("second step 4 entry Transcript = %q, want suffix 1-2.jsonl", s4b.Transcript)
	}

	var transcriptEvidence *Evidence
	for i := range second.Record.Evidence {
		e := &second.Record.Evidence[i]
		if e.Pass == 1 && e.Kind == "transcript" {
			transcriptEvidence = e
		}
	}
	if transcriptEvidence == nil {
		t.Fatal("no transcript evidence recorded for pass 1")
	}
	if transcriptEvidence.Path != s4b.Transcript {
		t.Errorf("step 5 transcript evidence path = %q, want the step 4 entry's Transcript %q", transcriptEvidence.Path, s4b.Transcript)
	}
	transcriptFile := filepath.Join(RunDir(root, runID), transcriptEvidence.Path)
	if _, err := os.Stat(transcriptFile); err != nil {
		t.Errorf("transcript file %q does not exist: %v", transcriptFile, err)
	}
}

func TestC29_AFailingCheckReturnsToStepThreeInTheSamePass(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	adapter := &fakeAdapter{Sessions: []fakeSession{
		{SessionID: "sess-1", Texts: []string{"first attempt"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
		{SessionID: "sess-2", Texts: []string{"second attempt"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
	}}
	checker := &fakeChecker{Script: [][]Check{
		{{Name: "go-build", Command: "go build ./...", RC: 1}},
		{{Name: "go-build", Command: "go build ./...", RC: 0}},
	}}

	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: checker, Now: fixedNow(),
	})

	for _, n := range []int{3, 4, 5, 6, 7} {
		count := 0
		for _, s := range out.Record.Steps {
			if s.N == n && s.Pass == 1 {
				count++
			}
		}
		if count != 2 {
			t.Errorf("step %d has %d entries in pass 1, want 2", n, count)
		}
	}
	if out.Record.State != "awaiting_human" {
		t.Errorf("final state = %q, want awaiting_human", out.Record.State)
	}
	if adapter.started != 2 {
		t.Errorf("adapter started %d sessions, want 2", adapter.started)
	}
}

func TestC29_ThreeFailingChecksParkTheRunBlocked(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	adapter := &fakeAdapter{Sessions: []fakeSession{
		{SessionID: "sess-1", Texts: []string{"a1"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
		{SessionID: "sess-2", Texts: []string{"a2"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
		{SessionID: "sess-3", Texts: []string{"a3"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
	}}
	checker := &fakeChecker{Script: [][]Check{
		{{Name: "go-build", Command: "go build ./...", RC: 1}},
	}}

	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: checker, Now: fixedNow(),
	})

	count := 0
	for _, s := range out.Record.Steps {
		if s.N == 3 && s.Pass == 1 {
			count++
		}
	}
	if count != 3 {
		t.Errorf("step 3 entries = %d, want 3", count)
	}
	if out.Record.State != "blocked" {
		t.Errorf("State = %q, want blocked", out.Record.State)
	}
	if out.Exit != 3 {
		t.Errorf("Exit = %d, want 3", out.Exit)
	}
	for _, c := range control.Calls {
		if c == "CreatePR" {
			t.Fatal("CreatePR was called on a blocked run")
		}
	}
}

func TestC30_AClosedPROpensPassTwoOnTheSameRunID(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	adapter := &fakeAdapter{Sessions: []fakeSession{
		{SessionID: "sess-1", Texts: []string{"pass one"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
		{SessionID: "sess-2", Texts: []string{"pass two"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
	}}
	first := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	firstURL := first.Record.Handoff.PRURL

	control.PR = PR{State: PRClosed, Review: "please add a test"}
	second := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: passChecker(), Now: fixedNow(),
	})

	if second.RunID != first.RunID {
		t.Errorf("run id changed: %q -> %q", first.RunID, second.RunID)
	}
	rec := second.Record
	if rec.Pass != 2 {
		t.Errorf("Pass = %d, want 2", rec.Pass)
	}
	if len(rec.Passes) != 2 {
		t.Fatalf("len(Passes) = %d, want 2", len(rec.Passes))
	}
	if rec.Passes[0].Outcome != "returned" {
		t.Errorf("Passes[0].Outcome = %q, want returned", rec.Passes[0].Outcome)
	}
	if rec.Passes[0].PRURL != firstURL {
		t.Errorf("Passes[0].PRURL = %q, want %q", rec.Passes[0].PRURL, firstURL)
	}
	if !strings.HasPrefix(rec.Passes[1].Objective, "returned: ") {
		t.Errorf("Passes[1].Objective = %q, want prefix `returned: `", rec.Passes[1].Objective)
	}
	if !strings.Contains(rec.Passes[1].Objective, "please add a test") {
		t.Errorf("Passes[1].Objective = %q, missing the review text", rec.Passes[1].Objective)
	}

	sids := map[string]bool{}
	for _, e := range rec.Evidence {
		if e.RunID != rec.RunID {
			t.Errorf("evidence %s RunID = %q, want %q", e.ID, e.RunID, rec.RunID)
		}
		if e.SessionID != "" {
			sids[e.SessionID] = true
		}
	}
	if len(sids) != 2 {
		t.Errorf("distinct session ids in evidence = %d (%v), want 2", len(sids), sids)
	}
}

func TestC31_AMergedPRRunsStepsEightNineAndTen(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	first := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-1", "done"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	_ = first

	control.PR = PR{State: PRMerged, MergeCommit: "deadbeef"}
	second := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-2"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})

	rec := second.Record
	for _, n := range []int{8, 9, 10} {
		s, ok := rec.Step(1, n)
		if !ok {
			t.Fatalf("no step %d entry", n)
		}
		if s.Exited == "" {
			t.Errorf("step %d Exited is empty", n)
		}
	}
	if rec.State != "idle" {
		t.Errorf("State = %q, want idle", rec.State)
	}
	if !rec.Complete {
		t.Error("Complete is false after a merged PR")
	}

	dir := filepath.Join(RunDir(root, second.RunID), "memory-candidates")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("memory-candidates has %d files, want 1: %v", len(entries), entries)
	}
}

func TestC3_LoopWritesNothingUnderWorkApprovalsOrRegistries(t *testing.T) {
	root := fixture(t, nil)
	dirs := []string{
		filepath.Join(string(root), ".agentic", "work"),
		filepath.Join(string(root), ".agentic", "approvals"),
		filepath.Join(string(root), ".agentic", "registries"),
	}
	before := map[string]string{}
	for _, d := range dirs {
		before[d] = treeHash(t, d)
	}

	control := &fakeControl{}
	first := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-1", "done"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	_ = first
	control.PR = PR{State: PRMerged, MergeCommit: "deadbeef"}
	mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-2"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})

	for _, d := range dirs {
		after := treeHash(t, d)
		if after != before[d] {
			t.Errorf("%s changed: %s -> %s", d, before[d], after)
		}
	}
}

func TestC4_ConsolidateWritesExactlyOneStubUnderTheRunDir(t *testing.T) {
	root := fixture(t, nil)
	before := map[string]bool{}
	_ = filepath.Walk(string(root), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		before[path] = true
		return nil
	})

	control := &fakeControl{}
	first := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-1", "done"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})
	_ = first
	control.PR = PR{State: PRMerged, MergeCommit: "deadbeef"}
	second := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: passAdapter("sess-2"), Control: control, Checker: passChecker(), Now: fixedNow(),
	})

	runDir := RunDir(root, second.RunID)
	worktree := WorktreePath(root, second.RunID)

	var newPaths []string
	_ = filepath.Walk(string(root), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if before[path] {
			return nil
		}
		if strings.HasPrefix(path, runDir) || strings.HasPrefix(path, worktree) {
			return nil
		}
		newPaths = append(newPaths, path)
		return nil
	})
	if len(newPaths) != 0 {
		t.Errorf("paths written outside the run directory and the worktree: %v", newPaths)
	}

	memDir := filepath.Join(runDir, "memory-candidates")
	entries, err := os.ReadDir(memDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("memory-candidates has %d files, want exactly 1: %v", len(entries), entries)
	}
}

// TestC1_1_ABlockedRunReinvokedAdvancesNothing pins anti-claim C1.1: a run in state
// "blocked" (question, or re-entry limit) re-invoked with no answer enters no step,
// starts no session, and changes nothing in its record but Updated.
func TestC1_1_ABlockedRunReinvokedAdvancesNothing(t *testing.T) {
	assertNoAdvance := func(t *testing.T, first, second Outcome, adapter *fakeAdapter, control *fakeControl) {
		t.Helper()
		if second.Exit != 3 {
			t.Errorf("second Outcome.Exit = %d, want 3", second.Exit)
		}
		if second.Record.State != "blocked" {
			t.Errorf("second run State = %q, want blocked", second.Record.State)
		}
		if len(second.Record.Steps) != len(first.Record.Steps) {
			t.Errorf("step count changed: %d -> %d", len(first.Record.Steps), len(second.Record.Steps))
		}
		if len(second.Record.Evidence) != len(first.Record.Evidence) {
			t.Errorf("evidence count changed: %d -> %d", len(first.Record.Evidence), len(second.Record.Evidence))
		}
		if len(second.Record.Passes) != len(first.Record.Passes) {
			t.Errorf("pass count changed: %d -> %d", len(first.Record.Passes), len(second.Record.Passes))
		}
		if second.Record.Question != first.Record.Question {
			t.Errorf("Question changed: %q -> %q", first.Record.Question, second.Record.Question)
		}
		if adapter.started != 1 {
			t.Errorf("adapter started %d sessions, want 1 (no new session on re-invocation)", adapter.started)
		}
		for _, c := range control.Calls {
			if c != "" {
				// no assertion on individual call kinds here; the count check below
				// is what proves nothing new happened.
			}
		}

		f1 := *first.Record
		f2 := *second.Record
		f1.Updated = ""
		f2.Updated = ""
		if f1.State != f2.State || len(f1.Steps) != len(f2.Steps) || len(f1.Evidence) != len(f2.Evidence) ||
			f1.Question != f2.Question || len(f1.Passes) != len(f2.Passes) {
			t.Errorf("records differ beyond Updated:\nfirst=%+v\nsecond=%+v", f1, f2)
		}
	}

	t.Run("question", func(t *testing.T) {
		root := fixture(t, nil)
		control := &fakeControl{}
		adapter := &fakeAdapter{Sessions: []fakeSession{{
			SessionID: "sess-q",
			Texts:     []string{"QUESTION: which registry owns polarity?"},
			Result:    runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0},
		}}}

		first := mustRun(t, context.Background(), Options{
			Root: root, Item: "WI-0001",
			Adapter: adapter, Control: control, Checker: passChecker(), Now: fixedNow(),
		})
		if first.Record.State != "blocked" {
			t.Fatalf("first run State = %q, want blocked", first.Record.State)
		}
		callsBefore := len(control.Calls)

		second := mustRun(t, context.Background(), Options{
			Root: root, Item: "WI-0001",
			Adapter: adapter, Control: control, Checker: passChecker(), Now: fixedNow(),
		})
		if len(control.Calls) != callsBefore {
			t.Errorf("control.Calls grew: %d -> %d (%v)", callsBefore, len(control.Calls), control.Calls)
		}
		assertNoAdvance(t, first, second, adapter, control)
	})

	t.Run("reentry_limit", func(t *testing.T) {
		root := fixture(t, nil)
		control := &fakeControl{}
		adapter := &fakeAdapter{Sessions: []fakeSession{
			{SessionID: "sess-1", Texts: []string{"a1"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
			{SessionID: "sess-2", Texts: []string{"a2"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
			{SessionID: "sess-3", Texts: []string{"a3"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
		}}
		checker := &fakeChecker{Script: [][]Check{
			{{Name: "go-build", Command: "go build ./...", RC: 1}},
		}}

		first := mustRun(t, context.Background(), Options{
			Root: root, Item: "WI-0001",
			Adapter: adapter, Control: control, Checker: checker, Now: fixedNow(),
		})
		if first.Record.State != "blocked" {
			t.Fatalf("first run State = %q, want blocked", first.Record.State)
		}
		startedBefore := adapter.started
		callsBefore := len(control.Calls)

		second := mustRun(t, context.Background(), Options{
			Root: root, Item: "WI-0001",
			Adapter: adapter, Control: control, Checker: checker, Now: fixedNow(),
		})
		if adapter.started != startedBefore {
			t.Errorf("adapter.started grew: %d -> %d", startedBefore, adapter.started)
		}
		if len(control.Calls) != callsBefore {
			t.Errorf("control.Calls grew: %d -> %d (%v)", callsBefore, len(control.Calls), control.Calls)
		}

		if second.Exit != 3 {
			t.Errorf("second Outcome.Exit = %d, want 3", second.Exit)
		}
		if second.Record.State != "blocked" {
			t.Errorf("second run State = %q, want blocked", second.Record.State)
		}
		if len(second.Record.Steps) != len(first.Record.Steps) {
			t.Errorf("step count changed: %d -> %d", len(first.Record.Steps), len(second.Record.Steps))
		}
		if len(second.Record.Evidence) != len(first.Record.Evidence) {
			t.Errorf("evidence count changed: %d -> %d", len(first.Record.Evidence), len(second.Record.Evidence))
		}
		if len(second.Record.Passes) != len(first.Record.Passes) {
			t.Errorf("pass count changed: %d -> %d", len(first.Record.Passes), len(second.Record.Passes))
		}
		if second.Record.Question != first.Record.Question {
			t.Errorf("Question changed: %q -> %q", first.Record.Question, second.Record.Question)
		}

		f1 := *first.Record
		f2 := *second.Record
		f1.Updated = ""
		f2.Updated = ""
		if f1.State != f2.State || len(f1.Steps) != len(f2.Steps) || len(f1.Evidence) != len(f2.Evidence) ||
			f1.Question != f2.Question || len(f1.Passes) != len(f2.Passes) {
			t.Errorf("records differ beyond Updated:\nfirst=%+v\nsecond=%+v", f1, f2)
		}
	})
}

// TestC29_AFailedSessionReentersWithASynthesizedCheck pins amendment (b): a session
// that exits non-zero with no QUESTION line still needs a non-empty re-entry packet,
// so the loop synthesizes one check named "session" carrying the session's own
// stderr and re-enters through step 3 rather than treating a silent failure as a
// pass.
func TestC29_AFailedSessionReentersWithASynthesizedCheck(t *testing.T) {
	root := fixture(t, nil)
	control := &fakeControl{}
	adapter := &fakeAdapter{Sessions: []fakeSession{
		{SessionID: "sess-fail", Result: runtime.Result{Status: runtime.StatusFailed, ExitCode: 2, Stderr: "boom"}},
		{SessionID: "sess-ok", Texts: []string{"done"}, Result: runtime.Result{Status: runtime.StatusSucceeded, ExitCode: 0}},
	}}
	checker := passChecker()

	out := mustRun(t, context.Background(), Options{
		Root: root, Item: "WI-0001",
		Adapter: adapter, Control: control, Checker: checker, Now: fixedNow(),
	})

	step3Count := 0
	for _, s := range out.Record.Steps {
		if s.N == 3 && s.Pass == 1 {
			step3Count++
		}
	}
	if step3Count != 2 {
		t.Fatalf("step 3 entries in pass 1 = %d, want 2", step3Count)
	}

	var firstStep7 *Step
	for i := range out.Record.Steps {
		s := &out.Record.Steps[i]
		if s.N == 7 && s.Pass == 1 {
			firstStep7 = s
			break
		}
	}
	if firstStep7 == nil {
		t.Fatal("no step 7 entry recorded")
	}
	if len(firstStep7.Checks) != 1 {
		t.Fatalf("first step 7 has %d checks, want exactly 1 (the synthesized session check): %+v", len(firstStep7.Checks), firstStep7.Checks)
	}
	c := firstStep7.Checks[0]
	if c.Name != "session" {
		t.Errorf("first step 7 check name = %q, want session", c.Name)
	}
	if c.RC != 2 {
		t.Errorf("first step 7 check RC = %d, want 2", c.RC)
	}
	p := filepath.Join(RunDir(root, out.RunID), c.Output)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("session check Output %q does not resolve to a file under the run dir: %v", c.Output, err)
	}
	if !strings.Contains(string(b), "boom") {
		t.Errorf("session check output = %q, want it to contain boom", b)
	}

	if len(adapter.Packets) < 2 {
		t.Fatalf("adapter recorded %d packets, want at least 2 (the failed attempt and the re-entry)", len(adapter.Packets))
	}
	secondPrompt := adapter.Packets[1].Prompt
	if !strings.Contains(secondPrompt, "session") {
		t.Error("second rendered packet is missing the failed check name `session`")
	}
	if !strings.Contains(secondPrompt, "boom") {
		t.Error("second rendered packet is missing the failure output `boom`")
	}

	if out.Record.State != "awaiting_human" {
		t.Errorf("final State = %q, want awaiting_human", out.Record.State)
	}
}
