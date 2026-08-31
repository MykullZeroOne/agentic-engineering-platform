// Command devctl is the AEP command line.
//
// Phase 1 of docs/roadmap/IMPLEMENTATION_ROADMAP.md, and the first product code in this
// repository. Only `doctor` exists: the roadmap also lists `init` and `adopt`, and
// stubbing them now would commit to a command shape before any of it has been used.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/MykullZeroOne/agentic-engineering-platform/internal/config"
	"github.com/MykullZeroOne/agentic-engineering-platform/internal/doctor"
)

const usage = `devctl - the Agentic Engineering Platform command line

usage:
  devctl doctor [--root DIR] [--quiet]

commands:
  doctor   Report whether a project's .agentic/ configuration is coherent.
           Read-only. Exits 1 when a finding would break a runtime.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "doctor":
		os.Exit(runDoctor(os.Args[2:]))
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "devctl: unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
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
