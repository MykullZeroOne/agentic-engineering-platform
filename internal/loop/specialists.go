package loop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// specialistReport is one advisory preflight report written at step 3 for the BA
// lead to synthesize (ADR-005). Specialists never speak to the human directly.
type specialistReport struct {
	Role    string
	Path    string
	Summary string
}

// runSpecialistPreflight writes compliance and legal advisory reports under
// specialists/ in the run directory. Deterministic stubs until WI-0050 grows real
// specialist roles; the routing contract is what this slice proves (ISC-11).
func runSpecialistPreflight(runDir string, idea string) ([]specialistReport, error) {
	dir := filepath.Join(runDir, "specialists")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	reports := []struct {
		role, filename, body string
	}{
		{
			"compliance.primary",
			"compliance.md",
			fmt.Sprintf("# Compliance preflight\n\nAdvisory report for the BA lead.\n\n## Idea under review\n\n%s\n\n## Findings\n\n- No approved policy artifact governs this idea yet; treat as draft intent only.\n- Route material compliance questions to the human before the product_spec gate closes.\n", idea),
		},
		{
			"legal.preflight",
			"legal.md",
			fmt.Sprintf("# Legal preflight\n\nAdvisory report for the BA lead. Not legal approval.\n\n## Idea under review\n\n%s\n\n## Findings\n\n- No legal review requested; flag export, privacy, and terms-of-use unknowns for the human.\n", idea),
		},
	}
	out := make([]specialistReport, 0, len(reports))
	for _, r := range reports {
		rel := filepath.Join("specialists", r.filename)
		abs := filepath.Join(runDir, rel)
		if err := os.WriteFile(abs, []byte(r.body), 0o644); err != nil {
			return nil, err
		}
		first := strings.SplitN(strings.TrimSpace(r.body), "\n", 2)[0]
		out = append(out, specialistReport{Role: r.role, Path: rel, Summary: first})
	}
	return out, nil
}
