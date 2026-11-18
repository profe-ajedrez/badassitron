# M1 hardening and M2-CANDIDATE-A (not a public cutover)

**Date:** 2026-11-17. **Authority:** the preceding accumulated artifact
`badassitron_M1_parser_signed_zero_20261117.zip`, not an unverified GitHub
snapshot. The upstream source archive was consulted only as an isolated local
Go dependency for testing. **No Jokruger source is distributed or vendored.**

## Implemented in this checkpoint

1. Removed all remaining calls to *historical arithmetic* from the normal
   bridge fallback paths. `canonicalOperand` now strips exact fractional
   zeros with Jokruger's exported `uint128.QuoRem64` instead of calling the
   historical `Dec128.Canonical`. Add/Sub/Mul/Div fallback paths share it.
   Public legacy types are still used at the bridge's interop boundary.
2. Added the **additive** `dec128.NewStoredZero(scale, negative)` constructor.
   The existing `New(0, 0, true)` continues to normalize to positive zero.
   The new constructor permits the historical raw negative-zero/scale-zero
   state *without calling* legacy `Rescale(0)` from the bridge.
3. Extended `DivAtPrecision` to preserve a raw negative-zero quotient when
   precision is zero. Added regression tests for the bridge and public
   constructor. Codec and SQL paths remain independent of historical codecs.
4. Built `internal/dec128candidate` as an **isolated full-declaration-surface
   candidate**: historical `Dec128` methods/functions, values, constructors,
   signed-zero/scale/error state, parsing, JSON/text/SQL, exact Add/Sub/Mul,
   explicit arithmetic rounding with inexact status, truncated Div, Mod,
   QuoRem, PowInt, sqrt, conversions, and mathematical utilities.
   The candidate does **not import `badassitron/dec128` itself**, only the
   historical public `errors` and `uint128` subpackages. It calls the external
   Jokruger dependency and its unsigned primitives, with explicit modes and
   no dependence on Jokruger's process-wide arithmetic settings.
5. Matched historical `Round`/`RoundBank` semantics separately from
   `RescaleRound`: requesting a scale not lower than the stored one is a
   no-op, even if that request is greater than 19. Round-family operations
   also preserve their raw zero sign, unlike quantization of a nonzero
   negative magnitude by `RescaleRoundWithStatus`.
6. Added repeatable candidate oracle and exported surface tools, seeded
   arithmetic and sqrt differentials, stored-zero matrix, concrete Go interface
   assertions and SQL scan client fixture, and upstream-policy independence
   tests. Candidate `Dec128` stores a compact 128-bit coefficient,
   scale and 16-bit state; common equal-scale Add/Sub and 64-bit Mul have
   direct exact fast paths. These are local algorithms and imported primitives,
   not copied upstream source.

**Rejected experiment:** Delegating every M1 bridge operation through the
candidate passed the functional tests, but made the bridge microbenchmarks
**54–138% slower** on sampled operations. That change was **fully reverted**.
Separate M1 fast paths and M2 candidate core are retained during validation;
this temporary duplication must be eliminated only after the atomic public
cutover and a measured post-cutover review. See the rejected-experiment
observations in the performance section below.

## Validation and observed status

Environment: `linux/amd64`, AMD EPYC 9V74 (5 logical CPUs visible),
**Go 1.23.2**. Local tests temporarily changed the working `go.mod` to
`go 1.23` with `replace github.com/jokruger/dec128 => ../testdeps/dec128-main`;
the archived `go.mod` is restored to **Go 1.27**, official `require v1.4.0`,
**no `replace`**. The upstream `main` snapshot is **not authenticated as the
v1.4.0 tag**. The published module and Go 1.27 toolchain were unavailable
because external module downloads failed. The included `go.sum` is still
empty; **official build reproducibility is blocked**.

| Gate | Checkpoint result | Local evidence |
| --- | --- | --- |
| M2 primary 3,518-case oracle | **PASS**, no differences | [primary compare](m2-candidate-compare-primary-20261117.txt); [candidate outputs](m2-candidate-primary-results-20261117.jsonl) |
| M2 extended 1,725-case oracle | **PASS**, no differences | [extended compare](m2-candidate-compare-extended-20261117.txt); [outputs](m2-candidate-extended-results-20261117.jsonl) |
| Independent primary mathematical checks | **PASS**, 389 checked, 159 skipped, zero mismatches | [exact primary](m2-candidate-exact-primary-20261117.txt) |
| Independent extended checks | **PASS**, 1,634 checked, 563 skipped, 38 lossy encodes, zero mismatches | [exact extended](m2-candidate-exact-extended-20261117.txt) |
| Exported AST surface | **PASS**, 83 original methods/functions, 37 constants/variables, 2 type names, zero missing or signature-mismatched; two extra candidate methods | [surface comparison](m2-candidate-surface-20261117.txt) |
| Public legacy exported API vs frozen M0 | **No removals or signature changes**; prior additive `StoredNegative`, new additive `NewStoredZero` | AST inventory generated locally |
| Additional candidate differential tests | **PASS**: 5,000 random arithmetic pairs; 1,500 random plus boundary sqrt inputs × 7 scales × 7 modes, signed zero, Go interface/SQL fixture and policy-independence tests | [candidate test log](m2-candidate-unit-20261117.txt) |
| Complete repository suite, vet, and race | **PASS locally**, 14 tested packages and other packages without tests; `go vet` exit 0; full `-race` exit 0, including `optimus` | [suite](m2-candidate-all-tests-20261117.txt), [vet](m2-candidate-vet-20261117.txt), [race](m2-candidate-race-20261117.txt) |
| Documentation links | **PASS**, 207 links checked in 54 Markdown files | [link validation](m2-candidate-doc-links-20261117.txt) |

The frozen primary corpus includes text/JSON/SQL, parse, NaN, inexact,
scales, rounding, and arithmetic. The extended corpus adds Mod, QuoRem,
PowInt and unsigned encoding. These suites now execute against the **new
candidate**, unlike the earlier M1 public-legacy oracle replay. The
randomized tests additionally compare coefficient, stored zero sign, stored
scale and error reason, not only strings.

**Important limits:** AST equality does not prove interface or binary
compatibility. These deterministic seed sets do not exhaustively cover the
range of all 128-bit operands. No external SQL database driver was tested;
SQL behavior is based on `driver.Valuer`, scanner cases, a compile-time
interface/client fixture and golden fixtures.
The candidate is not wired into any public `dec128`, `dec256`, `dec512` or
`optimus` path. No release/cutover is authorized by these local successes.

## Reproduction commands and evidence boundaries

With **Go 1.23.2 only**, testing used a temporary sibling copy of the uploaded
upstream-main snapshot and an unshipped `go.mod` override. The following
commands all completed with exit status zero in that temporary environment:

```sh
GOTOOLCHAIN=local GOWORK=off go test ./... -count=1 -timeout=150s
GOTOOLCHAIN=local GOWORK=off go vet ./...
GOTOOLCHAIN=local GOWORK=off go test -race ./... -count=1 -timeout=180s
GOTOOLCHAIN=local GOWORK=off go test ./internal/dec128candidate -count=1 -v
python3 tools/check_doc_links.py
```

Candidate oracles and exact audits were regenerated after the final production
code changes using the commands in
[`tools/migration-candidate-oracle/README.md`](../../tools/migration-candidate-oracle/README.md).
No claim of full Go 1.27 or published-tag test coverage is made. Source-level
`internal` package import restrictions mean this fixture cannot yet prove
compatibility for arbitrary third-party client modules.

## Performance (local exploratory baseline)

The final [raw candidate benchmark](m2-candidate-bench-final-20261117.txt)
ran 5 samples per case, `-benchtime=150ms -benchmem` on this same Go 1.23.2
host. Medians are ns/op, with zero B/op and zero allocs/op for each case.
It uses **one operand pair per operation**; values are not a claim of
universal speed. Direct upstream operations have potentially different
rounding/loss semantics, so they are informational only.

| Operation | Public historical | Candidate | Direct Jokruger snapshot |
| --- | ---: | ---: | ---: |
| Add | 3.97 | 4.08 | 3.51 |
| Mul | 4.23 | 4.56 | 4.51 |
| Div | 10.48 | 10.92 | 13.13 |
| Parse | 22.89 | 18.94 | 16.32 |

The candidate preserves comparable performance on these *sampled* common
operands, but the broader M6 workload (errors, varied coefficient widths,
scale boundaries, scanners, sqrt, promotions and global-policy settings)
is still required before substitution.

**Discarded bridge delegation experiment:** using the same Go and host,
three 120ms benchmark samples showed before → after median (ns/op):
AddStrict 4.56 → 9.46; SubStrict 5.13 → 11.08; MulStrict 4.10 → 9.75;
DivAtPrecision 10.79 → 17.83; MulRound 12.96 → 19.91;
DivRound 14.36 → 25.53. **Reverted**, not shipped.

## Stage gates and remaining work

| Stage | Status | Remaining gate |
| --- | --- | --- |
| M0 | **Partial** | Final provenance/license/NOTICE review and official source checks |
| M1 | **Advanced, not formally closed** | Authenticate Jokruger tag, Go 1.27 build, run full official differential and speed gates |
| M2 | **Candidate A implemented and locally verified; public cutover blocked** | Third-party compile fixtures, full API review including behavior not in corpora, authentic module+toolchain tests, atomic cutover plus all downstream integration tests |
| M3 | **Pending** | Only start `uint128` substitution after validated M2 |
| M4 | **Pending** | Wide type 256/512 independent proof |
| M5 | **Pending** | `optimus` promotion/tier correctness |
| M6 | **Partial evidence** | Controlled benchmarks beyond the sampled workloads and final verified toolchain |
| M7 | **Blocked** | Provenance, Go 1.27/v1.4.0 and release gates |

**No silent public replacement occurred.** The current public `dec128`
continues to use its historical arithmetic except for the additive
`NewStoredZero` constructor. This is a deliberately separate candidate; a
partial public hybrid is prohibited. `dec256`, `dec512`, and `optimus` retain
their existing source and semantics while the candidate is tested.

## Change inventory

The [file-level manifest](m2-candidate-manifest-20261117.txt) records the
added and modified paths relative to the preceding cumulative ZIP. The
worktree was re-packed only after restoring the official module declaration.
No local module replacement or external source tree is included.
