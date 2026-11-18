> **Historical migration record (archived at M7-H).** This file contains

> **Archived engineering record.** This document describes a dated source
> checkpoint, not the currently recommended API or release state.
> For supported usage, see the [documentation index](../../../docs/README.md).
> time-specific stage status, proposed code and, in places, old rollback
> tag instructions that **do not apply to the current source tree**.
> The authoritative state is [M7-H release status](../../../docs/RELEASE_STATUS.md);
> do not follow an old "current" or "pending" claim as a live rule.

# Migration decisions — M0 ledger

M2-E current status (2026-11-17): D34's original **BLOCKED** whole-package
release applies to the earlier opt-in source tree. D45 records the **unpublished
source-default promotion**; D48 explicitly continues to block publication.

Status vocabulary: **KEEP** (preserve), **FIX** (approved behavior correction),
**DEPRECATE** (preserve symbol while steering consumers away), **OPEN** (decision
and/or verification pending). Changes to this ledger require corresponding
tests and a release-note entry.

| ID | Topic | Decision | New wrapper rule / verification |
| --- | --- | --- | --- |
| D45 | M2-E default-selection cutover | **STAGE (UNPUBLISHED)** | Default chooses exactly six candidate files; `badassitron_legacy` chooses exactly 12 historical files. Keep legacy source intact as rollback until final official M2-E tests and M7 review pass. `badassitron_m2` is a deprecated alias of default. No accidental engine mixing. |
| D46 | M6 `Rescale/Truncate` 12.6% relative slowdown | **DOCUMENTED EXCEPTION FOR EXPERIMENTAL DEFAULT; OPEN FOR RELEASE** | Supplied Go 1.27.1 medians 3.54 ns vs 3.98 ns (+0.44 ns), zero allocations. A direct uint64 fast-path experiment did not improve local Go 1.23.2 results and was reverted. Additional independent `math/big` tests and signed-zero/error precedence tests pass both variants. Rebenchmark representative workloads before M7 approval. |
| D47 | Official module checksum and licenses | **USER-SUPPLIED EVIDENCE; REVIEW** | Go 1.27.1 `go mod verify` report plus module metadata yielded two `h1:` entries, recorded in `go.sum`. The supplied audit reported `LICENSE`/`NOTICE` bytes identical to the snapshot. This environment could not re-download/verify; the final official-host rerun and historical Badassitron rights review remain open. |
| D48 | Cutover release boundary | **KEEP BLOCKED** | M2-E changes the source default but does not create a release tag or authorize publication. Run full checks and external-client tests on final code under Go 1.27.1 and official v1.4.0; resolve legal notices and documented M6 decision first. |
| D01 | Existing import paths and signatures | KEEP | Compare exported API snapshot and external build. |
| D02 | `Mul` precise-or-NaN, historic Overflow | KEEP | Explicit exact operation; map upstream Inexact in this context to historical Overflow. |
| D03 | Boundary overflow bug in `Mul` | FIX | Independent `math/big` boundary oracle; never preserve incorrect wrapped value. |
| D04 | `PowInt(math.MinInt)` panic | FIX | Guard integer exponent negation; distinguish valid result and domain/capacity errors. |
| D05 | `errors.Error.Value()` invalid index panic | FIX | Bounds check; return descriptive `invalid state code` error. Preserve 0–10 codes. |
| D06 | Exact `1e20 / 1` returned NaN | FIX | Permit exact representable result by reducing scale. |
| D07 | Parser accepts `"5."` | KEEP | Parse golden test. |
| D08 | Parser maps empty input to zero | KEEP | Parse golden test; warn that required inputs need validation. |
| D09 | `Scan` value receiver | DEPRECATE | Keep method set initially; recommend `ScanDestination`. Pointer receiver migration needs separate approval. |
| D10 | `Precision` stored scale and format | KEEP | Preserve zeros and string output; tests. |
| D11 | Rounding enum numeric values and behavior | KEEP | Map by semantic mode, never enum ordinal. |
| D12 | Error enum numeric values 0–10 | KEEP | Preserve integers, append any new states after old values. |
| D13 | Upstream Inexact/Null/InvalidRoundingMode/DomainError | OPEN | Decide per public method; never silently collapse into a generic error. |
| D14 | Signed zero / NaN identity and serialization | KEEP | Extend fixture set before M2. |
| D15 | Normal arithmetic affected by upstream globals | KEEP | Explicit deterministic per-call policy; test under changed globals. |
| D16 | `uint128` external facade | KEEP | Retain historic exported API; internal hot paths may use upstream directly. |
| D17 | Go minimum | KEEP | Leave `go 1.27` unchanged pending separate version matrix. |
| D18 | `withdec128`, `withfloat64` | KEEP | Do not remove during migration. |

## M1 continuation — 2026-11-17 (experimental only)

| ID | Topic | Decision | Verification |
| --- | --- | --- | --- |
| D19 | Long parser's overflow/format/scale precedence | **KEEP** | Import upstream `uint128.FromString` for substrings without a redundant pre-scan; preserve original short-vs-long selection by full input length. Long-input malformed data can report overflow before a later invalid character; long fractional scale is checked first. `parser_gates_test.go`. |
| D20 | Stored negative zero at scale zero | **KEEP** | Historical `New(0,0,true)` normalizes, but long-integer parsing and `Rescale(0)` can retain a negative-zero sign bit. Migration-only conversion must preserve it; `signed_zero_migration_test.go` and `parser_gates_test.go`. |
| D21 | Exact zero vs rounding to zero | **KEEP** | Historical `RescaleRoundWithStatus` retains sign metadata for existing zero, but normalizes negative nonzero values that round to zero. `RescaleRoundStrict` preserves this distinction. |
| D22 | Zero results of Add/Sub | **KEEP** | Preserve stored scale/sign for the old same-sign zero path and canonical zero after cancellation; the equal-scale bridge fast path must check `StoredNegative`, not just `IsNegative`. |
| D23 | Parser speed tradeoff | **OPEN (performance)** | Long input speedup on this host, short input and invalid-scale regressions. Do not claim general speedup until full Go 1.27 / published-tag M6 benchmarks. |

[Full M1 checkpoint report](../m1-parser-signed-zero-20261117.md).

## Context-dependent error translation

The same upstream reason may need different historical public codes. For
example, `Mul` with an unrepresentable exact-scale result historically reports
`Overflow`. This must be a policy in its adapter, not a global `Inexact`
conversion. Domain, invalid mode, NULL, divide-by-zero and scale-related
statuses must be tested by public operation before a mapping is finalized.

## Protected performance path

For exact multiplication with 64-bit coefficients and permitted summed scale,
`math/bits.Mul64` can construct the 128-bit coefficient without a Jokruger
rounding call. The bypass is valid only if all representation, sign, scale,
error and global-independence invariants hold. Its benefit is subject to a
same-machine before/after benchmark; it is **not** assumed to be faster.

## M1 signed-zero representation checkpoint (2026-11-17)

The experimental bridge must retain the historical finite negative-zero sign bit and stored scale using explicit wrapper metadata; raw conversions through upstream `jk.New` are lossy. The read-only `Dec128.StoredNegative()` accessor is additive and leaves the numeric API unchanged. This is an engineering choice for the experimental bridge, **not** approval of an M2 cutover or proof of operation-level signed-zero compatibility. See [checkpoint](../m1-compatible-representation.md).

## M1-final and M2 candidate-A checkpoint — 2026-11-17

| ID | Topic | Decision | Verification / risk |
| --- | --- | --- | --- |
| D24 | Raw negative zero at scale 0 | **KEEP / additive constructor** | `NewStoredZero(scale, negative)` is additive. `New(0,0,true)` keeps its original normalization. No legacy arithmetic is required in the bridge to create this state. `migration_stored_zero_test.go`, `zero_division_contract_test.go`. |
| D25 | `Div` underflow to raw negative zero | **KEEP** | The historical truncated quotient may retain a negative sign bit with coefficient zero and scale zero. Bridge `DivAtPrecision` and M2 candidate both reproduce it. |
| D26 | `Round` versus `RescaleRound` | **KEEP** | Round-family operations do not increase stored scale or validate a higher-than-existing target scale; `RescaleRound` validates and quantizes explicitly. Their signed-zero behavior may differ. M2 primary frozen oracle and randomized matrices. |
| D27 | Candidate declaration surface | **PARTIAL / NOT CUT OVER** | AST checker finds 83 original methods/functions, 37 declarations and both type names with no missing/mismatched signatures. This does not prove source interface satisfaction or binary compatibility. |
| D28 | Per-operation error and inexact policy | **KEEP** | Candidate uses imported Jokruger primitives with explicit modes, retaining old error codes. All 5,243 frozen cases and independent mathematical checks pass locally; official tag validation is open. |
| D29 | Delegate M1 bridge via M2 candidate | **REJECTED on performance** | Tests passed but measured bridge paths regressed 54–138%. The delegation experiment was reverted, with raw comparison recorded in the M2 candidate-A report. |
| D30 | Atomic public replacement | **BLOCKED** | Candidate tests do not authorize cutting over public `dec128`. Require Go 1.27, published verified Jokruger v1.4.0, official `go.sum`, external API/client compile, full M2/M6 gates and provenance review. |

[Full M1-final/M2-CANDIDATE-A report](../m1-final-m2-candidate-a-20261117.md).


## M2-B experimental public cutover — 2026-11-17

| ID | Topic | Decision | Evidence / remaining risk |
| --- | --- | --- | --- |
| D31 | Public candidate cutover as a whole package | **STAGE, NOT RELEASE** | Compile `dec128/m2_*.go` only under `badassitron_m2`; compile all legacy arithmetic only without that tag. Default API remains historical. Both complete functional suites pass locally. |
| D32 | Mixed-scale Add/Sub performance | **OPTIMIZE, KEEP SEMANTICS** | Exact signed 128-bit alignment using imported upstream `uint128` primitives, canonical retry on overflow, signed-zero and stored-scale oracles. All 5,243 public frozen cases pass. |
| D33 | Rescale and explicit rounding performance | **OPTIMIZE, KEEP SEMANTICS** | Imported coefficient operations; local 64-bit remainder rounding path for at most 19 discarded digits, seven public rounding modes, normalized negative rounded-to-zero behavior. Separate five-sample half-even benchmark improved. |
| D34 | Whole-package M2 go-live | **BLOCKED** | Go 1.27 and official v1.4.0 checksums, final provenance, complete tagged race, external official-module compilation, and documented M6 exceptions remain. A build tag is not approval to publish. |
| D35 | Temporary M2 implementation mirror | **STAGE / REMOVE LATER** | Six tagged public files mirror `internal/dec128candidate` under `check-mirror.py`. No upstream code duplicated. Delete redundant candidate after final atomic cutover and preserve separately frozen history for oracle. |

[M2-B full report](../m2b-experimental-public-cutover-20261117.md).


## M2-C decisions — 2026-11-17

| ID | Topic | Decision | Evidence and remaining gate |
| --- | --- | --- | --- |
| D36 | Sqrt inexactness | **OPTIMIZE; KEEP CONTRACT** | One explicit upstream `SqrtRound`, with an independent squared-coefficient check using imported `uint128`. Independent `math/big` and historical 1,500-case randomized status tests pass locally. |
| D37 | SQL scanner | **OPTIMIZE; ATOMIC DESTINATION** | Direct type-specific `ScanInto`, destination changed only after successful parse, historical `Dec128.Scan` value receiver retained. Type/error regression tests pass. |
| D38 | Fixed scale formatting | **OPTIMIZE; KEEP FORMATS** | Imported `uint128.StringToBuf` plus wrapper-managed sign/scale; zero allocations beyond the returned string for the sampled nonzero value. Parser short path simplified. |
| D39 | Multiplication overflow retry | **OPTIMIZE ONLY WHEN EQUIVALENT** | Skip second 256-bit product if canonicalized scales do not change. Keep retry when canonicalization changes either operand. |
| D40 | M6 go-live | **STILL GATED** | Paired performance evidence on Go 1.23.2 is informative; Go 1.27.1 official module and full races need rerun on M2-C. Historical provenance and official `go.sum` still open. |

[M2-C report](../m6c-performance-provenance-20261117.md).

## M2-D decisions — 2026-11-17

| ID | Topic | Decision | Evidence and remaining gate |
| --- | --- | --- | --- |
| D41 | `DivRoundWithStatus` 64/128-bit fast path | **OPTIMIZE; KEEP CONTRACT** | Imported `uint128.QuoRem64`/`QuoRem` when aligned numerator is 128-bit; imported 256/128 division remains for a wide numerator. A deterministic independent rational `math/big` test covers 1,507 pairs × six scales × seven rounding modes, plus all frozen cases. |
| D42 | Half-way quotient comparison | **KEEP EXACTNESS** | For a 128-bit divisor, compare the remainder with divisor minus remainder; for a 256-bit divisor, retain the high-limb comparison. No truncation before rounding or implicit use of upstream global policy. |
| D43 | Official dependency metadata audit | **VERIFY ON OFFICIAL HOST; NOT CLEARED LOCALLY** | Compare module `h1:`/go.mod sums with `go.sum` and report installed official `LICENSE`/`NOTICE` SHA-256 against precautionary snapshot. Tool unit tests pass locally; source-chain and notice sufficiency require human review. |
| D44 | M6-D promotion | **NOT APPROVED** | Five 75ms paired samples of 20 operations show no >10% sampled review flags, but Go 1.27.1, official upstream, longer representative runs and provenance are still outstanding. Keep the default implementation historical. |

[M2-D full report](../m2d-fast-division-provenance-20261117.md).

## M3-A unsigned public adapter — 2026-11-17

| ID | Topic | Decision | Evidence and remaining gate |
| --- | --- | --- | --- |
| D49 | Historical `uint128` public surface | **KEEP** | Preserve exported `Uint128{Lo,Hi}`, signature-level API and historical `errors.Error` codes. All 66 exported declarations match between default/M3 and an external client compiles in both configurations. |
| D50 | Upstream primitives at public `uint128` boundary | **OPT-IN ONLY** | `badassitron_m3` selects one set of arithmetic/bitwise/wide division implementations, without copying upstream or making an alias. Common historical codec and constructors remain unchanged deliberately. |
| D51 | M3 performance gate | **REJECT DEFAULT PROMOTION** | Local five-sample microbenchmarks show +90.3% Add and +26.1% Mul overflow (last paired 200ms run); exactness PASS does not override the regression budget. |
| D52 | M2-E and M3 official gate | **STILL BLOCKED** | The supplied Go 1.27.1 verification is for M2-D, not this M2-E/M3 tree. Exact official tag rerun, M6 decision and historical-source provenance approval remain required. |

[M3-A detailed report](../m3-uint128-opt-in-20261117.md).

## M3-C multiplication-wrapper decisions — 2026-11-17

| ID | Topic | Decision | Evidence and remaining gate |
| --- | --- | --- | --- |
| D53 | Delegation and error semantics | **KEEP** | `Mul` still calls imported Jokruger, returns historical `Overflow`/zero on error; independent exact/overflow checks pass. |
| D54 | Compiler inlining | **OPTIMIZE** | Direct Go value conversions reduce M3 wrapper's local cost from 107 to 77 (under the 80 inlining budget). No upstream code copied. |
| D55 | Benchmark path coverage | **EXPAND** | Added 128×128 successful `Mul` (narrow/high limb) to avoid interpreting `Mul64` as successful `Mul`; three-way comparisons use seven cases. |
| D56 | Default M3 promotion | **REJECT FOR NOW** | Exact-success/overflow regressions remain in the sampled local runs; repeat under official Go 1.27.1 and published module before considering cutover. |

[M3-C report](../m3c-mul-inline-20261117.md).

## M3-D unsigned division mapping — 2026-11-17

| ID | Topic | Decision | Evidence / remaining gate |
| --- | --- | --- | --- |
| D57 | Ordinary unsigned division error codes | **OPTIMIZE; PIN CONTRACT** | `Div`, `Mod`, `QuoRem` and 64-bit variants map only Jokruger v1.4.0 `OK`/`DivisionByZero` statuses and retain historical result/error codes. 2,000 interop input sets pass both builds. Wide division still uses full mapping. |
| D58 | Benchmark coverage | **EXPAND** | Compare eleven facade/direct paths, separating exact multiplication, overflow, 64-bit division, division by zero and wide overflow. |
| D59 | Narrow-multiplication dispatcher | **REJECT** | A `Mul64` fast-path experiment lost wrapper inlining and regressed full-width overflow; the change was reverted. |
| D60 | Default `uint128` promotion | **NOT APPROVED** | Measured zero-divisor and wide-overflow costs remain high. No Go 1.27.1 official-module validation on M3-D; provenance and M6 gates remain open. |

[M3-D report](../m3d-division-status-three-way-20261117.md).

## M3-E — Architectural resolution (2026-11-17)

| ID | Topic | Decision | Rationale and evidence |
| --- | --- | --- | --- |
| D61 | Default decimal backend versus unsigned public facade | **KEEP SPLIT; NO M3 DEFAULT PROMOTION** | M2 stores upstream coefficients directly; the public `uint128` package retains its historical nominal type. The official M3-D eleven-path three-way comparison shows nine adapter paths exceeding +10%, while M2 already uses Jokruger directly for its own coefficient arithmetic. |
| D62 | Internal coefficient construction | **REMOVE UNNECESSARY PUBLIC ROUND TRIP** | `fromParts` uses a new upstream-typed constructor; `New` remains the public boundary. 4,000 constructor cases pass, preserving scale, signed zero and overprecision. |
| D63 | Architecture regression guard | **REQUIRE** | New Go-AST source guard, its tests, the existing candidate mirror/API/build-selection checks and optional assembly diagnostics. Source-level analysis alone is not a complete call-graph proof. |
| D64 | Public unsigned full delegation | **DEFER** | Preserve M3 opt-in for experimentation. Revisit only with representative Go 1.27.1 performance and no API/error drift. No hybrid per-method facade or copied upstream source. |

[M3-E report](../m3e-architecture-resolution-20261117.md).

## M3-F — Imported public unsigned backend by default (2026-11-17)

These decisions **supersede D61 and D64 as default selection**, while keeping
the original M3-E report as historical evidence.

| ID | Topic | Decision | Reason and control |
| --- | --- | --- | --- |
| D65 | Default public `uint128` backend | **PROMOTE IN SOURCE, UNRELEASED** | Imported Jokruger is the only normal arithmetic implementation; full historical version is test/rollback-only. Source selection and 66 exported signatures are checked. |
| D66 | Compatibility versus raw throughput | **KEEP API AND ERROR SEMANTICS** | Nominal `Uint128` and `errors.Error` prevent a transparent alias; use zero-copy Go conversions and operation-specific mapping without copying upstream arithmetic. Benchmark facade vs direct Jokruger on the same inputs. |
| D67 | Unsigned codecs | **DELEGATE TO JOKRUGER BY DEFAULT** | Historical `FromString` and `StringToBuf` move behind rollback build tags; random/boundary differential tests compare both modes to imported Jokruger. |
| D68 | Release gate | **BLOCKED PENDING OFFICIAL RE-VALIDATION** | Full Go 1.27.1/v1.4.0 test/race/oracle/benchmark rerun, provenance approval, and decision on remaining adapter regressions are required; no tag is published by this checkpoint. |

[M3-F report](../m3f-default-imported-uint128-20261117.md).

## M7-D — Imported AddCarry performance boundary (2026-11-18)

| ID | Topic | Decision | Evidence and release risk |
| --- | --- | --- | --- |
| D69 | Addition backend | **OPTIMIZE PUBLIC ADAPTER; KEEP UPSTREAM ENGINE** | Use imported `Jokruger.Uint128.AddCarry` and its 0/1 carry rather than the imported `Add` status. Mask the modular sum to preserve historical zero-on-overflow and return `errors.Overflow=3`. 24,000 upstream comparisons and 4,096 independent `math/big` additions pass locally. No 128-bit addition algorithm is copied. |
| D70 | Exceptional division and short parsing | **REJECT UNVERIFIED SHORTCUTS** | Prechecks in `QuoRem`/`QuoRem256By128` lost wrapper inlining (local costs 80→97 and 80→125); simpler `FromString` status mapping did not improve short parsing reliably. Keep M7-C implementations, record exceptional-path benchmarks, guard wide division error precedence. |
| D71 | Promotion and publication | **PENDING OFFICIAL M7-D GATE** | Official Go 1.27.1 M7-C passes, but no Go 1.27.1 verification or benchmark is yet available for this changed M7-D tree. Compare same-host facade/direct runs and watch consumers. Historical Badassitron licensing and redistribution clearance are separate release blockers. |

[M7-D detailed report](../m7d-addcarry-review-20261118.md).
