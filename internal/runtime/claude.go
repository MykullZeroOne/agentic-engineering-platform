package runtime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
)

// claudeBinary is the executable this adapter resolves on PATH.
const claudeBinary = "claude"

// permissionMode is passed to the CLI so a session started with no terminal can
// actually edit files in its worktree. See the plan's Questions section: ISA decision
// 3 does not list this flag, and it is added here because without it the dogfood
// session cannot write. If the helm rules against it, delete this constant and the
// flag; nothing else changes.
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
	// Resolve the binary first, before touching the filesystem: a session that
	// cannot start must leave no trace that it tried (C17).
	binary := a.binary
	if binary == "" {
		resolved, err := exec.LookPath(claudeBinary)
		if err != nil {
			return nil, fmt.Errorf("runtime: %s not found on PATH: %w", claudeBinary, err)
		}
		binary = resolved
	}

	allowed, unknownAllow := mapTools(p.Tools)
	if len(unknownAllow) > 0 {
		return nil, fmt.Errorf("runtime: unknown tool token(s): %s", strings.Join(unknownAllow, ", "))
	}
	denied, unknownDeny := mapDeny(p.Deny)
	if len(unknownDeny) > 0 {
		return nil, fmt.Errorf("runtime: unknown deny token(s): %s", strings.Join(unknownDeny, ", "))
	}

	args := []string{
		"-p", p.Prompt,
		"--output-format", "stream-json",
		"--verbose",
		"--allowedTools", strings.Join(allowed, ","),
	}
	if len(denied) > 0 {
		args = append(args, "--disallowedTools", strings.Join(denied, ","))
	}
	args = append(args, "--permission-mode", permissionMode)

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = p.Worktree
	// A resumed run must read no stdin (C28); nil gives the child /dev/null.
	cmd.Stdin = nil

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("runtime: stdout pipe: %w", err)
	}
	stderrBuf := &bytes.Buffer{}
	cmd.Stderr = stderrBuf

	var transcript *os.File
	if p.TranscriptPath != "" {
		f, err := os.Create(p.TranscriptPath)
		if err != nil {
			return nil, fmt.Errorf("runtime: creating transcript file: %w", err)
		}
		transcript = f
	}

	if err := cmd.Start(); err != nil {
		if transcript != nil {
			transcript.Close()
		}
		return nil, fmt.Errorf("runtime: starting %s: %w", claudeBinary, err)
	}

	s := &claudeSession{
		cmd:            cmd,
		events:         make(chan Event),
		transcriptPath: p.TranscriptPath,
		status:         StatusRunning,
		stderr:         stderrBuf,
		scanDone:       make(chan struct{}),
	}
	go s.scan(stdout, transcript)

	return s, nil
}

// claudeSession is one `claude -p` process.
type claudeSession struct {
	cmd            *exec.Cmd
	events         chan Event
	transcriptPath string
	stderr         *bytes.Buffer

	// scanDone is closed once the stdout scanner has hit EOF, meaning cmd.Wait can
	// safely be called and the process has produced everything it will produce.
	scanDone chan struct{}

	mu             sync.Mutex
	sessionID      string
	status         Status
	lastResultText string

	waitOnce   sync.Once
	waitResult Result
	waitErr    error
}

// streamLine is the shape of one stream-json line this adapter understands.
type streamLine struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype"`
	SessionID string `json:"session_id"`
	Message   *struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
			Name string `json:"name"`
		} `json:"content"`
	} `json:"message"`
	Result string `json:"result"`
}

// scan reads the child's stdout line by line, writes each raw line to the transcript
// verbatim when one is configured, decodes it, and pushes typed events. It closes the
// events channel and scanDone when the child's stdout reaches EOF.
func (s *claudeSession) scan(stdout io.ReadCloser, transcript *os.File) {
	defer close(s.events)
	defer close(s.scanDone)
	if transcript != nil {
		defer transcript.Close()
	}

	scanner := bufio.NewScanner(stdout)
	// stream-json lines carry whole tool results; the 64 KiB default scanner buffer
	// truncates them.
	scanner.Buffer(make([]byte, 0, 64*1024), 4<<20)

	for scanner.Scan() {
		line := scanner.Text()
		if transcript != nil {
			fmt.Fprintln(transcript, line)
		}
		s.decode(line)
	}
}

// decode turns one raw stream-json line into zero or more events. A line that does
// not parse as JSON is not an error: the CLI writes non-JSON diagnostics on occasion,
// and a decoder that dies on one loses the whole session.
func (s *claudeSession) decode(line string) {
	var sl streamLine
	if err := json.Unmarshal([]byte(line), &sl); err != nil {
		return
	}

	switch sl.Type {
	case "system":
		if sl.Subtype == "init" {
			s.setSessionID(sl.SessionID)
			s.events <- Event{Kind: EventStarted, SessionID: sl.SessionID, Raw: line}
		}
	case "assistant":
		if sl.Message == nil {
			return
		}
		sid := s.getSessionID()
		for _, c := range sl.Message.Content {
			switch c.Type {
			case "text":
				s.events <- Event{Kind: EventText, SessionID: sid, Text: c.Text, Raw: line}
			case "tool_use":
				s.events <- Event{Kind: EventToolUse, SessionID: sid, Tool: c.Name, Raw: line}
			}
		}
	case "user":
		// Tool results: written to the transcript above, no event.
	case "result":
		sid := sl.SessionID
		if sid == "" {
			sid = s.getSessionID()
		}
		s.setSessionID(sid)
		s.setLastResultText(sl.Result)
		s.events <- Event{Kind: EventResult, SessionID: sid, Text: sl.Result, Raw: line}
	}
}

func (s *claudeSession) setSessionID(id string) {
	if id == "" {
		return
	}
	s.mu.Lock()
	s.sessionID = id
	s.mu.Unlock()
}

func (s *claudeSession) getSessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionID
}

func (s *claudeSession) setLastResultText(text string) {
	s.mu.Lock()
	s.lastResultText = text
	s.mu.Unlock()
}

func (s *claudeSession) getLastResultText() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastResultText
}

func (s *claudeSession) setStatus(status Status) {
	s.mu.Lock()
	s.status = status
	s.mu.Unlock()
}

func (s *claudeSession) ID() string {
	return s.getSessionID()
}

func (s *claudeSession) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *claudeSession) Events() <-chan Event {
	return s.events
}

// Wait blocks until the stdout scanner has drained (see scanDone) and the process has
// exited, then builds and caches the Result. A non-zero exit is data on the Result,
// never a Go error; Wait returns a non-nil error only when the session could not be
// observed at all.
func (s *claudeSession) Wait() (Result, error) {
	s.waitOnce.Do(func() {
		<-s.scanDone
		err := s.cmd.Wait()

		exitCode := 0
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				exitCode = exitErr.ExitCode()
			} else {
				s.waitErr = fmt.Errorf("runtime: wait: %w", err)
				return
			}
		}

		status := StatusSucceeded
		if exitCode != 0 {
			status = StatusFailed
		}
		s.setStatus(status)

		s.waitResult = Result{
			SessionID:      s.getSessionID(),
			Status:         status,
			ExitCode:       exitCode,
			Stderr:         s.stderr.String(),
			Text:           s.getLastResultText(),
			TranscriptPath: s.transcriptPath,
		}
	})
	return s.waitResult, s.waitErr
}
