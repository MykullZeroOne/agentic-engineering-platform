#!/usr/bin/env python3
"""Tests for scripts/validate_docs.py.

Runnable with no test dependency beyond the standard library:

    python3 scripts/test_validate_docs.py

Covers the ISA claim polarity marker set (WI-0066). That check was verified once by
hand -- a temporary `anti:` and `Negative:` were made to error, then reverted -- and
nothing kept it true afterwards. These cases keep it true.

Each case runs the real validator end to end against a fixture root: a temporary
directory holding a copy of the script and symlinks to everything else in the
repository, with ISA.md replaced by the real one plus one appended claim. Copying the
script rather than symlinking it matters -- ROOT is `__file__.resolve().parent.parent`,
and resolving a symlink would land back on the repository and read the real ISA.md.
"""
from __future__ import annotations

import pathlib
import shutil
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
VALIDATOR = ROOT / "scripts" / "validate_docs.py"
CLAIM_ID = "ISC-9001"


def run_with_claim(claim_text: str) -> subprocess.CompletedProcess:
    """Run the validator against a fixture root whose ISA.md carries one extra claim."""
    with tempfile.TemporaryDirectory() as tmp:
        fixture = pathlib.Path(tmp) / "repo"
        (fixture / "scripts").mkdir(parents=True)
        shutil.copy2(VALIDATOR, fixture / "scripts" / VALIDATOR.name)

        for entry in ROOT.iterdir():
            if entry.name in {"ISA.md", "scripts", ".git"}:
                continue
            (fixture / entry.name).symlink_to(entry, target_is_directory=entry.is_dir())

        isa = (ROOT / "ISA.md").read_text()
        (fixture / "ISA.md").write_text(
            f"{isa}\n- [ ] {CLAIM_ID}: {claim_text}\n"
        )

        return subprocess.run(
            [sys.executable, str(fixture / "scripts" / VALIDATOR.name)],
            capture_output=True, text=True,
        )


class IsaPolarityMarkers(unittest.TestCase):
    """ISA.md's Language section names a closed marker set; the validator enforces it."""

    def test_repository_is_clean(self):
        """Guard: the fixture only means something if the real corpus validates."""
        result = subprocess.run(
            [sys.executable, str(VALIDATOR)], capture_output=True, text=True,
        )
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_known_markers_and_no_marker_pass(self):
        for text in (
            "a claim with no marker, which is direct.",
            "Anti: something that must not happen.",
            "Advisory: a standing check.",
            "Antecedent: a precondition another claim reads through.",
        ):
            with self.subTest(claim=text):
                result = run_with_claim(text)
                self.assertEqual(
                    result.returncode, 0, result.stdout + result.stderr
                )

    def test_unknown_marker_fails_and_names_the_claim(self):
        for text in (
            "Negative: a marker nobody decided on.",
            "anti: the right word in the wrong case.",
            "Warning: plausible, and still not in the set.",
        ):
            with self.subTest(claim=text):
                result = run_with_claim(text)
                self.assertEqual(
                    result.returncode, 1, result.stdout + result.stderr
                )
                marker = text.split(":", 1)[0]
                self.assertIn(
                    f"ISA.md: {CLAIM_ID} opens with '{marker}:'", result.stdout
                )


if __name__ == "__main__":
    unittest.main()
