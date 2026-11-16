# Optimus v3 — Phase A execution evidence

All files in this directory are recorded outputs from the verification
performed on 2026-11-12. The project was evaluated with Go 1.23.2 and a
separate, temporary Go 1.23-compatible modfile. The original repository's
Go 1.24 module declaration is untouched.

- `before_*`: unmodified Optimus v2 source, except that the regression test
  was copied temporarily into the test directory for the explicit failure
  reproduction and rare-case benchmark.
- `before_reproduction_failure.txt`: **expected FAIL**, proving the original
  implementation returned a numerically incorrect answer in a specific case.
- `before_global_test.txt` / `global.txt`: unchanged pre-existing compile
  problem in `withfloat64/internal/math.go`.
- `api_original.txt` / `api_updated.txt`: byte-for-byte equal `go doc -all`
  output for the public `optimus` package.
- `core_test.txt`, `race.txt`, `vet.txt`: post-fix PASS evidence.
- `benchmarks.txt`: post-fix performance, including the focused rare path.
- `fuzz.txt`: post-fix native Go fuzzing, no reported failures.

`OPTIMUS_V3_AUDIT.md` contains the risk matrix, mathematical justification,
commands, interpretation and limitations.
