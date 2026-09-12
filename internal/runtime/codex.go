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
	"strings"
	"sync"
)

// codexBinary is the executable this adapter resolves on PATH.
const codexBinary = "codex"

// codexCapabilities are the role tool tokens this adapter understands on the allow
// list. Unlike Claude's per-tool CLI flags, Codex enforces capability through its
// sandbox mode; the map exists so an unknown token is never silently ignored.
var codexCapabilities = map[string]struct{}{
	"repository.read":   {},
	"repository.write":  {},
	"shell.exec":        {},
	"knowledge.search":  {},
	"knowledge.context": {},
}

// codexDenied are deny tokens this adapter recognizes. Codex has no --disallowedTools
// analogue; github.merge is enforced at the loop's control plane (ISC-30) and by
// keeping network off in workspace-write unless configured otherwise. The map exists
// so an unknown deny token is an error, not a silent pass.
var codexDenied = map[string]struct{}{
	"github.merge": {},
}

// codexNeedsWriteSandbox reports whether the allow list requires write access.
func codexNeedsWriteSandbox(tokens []string) (bool, []string) {
	needsWrite := false
	var unknown []string
	for _, tok := range tokens {
		if _, ok := codexCapabilities[tok]; !ok {
			unknown = append(unknown, tok)
			continue
		}
		switch tok {
		case "repository.write", "shell.exec":
			needsWrite = true
		}
	}
	return needsWrite, unknown
}

// codexUnknownDeny returns any deny token this adapter does not recognize.
func codexUnknownDeny(tokens []string) []string {
	var unknown []string
	for _, tok := range tokens {
		if _, ok := codexDenied[tok]; !ok {
			unknown = append(unknown, tok)
		}
	}
	return unknown
}

// codexSandbox picks the least sandbox mode that satisfies the allow list.
func codexSandbox(needsWrite bool) string {
	if needsWrite {
		return "workspace-write"
	}
	return "read-only"
}

// codexAdapter runs the `codex` CLI in exec mode (ADR-008: subscription CLI, shelled
// out to; no API client and no token handling anywhere in this file).
type codexAdapter struct {
	// binary overrides PATH resolution. Empty means look up codexBinary on PATH.
	binary string
}

func (a *codexAdapter) Start(ctx context.Context, p WorkPacket) (Session, error) {
	binary := a.binary
	if binary == "" {
		resolved, err := exec.LookPath(codexBinary)
		if err != nil {
			return nil, fmt.Errorf("runtime: %s not found on PATH: %w", codexBinary, err)
		}
		binary = resolved
	}

	needsWrite, unknownAllow := codexNeedsWriteSandbox(p.Tools)
	if len(unknownAllow) > 0 {
		return nil, fmt.Errorf("runtime: unknown tool token(s): %s", strings.Join(unknownAllow, ", "))
	}
	unknownDeny := codexUnknownDeny(p.Deny)
	if len(unknownDeny) > 0 {
		return nil, fmt.Errorf("runtime: unknown deny token(s): %s", strings.Join(unknownDeny, ", "))
	}

	args := []string{
		"exec",
		"--json",
		"--sandbox", codexSandbox(needsWrite),
		"--ask-for-approval", "never",
		"-C", p.Worktree,
		p.Prompt,
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = p.Worktree
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
		return nil, fmt.Errorf("runtime: starting %s: %w", codexBinary, err)
	}

	s := &codexSession{
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

// codexSession is one `codex exec` process.
type codexSession struct {
	cmd            *exec.Cmd
	events         chan Event
	transcriptPath string
	stderr         *bytes.Buffer
	scanDone       chan struct{}

	mu             sync.Mutex
	sessionID      string
	status         Status
	lastResultText string

	waitOnce   sync.Once
	waitResult Result
	waitErr    error
}

// codexLine is the subset of codex exec --json events this adapter understands.
// Codex emits newline-delimited JSON with a type tag; see learn.chatgpt.com/docs/
// non-interactive-mode for the stable event vocabulary.
type codexLine struct {
	Type     string `json:"type"`
	ThreadID string `json:"thread_id"`
	Item     *struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Command string `json:"command"`
		Text    string `json:"text"`
	} `json:"item"`
	Message string `json:"message"`
}

func (s *codexSession) scan(stdout io.ReadCloser, transcript *os.File) {
	defer close(s.events)
	defer close(s.scanDone)
	if transcript != nil {
		defer transcript.Close()
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 4<<20)

	for scanner.Scan() {
		line := scanner.Text()
		if transcript != nil {
			fmt.Fprintln(transcript, line)
		}
		s.decode(line)
	}
}

func (s *codexSession) decode(line string) {
	var cl codexLine
	if err := json.Unmarshal([]byte(line), &cl); err != nil {
		return
	}

	sid := s.getSessionID()

	switch cl.Type {
	case "thread.started":
		s.setSessionID(cl.ThreadID)
		s.events <- Event{Kind: EventStarted, SessionID: cl.ThreadID, Raw: line}
	case "item.started":
		if cl.Item == nil {
			return
		}
		tool := cl.Item.Type
		if cl.Item.Type == "command_execution" && cl.Item.Command != "" {
			tool = cl.Item.Command
		}
		s.events <- Event{Kind: EventToolUse, SessionID: sid, Tool: tool, Raw: line}
	case "item.completed":
		if cl.Item == nil {
			return
		}
		if cl.Item.Type == "agent_message" && cl.Item.Text != "" {
			s.events <- Event{Kind: EventText, SessionID: sid, Text: cl.Item.Text, Raw: line}
			s.setLastResultText(cl.Item.Text)
		}
	case "turn.completed":
		sid = s.getSessionID()
		text := s.getLastResultText()
		s.events <- Event{Kind: EventResult, SessionID: sid, Text: text, Raw: line}
	case "turn.failed", "error":
		msg := cl.Message
		if msg == "" && cl.Item != nil && cl.Item.Text != "" {
			msg = cl.Item.Text
		}
		if msg != "" {
			s.setLastResultText(msg)
		}
		sid = s.getSessionID()
		s.events <- Event{Kind: EventResult, SessionID: sid, Text: msg, Raw: line}
	}
}

func (s *codexSession) setSessionID(id string) {
	if id == "" {
		return
	}
	s.mu.Lock()
	s.sessionID = id
	s.mu.Unlock()
}

func (s *codexSession) getSessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionID
}

func (s *codexSession) setLastResultText(text string) {
	s.mu.Lock()
	s.lastResultText = text
	s.mu.Unlock()
}

func (s *codexSession) getLastResultText() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastResultText
}

func (s *codexSession) setStatus(status Status) {
	s.mu.Lock()
	s.status = status
	s.mu.Unlock()
}

func (s *codexSession) ID() string {
	return s.getSessionID()
}

func (s *codexSession) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *codexSession) Events() <-chan Event {
	return s.events
}

func (s *codexSession) Wait() (Result, error) {
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
