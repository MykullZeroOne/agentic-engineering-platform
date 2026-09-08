// Command devctl is the AEP command line.
//
// Phase 1 of docs/roadmap/IMPLEMENTATION_ROADMAP.md, and the first product code in this
// repository. Only `doctor` exists: the roadmap also lists `init` and `adopt`, and
// stubbing them now would commit to a command shape before any of it has been used.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/approvals"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/doctor"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/loop"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/work"
)

const usage = `devctl - the Agentic Engineering Platform command line

usage:
  devctl doctor [--root DIR] [--quiet]
  devctl work list [--state S] [--type T] [--priority P] [--root DIR]
  devctl work show WI-NNNN [--root DIR]
  devctl work advance [--dry-run] [--ref REF] [--root DIR]
  devctl run WI-NNNN [--runtime TOKEN] [--role ID] [--root DIR]

commands:
  doctor   Report whether a project's .agentic/ configuration is coherent.
           Read-only. Exits 1 when a finding would break a runtime.
  work     Read the local work store, and derive the state of merged items.
           list and show are read-only. advance writes, and is the only command
           that does -- it executes the work.advance_state rule from
           .agentic/hooks/hooks.yaml, which has been bound and inert since
           WI-0013. --dry-run reports drift without writing and exits 1 when
           there is any, which is the shape a CI check needs.
  run      Execute ADR-023's universal agent loop for one work item, resuming
           an open run when one exists. Parks at step 7 on a human-owned gate:
           the pull request it opens is the next thing a human acts on. Exit
           codes: 0 parked or complete, 1 unknown work item, 2 a configuration
           error, 3 blocked on a question a human must answer.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "doctor":
		os.Exit(runDoctor(os.Args[2:]))
	case "work":
		os.Exit(runWork(os.Args[2:]))
	case "run":
		os.Exit(runRun(os.Args[2:]))
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "devctl: unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}

func runWork(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, "devctl work: expected `list`, `show` or `advance`\n\n"+usage)
		return 2
	}
	sub, rest := args[0], args[1:]

	fs := flag.NewFlagSet("work "+sub, flag.ExitOnError)
	root := fs.String("root", ".", "project root: the directory holding .agentic/")
	var f work.Filter
	if sub == "list" {
		fs.StringVar(&f.State, "state", "", "only this work_state")
		fs.StringVar(&f.Type, "type", "", "only this type")
		fs.StringVar(&f.Priority, "priority", "", "only this priority")
	}
	var dryRun *bool
	var ref *string
	if sub == "advance" {
		dryRun = fs.Bool("dry-run", false, "report drift without writing; exit 1 if any")
		ref = fs.String("ref", "origin/main", "the trunk ref whose merge commits are read")
	}

	switch sub {
	case "list":
		_ = fs.Parse(rest)
		items, err := work.List(config.Root(*root), f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "devctl work list: %v\n", err)
			return 2
		}
		fmt.Print(work.FormatList(items))
	case "show":
		if len(rest) == 0 || strings.HasPrefix(rest[0], "-") {
			fmt.Fprint(os.Stderr, "devctl work show: expected a work item id, e.g. WI-0001\n")
			return 2
		}
		id := rest[0]
		_ = fs.Parse(rest[1:])
		w, err := work.Find(config.Root(*root), id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "devctl work show: %v\n", err)
			return 1
		}
		fmt.Print(work.FormatItem(w))
	case "advance":
		_ = fs.Parse(rest)
		return runAdvance(*root, *ref, *dryRun)
	default:
		fmt.Fprintf(os.Stderr, "devctl work: unknown subcommand %q\n\n%s", sub, usage)
		return 2
	}
	return 0
}

// runAdvance executes the work.advance_state rule.
//
// Exit codes carry meaning, because this is meant to be usable as a check: --dry-run
// exits 1 on drift so CI can fail on a stale store, while a write exits 0 because it
// fixed what it found. Both exit 2 on an error, which is a different thing from drift.
func runAdvance(root, ref string, dryRun bool) int {
	items, err := config.LoadWorkItems(config.Root(root))
	if err != nil {
		fmt.Fprintf(os.Stderr, "devctl work advance: %v\n", err)
		return 2
	}
	merged, err := work.MergedItems(root, ref, items, work.GH{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "devctl work advance: %v\n", err)
		return 2
	}
	records, err := approvals.Load(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "devctl work advance: %v\n", err)
		return 2
	}

	changes := work.Plan(root, items, merged, records)
	fmt.Print(work.FormatPlan(changes, !dryRun && len(changes) > 0))

	// Which items have no merge record at all. Named rather than left implicit: an item
	// this command never considered looks identical, in its output, to one it considered
	// and found correct. PRs #8 and #9 merged without a Closes trailer, so their items
	// are permanently in this set and nothing here can derive their state.
	var unmerged []string
	for _, it := range items {
		if _, ok := merged[it.ID]; !ok && !isTerminal(it.WorkState) {
			unmerged = append(unmerged, it.ID)
		}
	}
	if len(unmerged) > 0 {
		fmt.Printf("\n%d item(s) have no `Closes` trailer on %s and were not evaluated: %s\n",
			len(unmerged), ref, strings.Join(unmerged, " "))
	}

	if len(changes) == 0 {
		return 0
	}
	if dryRun {
		return 1
	}
	if err := work.Apply(root, changes); err != nil {
		fmt.Fprintf(os.Stderr, "devctl work advance: %v\n", err)
		return 2
	}
	return 0
}

func isTerminal(s string) bool { return s == "done" || s == "cancelled" }

// runRun executes the universal agent loop for one work item.
//
// Exit codes carry meaning, as they do for `work advance`: 0 parked or complete,
// 1 the work item does not exist, 2 a configuration error the run cannot proceed
// through, 3 the run is blocked on a question a human must answer.
func runRun(args []string) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fmt.Fprint(os.Stderr, "devctl run: expected a work item id, e.g. WI-0001\n")
		return 2
	}
	item := args[0]

	fs := flag.NewFlagSet("run", flag.ExitOnError)
	root := fs.String("root", ".", "project root: the directory holding .agentic/")
	runtimeToken := fs.String("runtime", "", "override the resolved provider (project.yaml's runtime_preferences by default)")
	roleID := fs.String("role", "", "role id (default engineer.primary)")
	_ = fs.Parse(args[1:])

	out, err := loop.Run(context.Background(), loop.Options{
		Root:     config.Root(*root),
		Item:     item,
		Provider: *runtimeToken,
		RoleID:   *roleID,
	})
	if err != nil {
		switch {
		case errors.Is(err, loop.ErrUnknownItem):
			fmt.Fprintf(os.Stderr, "devctl run: %v\n", err)
			return 1
		case errors.Is(err, loop.ErrNoAdapter):
			fmt.Fprintf(os.Stderr, "devctl run: %v\n", err)
			return 2
		default:
			fmt.Fprintf(os.Stderr, "devctl run: %v\n", err)
			return 2
		}
	}
	return out.Exit
}

func runDoctor(args []string) int {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	root := fs.String("root", ".", "project root: the directory holding .agentic/")
	quiet := fs.Bool("quiet", false, "print findings only")
	_ = fs.Parse(args)

	rep, err := doctor.Run(config.Root(*root))
	if err != nil {
		fmt.Fprintf(os.Stderr, "devctl doctor: %v\n", err)
		return 2
	}

	for _, f := range rep.Findings {
		fmt.Println(f)
	}

	// A clean report says what it looked at. "OK" alone is indistinguishable from a
	// check that silently examined nothing, which is the failure mode worth guarding.
	if !*quiet {
		if len(rep.Findings) > 0 {
			fmt.Println()
		}
		keys := make([]string, 0, len(rep.Checked))
		for k := range rep.Checked {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := ""
		for i, k := range keys {
			if i > 0 {
				parts += ", "
			}
			parts += fmt.Sprintf("%d %s", rep.Checked[k], k)
		}
		if parts == "" {
			parts = "nothing"
		}
		if n := rep.Errors(); n > 0 {
			fmt.Printf("%d error(s) across %s.\n", n, parts)
		} else {
			fmt.Printf("OK: %s.\n", parts)
		}
	}

	if rep.Errors() > 0 {
		return 1
	}
	return 0
}
