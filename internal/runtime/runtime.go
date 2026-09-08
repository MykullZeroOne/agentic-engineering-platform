// Package runtime executes a durable agent identity in a transient session.
//
// ADR-002 draws the line this package sits on: the identity is durable and the session
// is not. Nothing here knows which provider answered. A provider token selects an
// adapter from a map in adapters.go, and that map is the only place a provider name
// appears in this package's public behaviour -- adding a second provider is one entry
// in it and no change to anything below.
package runtime

import "context"

// WorkPacket is everything one session needs to do one unit of work.
type WorkPacket struct {
	// RunID is the agent run this session belongs to. Evidence attaches to the run,
	// never to the session (ISC-26), and this is the key that makes that possible.
	RunID string
	// Objective is the session's assignment, in one line. On a return pass it is
	// prefixed "returned: " by the caller; this package does not inspect it.
	Objective string
	// Prompt is the rendered work packet handed to the session verbatim.
	Prompt string
	// Worktree is the absolute directory the session runs in. An adapter never
	// creates it: a session that cannot start must leave no trace that it tried.
	Worktree string
	// Tools are the role's tools.allow tokens, unmapped. Each adapter translates
	// them into whatever its own CLI calls those capabilities.
	Tools []string
	// Deny are the role's tools.deny tokens, unmapped. Each adapter translates
	// them into whatever its own CLI calls to refuse those capabilities.
	Deny []string
	// TranscriptPath, when non-empty, is where the adapter writes every raw line the
	// session emitted. It is the durable copy: an event channel is gone once drained,
	// and a run record must stay readable with no session alive.
	TranscriptPath string
}

// EventKind classifies a session event. Four kinds, deliberately: a vocabulary that
// grows with each provider's output format is not a vocabulary.
type EventKind string

const (
	// EventStarted is emitted once, when the session has an id.
	EventStarted EventKind = "started"
	// EventToolUse is the session invoking a tool.
	EventToolUse EventKind = "tool_use"
	// EventText is the session's own prose.
	EventText EventKind = "text"
	// EventResult is emitted once, last.
	EventResult EventKind = "result"
)

// Event is one observation from a running session.
type Event struct {
	Kind      EventKind
	SessionID string
	// Tool is set on EventToolUse only.
	Tool string
	// Text is set on EventText and EventResult.
	Text string
	// Raw is the undecoded line, so evidence keeps what typing dropped.
	Raw string
}

// Status is a session's lifecycle state, in this package's vocabulary rather than any
// provider's.
type Status string

const (
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

// Result is the outcome of a finished session.
type Result struct {
	// SessionID is the id the underlying tool assigned. Authoritative: Session.ID may
	// still be empty if the session died before announcing one.
	SessionID string
	Status    Status
	// ExitCode is the process's exit status. Recorded, never inferred.
	ExitCode int
	// Stderr is everything the process wrote to standard error. A failure whose output
	// is discarded is a failure nobody can diagnose.
	Stderr string
	// Text is the session's final message.
	Text string
	// TranscriptPath is where the raw lines were written, empty if the caller asked
	// for none.
	TranscriptPath string
}

// Session is one transient execution of an agent identity.
type Session interface {
	// ID is the underlying tool's session id, empty until the session announces one.
	ID() string
	// Status is the session's current state.
	Status() Status
	// Events is closed when the session ends. A caller must drain it; an undrained
	// channel stalls the session.
	Events() <-chan Event
	// Wait blocks until the session ends and returns its Result.
	//
	// A non-zero exit is DATA, not an error: Wait returns a Result with Status
	// StatusFailed, the exit code, and the captured stderr, and a nil error. Wait
	// returns a non-nil error only when the session could not be observed at all --
	// a pipe that could not be opened, a context cancelled before the process ran.
	Wait() (Result, error)
}

// Adapter starts sessions. One implementation per provider token.
type Adapter interface {
	Start(ctx context.Context, p WorkPacket) (Session, error)
}
