# M1 — Jokruger dependency bridge (experimental)

Date: 2026-11-17.

## Scope and provenance

- Project baseline: `badassitron_M1_extended_safety.zip`.
- Supplied `dec128-main.zip`: **main branch snapshot**, ZIP archive identifier `367475f9b33250c59b148d382fec07314ca050d3`, timestamp 2026-10-04. **It is not proof of being tagged v1.4.0.**
- Dependency declared in project `go.mod`: `github.com/jokruger/dec128 v1.4.0`.
- No upstream files have been copied or vendored into Badassitron.
- Testing used a **temporary** local `replace github.com/jokruger/dec128 => ../dec128-main`; this directive is NOT shipped.
- Go executable available locally: Go 1.23.2, not the project-declared Go 1.27. The temporary test workspace lowered both modules' `go` directives to 1.23. The delivered module retains Go 1.27; upstream unchanged. Therefore validation against genuine v1.4.0 / Go 1.27 is PENDING.
- The source files of upstream v1.4+ use `testing.B.Loop` in benchmarks and their own benchmark suite cannot compile under Go 1.23.2. This is a toolchain limitation, not a library defect.

## Implemented (non-public)

`internal/dec128bridge` imports `github.com/jokruger/dec128` and `github.com/jokruger/dec128/{state,uint128}`. `ToUpstream` and `FromUpstream` preserve finite coefficient, scale, and sign via the public constructors; no reflection or memory-layout aliasing. A contextual error translator handles historical errors. `MulStrict` calls `MulRound` with `ROUND_NAN` and explicit scale; upstream globals cannot silently alter this operation.

**Not integrated:** The exported `badassitron/dec128.Dec128` type and all callers remain untouched. This bridge is a proof of compatibility feasibility, NOT a ready-to-ship arithmetic replacement.

Known incompleteness: `MulStrict` currently returns Overflow when the sum of input scales exceeds 19, including cases where exact trailing-zero reduction would allow representation. Upstream-exclusive error codes are temporarily mapped to NotANumber. Neither policy is approved for the public wrapper. Error translation must be finalized per operation in M2.

## Test evidence

- `TestFiniteRoundTrip`: finite values retain coefficient, scale, and sign.
- `TestMulStrictDifferentialSmall`: 3,000 deterministic multiplication cases at scales 0–7, all matched historical string and scale in the local test environment.
- `TestUpstreamGlobalIndependentStrict`: changing the upstream rounding setting does not change the strict result.
- `go test ./... -count=1`: PASS (temporary local test workspace).
- `go vet ./...`: PASS (temporary local test workspace).
- `go test -race ./internal/dec128bridge -count=1`: PASS (temporary local test workspace).

## Preliminary benchmark (local, not release comparison)

Environment: linux/amd64, Intel Xeon Platinum 8573C, Go 1.23.2, `go test -bench '^BenchmarkMul' -benchmem -benchtime=500ms -count=3`, fixed input `1234567.890123 * 987.654321`, 0 allocations per operation. Reported nanoseconds are three repeated samples:

| Method | ns/op samples | Median ns/op | allocs/op |
| --- | --- | ---: | ---: |
| `MulStrict` prototype | 20.21, 20.39, 20.70 | 20.39 | 0 |
| Badassitron legacy `Mul` | 3.318, 3.408, 3.400 | 3.400 | 0 |
| Jokruger direct `MulRound(..., 12, ROUND_NAN)` | 5.702, 5.640, 5.754 | 5.702 | 0 |

Not equivalent production contracts; these measurements isolate dispatch/conversion overhead and must not be interpreted as a complete library performance ranking. **Performance gate: FAIL for this prototype.** Optimize before enabling it in `dec128` or Optimus.

## Next gates

1. Validate source and checksums for actual Go module `v1.4.0`, without using the main ZIP as a substitute for tag verification.
2. Re-run with supported Go 1.27, or at minimum supported upstream Go 1.24, without `go.mod` modifications.
3. Complete exact scale reduction, error-context rules, signed-zero and NaN/serialization compatibility.
4. Run the full 5,243-case differential oracle against an isolated candidate public wrapper, not just the small test corpus.
5. Implement fast paths which bypass the translation adapter for internal arithmetic, benchmark against `FASTPATH_V4` on identical hardware.
6. Only then commit atomic `dec128` implementation switch.
