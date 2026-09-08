package loop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
)

// PRState is a pull request's state in the control plane's own vocabulary.
type PRState string

const (
	PROpen   PRState = "OPEN"
	PRMerged PRState = "MERGED"
	PRClosed PRState = "CLOSED"
)

// PR is what the loop needs to know about a pull request it opened.
type PR struct {
	URL   string
	State PRState
	// Review is the latest review body or closing comment, which becomes the
	// objective of the next pass when a pull request is returned (ADR-023).
	Review string
	// MergeCommit is set when State is PRMerged.
	MergeCommit string
}

// Control is the loop's port onto the source-control plane.
//
// It has no Merge method, and the absence is the point. ADR-019 says merging closes no
// gate; POL-001 M5 says an agent never merges a pull request that crosses one; the
// ISA's C2 says the loop never invokes a merge by any path. An interface with no such
// method cannot be talked into one, and that is a stronger guarantee than a rule.
type Control interface {
	// Worktree creates a git worktree at dir, on a new branch cut from base.
	Worktree(repo, dir, branch, base string) error
	// RemoveWorktree removes one.
	RemoveWorktree(repo, dir string) error
	// Status returns the paths changed in the worktree at dir.
	Status(dir string) ([]string, error)
	// Commit stages everything under dir and commits it, returning the head sha.
	Commit(dir, msg string) (string, error)
	// Push publishes branch from dir.
	Push(dir, branch string) error
	// CreatePR opens a pull request and returns its URL.
	CreatePR(dir, branch, title, body string) (string, error)
	// ViewPR reports the current state of the pull request at url.
	ViewPR(dir, url string) (PR, error)
}

// GHControl drives git and the gh CLI.
//
// Every external invocation goes through one method, run, so there is exactly one
// place an argv can be built and exactly one place to look when asking what this type
// is capable of running. Log, when non-nil, receives every argv before it executes;
// the tests use it, and so may an operator debugging a run.
type GHControl struct {
	Log func(argv []string)
}

func (g GHControl) run(dir, name string, args ...string) (string, error) {
	argv := append([]string{name}, args...)
	if g.Log != nil {
		g.Log(argv)
	}
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%s: %w: %s", strings.Join(argv, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func (g GHControl) Worktree(repo, dir, branch, base string) error {
	_, err := g.run("", "git", "-C", repo, "worktree", "add", "-b", branch, dir, base)
	return err
}

func (g GHControl) RemoveWorktree(repo, dir string) error {
	_, err := g.run("", "git", "-C", repo, "worktree", "remove", "--force", dir)
	return err
}

func (g GHControl) Status(dir string) ([]string, error) {
	out, err := g.run(dir, "git", "-C", dir, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" || len(line) < 4 {
			continue
		}
		p := strings.TrimSpace(line[3:])
		if idx := strings.Index(p, " -> "); idx >= 0 {
			p = p[idx+len(" -> "):]
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func (g GHControl) Commit(dir, msg string) (string, error) {
	if _, err := g.run(dir, "git", "-C", dir, "add", "-A"); err != nil {
		return "", err
	}
	if _, err := g.run(dir, "git", "-C", dir, "commit", "-m", msg); err != nil {
		return "", err
	}
	out, err := g.run(dir, "git", "-C", dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (g GHControl) Push(dir, branch string) error {
	_, err := g.run(dir, "git", "-C", dir, "push", "-u", "origin", branch)
	return err
}

// CreatePR opens the pull request via `gh pr create`, with the body written to a temp
// file rather than argv: a PR body carrying a whole run record does not belong in
// argv. gh resolves the repository from the worktree's own git remote, so no --repo
// flag is built here; Control's CreatePR carries no repo parameter to build one from.
func (g GHControl) CreatePR(dir, branch, title, body string) (string, error) {
	// The temp file is deliberately not removed here: it is `gh`'s own argument, and a
	// caller or test may still want to read it back after CreatePR returns. It is a
	// handful of bytes in the OS temp directory, cleaned up by the OS in due course.
	tmp, err := os.CreateTemp("", "pr-body-*.md")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.WriteString(body); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	out, err := g.run(dir, "gh", "pr", "create", "--head", branch, "--title", title, "--body-file", tmpPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (g GHControl) ViewPR(dir, url string) (PR, error) {
	out, err := g.run(dir, "gh", "pr", "view", url, "--json", "state,url,mergeCommit,reviews,comments")
	if err != nil {
		return PR{}, err
	}
	var resp struct {
		State       string `json:"state"`
		URL         string `json:"url"`
		MergeCommit *struct {
			Oid string `json:"oid"`
		} `json:"mergeCommit"`
		Reviews []struct {
			Body string `json:"body"`
		} `json:"reviews"`
		Comments []struct {
			Body string `json:"body"`
		} `json:"comments"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return PR{}, fmt.Errorf("parse gh pr view output: %w", err)
	}
	pr := PR{URL: resp.URL, State: PRState(resp.State)}
	if resp.MergeCommit != nil {
		pr.MergeCommit = resp.MergeCommit.Oid
	}
	switch {
	case len(resp.Reviews) > 0:
		pr.Review = resp.Reviews[len(resp.Reviews)-1].Body
	case len(resp.Comments) > 0:
		pr.Review = resp.Comments[len(resp.Comments)-1].Body
	}
	return pr, nil
}

var nonAlnumRE = regexp.MustCompile(`[^a-z0-9]+`)

// kebab lowercases title, replaces runs of non-alphanumerics with a single hyphen, and
// truncates at a word boundary near maxLen characters.
func kebab(title string, maxLen int) string {
	s := nonAlnumRE.ReplaceAllString(strings.ToLower(title), "-")
	s = strings.Trim(s, "-")
	if len(s) <= maxLen {
		return s
	}
	cut := s[:maxLen]
	if idx := strings.LastIndex(cut, "-"); idx > 0 {
		cut = cut[:idx]
	}
	return strings.Trim(cut, "-")
}

// itemNumber strips the work item id down to its digits, with leading zeros removed:
// WI-0007 gives 7.
func itemNumber(id string) string {
	digits := digitsRE.FindString(id)
	if digits == "" {
		return id
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return digits
	}
	return strconv.Itoa(n)
}

// BranchName builds a POL-001 branch name from a work item: <type>/<number>-<summary>,
// the number taken from the item id and the summary kebab-cased from its title and
// truncated at a word boundary near 48 characters.
func BranchName(w config.WorkItem) string {
	return fmt.Sprintf("%s/%s-%s", w.Type, itemNumber(w.ID), kebab(w.Title, 48))
}

// commitSubject is the Conventional Commit subject line: <type>: <title>, the title
// kept exactly as written. It is also the PR title (POL-001 makes the title the
// squash subject). When the line would exceed 72 bytes it is cut at the last space
// before byte 72, with no ellipsis: a truncated commit subject is still meant to read
// as a sentence fragment, not a label that says it was cut short.
func commitSubject(w config.WorkItem) string {
	subj := fmt.Sprintf("%s: %s", w.Type, w.Title)
	if len(subj) <= 72 {
		return subj
	}
	cut := subj[:72]
	if idx := strings.LastIndex(cut, " "); idx > 0 {
		cut = cut[:idx]
	}
	return strings.TrimRight(cut, " ")
}

// CommitMessage builds the branch commit message. The `Closes WI-NNNN` trailer lives
// here, in the BRANCH commit, because GitHub's squash body defaults to the branch's
// commit messages and PRs #8 and #9 both reached main unlinked by relying on the body.
func CommitMessage(w config.WorkItem, objective string) string {
	return fmt.Sprintf("%s\n\n%s\n\nCloses %s", commitSubject(w), objective, w.ID)
}
