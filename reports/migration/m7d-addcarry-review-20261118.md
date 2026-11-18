# M7-D — Imported AddCarry optimization, error-path review

Date: 2026-11-18. Status: **local technical candidate; NOT published or provenance-cleared**.

## Goal and retained architecture

Badassitron's public `dec128/uint128.Uint128` remains an independently named
`{Lo, Hi}` type with its original return signatures, error codes, zero-on-error
behavior, and documentation. All 128-bit arithmetic remains delegated to
**imported** `github.com/jokruger/dec128/uint128 v1.4.0`; no upstream source or
historical arithmetic engine is distributed. M7-C previously passed its full
official Go 1.27.1 technical gate. This M7-D source has **not** yet passed its
own Go 1.27.1 gate.

## Implemented optimization — `Uint128.Add`

Old public wrapper:

```go
r, s := ju.Uint128(u).Add(ju.Uint128(other))
return Uint128(r), errors.Error((uint8(s) & 1) * uint8(errors.Overflow))
```

New public wrapper (arithmetical work is still in upstream Jokruger):

```go
r, carry := ju.Uint128(u).AddCarry(ju.Uint128(other))
mask := carry - 1
return Uint128{Lo: r.Lo & mask, Hi: r.Hi & mask},
    errors.Error(uint8(carry) * uint8(errors.Overflow))
```

`AddCarry` is a public imported primitive that returns the modular 128-bit
sum and a carry of **0 or 1**. For carry zero, `carry-1` is all ones, preserving
both limbs. For carry one, it is zero, preserving the historical **zero result
on overflow** without a conditional branch. `carry*3` preserves the historical
`errors.Overflow` code; the public nominal type and its ABI/source signatures
are not changed. This alternative avoids mapping upstream `state.Overflow=5`
after an imported `Add` call. The result still comes entirely from Jokruger's
addition primitive; this is **not** an independent 128-bit addition algorithm.

## Rejected experiments (no changes shipped)

- `QuoRem` zero-divisor precheck: improved the isolated error path in some local
  measurements but raised the Go 1.23.2 wrapper inlining cost from **80 to 97**.
  M7-C's hot success path stays in place.
- `QuoRem256By128` preliminary carry-vs-divisor check: removed inlining; the
  wrapper cost rose from **80 to 125**. It risked slowing successful wide
  division. Upstream has distinct division-by-zero/overflow precedence that
  any shortcut must preserve.
- `FromString` simplified early-return mapping: its short-input benchmark
  was inconsistent and showed no reproducible gain over M7-C. The existing
  Jokruger parser and comprehensive historical error mapping are preserved.

M7-D deliberately does **not** claim a fix for the three remaining official
M7-C overheads: `QuoRemZero` +13.6%, `WideOverflow` +9.6%, and `ParseShort`
+11.9%. Those are priority observations, not release blockers on their own.

## New correctness tests

- `TestM7DAddCarryCompatibility`: **24,000** deterministically seeded and
  boundary cases, comparing the facade against Jokruger's `Add` and confirming
  the independent `AddCarry` contract (carry in 0..1, modular result, overflow
  state and zero-on-error behavior).
- `TestM7DAddIndependentMathBig`: **4,096** independently calculated additions
  using `math/big`, covering exact results and 128-bit overflow.
- `TestM7DWideDivisionErrorPrecedence`: six explicit cases preserving the
  observable difference between zero carry / zero divisor and nonzero carry /
  zero divisor, as well as successful and overflowing wide division.
- Existing 12,000 M7-C upstream comparisons and the 35,000-input codec
  differential tests remain unchanged.

## Local validation (not official Go 1.27.1)

Environment: Go **1.23.2** on a shared Linux amd64 host. Tests were run in a
**scratch copy** with temporary `go 1.23` and a local `replace` to the supplied
upstream source snapshot. The shipped `go.mod` retains `go 1.27`, requires
`github.com/jokruger/dec128 v1.4.0`, and has **no** `replace`, vendor or
upstream Go files.

- Full `go test ./... -count=1`: PASS.
- `go vet ./...`: PASS.
- `go test -race -p 2 ./...`: **interrupted by execution timeout**. Separate
  completed race suites for all numeric packages and for `optimus`, internal
  packages and converters: PASS. This is not a global race PASS.
- Frozen decimal oracles: 3,518 + 1,725 = **5,243** cases, **zero unexpected
  differences**. The primary independent exact check found 389 compared and
  zero mismatches; the extended check found 1,634 examined, 563 skipped,
  38 lossy encodings and **zero mismatches**.
- Go 1.23.2 compiler inserts `Uint128.Add` in line (cost **66**); `QuoRem`
  and `QuoRem256By128` retain their M7-C inlinable costs of **80**.

[Raw local tests](m7d-local/tests-all.txt),
[race results for numeric packages](m7d-local/race-numeric-packages.txt),
[race results for consumers](m7d-local/race-consumer-packages.txt), and
[compiler decisions](m7d-local/inlining-selected.txt) are retained.

## Performance evidence: limited and not interchangeable

The **official M7-C** Go 1.27.1 Ryzen 7 7735HS baseline, supplied after M7-C
was packaged, showed `Add` **1.089 ns/op** versus direct Jokruger
**0.877 ns/op** (**+24.2%**) and `AddOverflow` **0.875 ns/op** versus
**0.880 ns/op** (**−0.6%**). It also showed 10 of 14 measured routes within
5% overhead. The [full official baseline](m7d-prior-official-m7c/comparison.md),
[raw seven-sample evidence](m7d-prior-official-m7c/raw.txt), and the
[previous M7-C verification log](m7d-prior-official-m7c/m7c-verify-go1271.log)
are included.

The local three-sample Go 1.23.2 experiment indicates that using the
imported `AddCarry` may reduce the cost of both normal and overflowing Add:

| Local build | Operation | Facade ns/op | Jokruger `Add` direct ns/op | Overhead |
|---|---|---:|---:|---:|
| M7-C | `Add` | 1.760 | 1.643 | +7.1% |
| M7-D | `Add` | 1.251 | 1.679 | −25.5% |
| M7-C | `AddOverflow` | 1.751 | 1.503 | +16.5% |
| M7-D | `AddOverflow` | 1.308 | 1.611 | −18.8% |

These are short, **noisy and from a different Go version / host**; negative
numbers are not a claim that Badassitron is inherently faster than Jokruger.
Both code paths have 0 B/op and 0 allocs/op. Full local evidence:
[comparison](m7d-local/comparison.md),
[M7-C raw](m7d-local/m7c-add-3x120ms-raw.txt) and
[M7-D raw](m7d-local/m7d-add-3x120ms-raw.txt).

The Go 1.27.1 compiler may generate different code. **Do not infer M7-D's
performance from local tests or claim the 5% target met until the final-tree
host run.** The full real-consumer suite should be benchmarked if the host
shows a meaningful regression.

## Official acceptance procedure

On the **shipped, unmodified** M7-D source with Go 1.27.1 or later and
published Jokruger v1.4.0:

```sh
GOWORK=off bash tools/release/verify.sh
BENCHTIME=250ms COUNT=7 bash tools/release/bench-backend.sh \
  reports/migration/m7d-official-backend
python3 tools/release/compare-iterations.py \
  reports/migration/m7d-prior-official-m7c/raw.txt \
  reports/migration/m7d-official-backend/raw.txt \
  --samples 7 --old-label M7-C --new-label M7-D \
  --output reports/migration/m7d-official-comparison.md
```

Accept `AddCarry` for release only after correctness, race and documented
Go 1.27.1 benchmarks show it is a stable net gain, including `AddOverflow`
and relevant consumers. Otherwise revert the **single** adapter method to
M7-C without touching the rest of the migrated backend.

The source provenance, licensing and redistribution review remains a separate
**publication blocker** regardless of technical performance or test results.
