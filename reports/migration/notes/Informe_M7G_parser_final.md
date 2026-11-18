# Badassitron M7-G — Final `ParseShort` Adapter Investigation

> **Archived engineering record.** This document describes a dated source
> checkpoint, not the currently recommended API or release state.
> For supported usage, see the [documentation index](../../../docs/README.md).

**Date:** 2026-11-18  
**Final decision: RESULT B — OPTIMIZATION REJECTED. Keep M7-D unchanged.**  
**Next stage:** M7-H (documentation, license/provenance audit and publication readiness); no fourth parser variant.

## 1. Scope, immutable reference and evidence provenance

This is the completed, bounded M7-G investigation. The production baseline is the M7-D code shipped unchanged in the attached M7-F archive. **No M7-E, G2 or G3 parser change has been promoted.** The original M7-G report was updated only to incorporate actual local-upstream tests, a real adapter disassembly and interleaved measurements made possible by the later Jokruger source attachment.

Inputs independently inspected:

1. `badassitron_M7F_interleaved_AB_D_stable_20261118(1).zip` — M7-D production and A/B tools.
2. `badassitron_M7E_tres_rutas_20261118(1).zip` — prior, unpromoted M7-E candidate.
3. `dec128-main(2).zip` — subsequently attached Jokruger source snapshot.
4. The user's supplied account of the official Go 1.27.1 M7-D/M7-E paired benchmark campaign.

**Source identity limitation:** the Jokruger ZIP is a `main` source snapshot, with module path `github.com/jokruger/dec128` and `go 1.24` in its original `go.mod`. It does **not** contain sufficient tag or checksum evidence to authenticate its content as **v1.4.0**. The parser implementation was inspected and used unmodified for **local diagnostics only**. Its original `uint128/from.go` SHA-256 is `b326095a6828cc57c64b56a84e7c835c8adb4f06fcb792d5673d95e8d23316fa`; the attached ZIP SHA-256 is `c93593087448a3d888e887fbfcbc6a98abfb9d9ae8484a1e7bb44c38d9101d3b`. Neither hash authenticates the official release tag. The official Go 1.27.1 compiler and remote Jokruger v1.4.0 module are unavailable here.

The local installed compiler is `go1.23.2 linux/amd64`, running on a shared KVM AMD EPYC 9V74 host. **Only in discarded temporary validation copies**, both module `go` directives were lowered to `1.23` and the Badassitron module temporarily used `replace github.com/jokruger/dec128 => <local-source>`, with `GOPROXY=off`, `GOSUMDB=off` and `GOTOOLCHAIN=local`. This avoids false claims about an official version and preserves the distributed module metadata and architecture. The source parser algorithm itself was not altered, replicated, vendored or moved into Badassitron.

The delivered `go.mod` still declares **`go 1.27`** and **`github.com/jokruger/dec128 v1.4.0`** with **no `replace`**, workspace or vendor directory. All **78 non-test production Go files**, the parser, `go.mod` and `go.sum` are byte-identical to the original M7-F production. The stable parser SHA-256 remains:

```
6351f382fb109a6f6db2fba6c25445bedb8367156e2a123d608cc8a6796c42cd
```

The official raw `Archivador(10).zip` was not supplied as an independent attached source, so the earlier official confidence intervals cannot be reconstructed independently. The local raw measurements described below **are included**.

## 2. Parser adaptation, semantics and real machine code

The M7-D implementation delegates every input directly to Jokruger, then translates its result:

```go
func FromString[S string | []byte](s S) (Uint128, errors.Error) {
    n, status := ju.FromString(s)
    switch status {
    case state.OK:
        return Uint128(n), errors.None
    case state.Overflow:
        return Zero, errors.Overflow
    case state.InvalidFormat:
        return Zero, errors.InvalidFormat
    default:
        return Zero, fromState(status)
    }
}
```

Jokruger's unsigned decimal parser on this source snapshot returns `state.OK = 0`, `state.Overflow = 5`, or `state.InvalidFormat = 9`, corresponding to historical Badassitron errors `None = 0`, `Overflow = 3` and `InvalidFormat = 7`. Error returns are the zero value. The adapter does not perform any string scanning itself, does not alter digit-validation/overflow precedence and retains `fromState` fallback for unexpected upstream codes. The `Uint128(n)` value conversion converts two congruent 64-bit limbs; it does not invoke an arithmetic algorithm.

The newly attached source enabled **actual disassembly of the real Go 1.23.2 adapter with the imported Jokruger parser**, replacing the previous synthetic-only assessment. In `assembly/adapter-fromstring-objdump-go1232.txt`, the string instantiation contains the meaningful sequence:

```asm
CALL github.com/jokruger/dec128/uint128.FromString[go.shape.string]
TESTL CL, CL     // OK
JE success
CMPL CL, $0x5   // overflow
JE overflow
CMPL CL, $0x9   // invalid format
JE invalid
...              // inlined fromState fallback for unexpected codes
```

The generic shape implementations **are not inlineable** under Go 1.23.2: the adapter has estimated cost 142 versus inline budget 80, and the upstream parser has estimated cost 241 versus 80. Compiler messages also say that the concrete `string` and `[]byte` adapter wrappers are inlineable with cost 72; that does **not** make the generic parser body inline. The machine code shows a call boundary, branch comparisons and an inlined error-fallback switch. There is no extra parsing pass and no new heap allocation. Assembly and inlining behavior **must not be extrapolated to Go 1.27.1** without compiling there.

Full upstream parser source or its entire assembly was deliberately **not bundled**, to respect the single-imported-backend architecture. Only Badassitron adapter disassembly and short compiler diagnostics are kept.

## 3. Exactly three hypotheses and variants

| Variant | Hypothesis / implementation | Semantics / decision |
|---|---|---|
| **M7-E (H1)** | Branchless status translation: `code := uint8(status); return Uint128(n), errors.Error(code - 2*((code+3)>>3))` | Valid for the pinned status domain 0/5/9, but lacks M7-D's future-state fallback and assumes upstream zero on errors. Official `ParseShort` threshold already missed. Rejected. |
| **G2 (H2)** | Single success check; on failure `return Zero, fromState(status)` | Preserves all historical error mappings; did not show a local improvement. Rejected. |
| **G3 (H3)** | Success check, explicit overflow/invalid returns, then `fromState` fallback | Preserves error mapping; generated control flow remains similar to M7-D. Rejected. |

All three variants exist only as inert `.go.txt` snapshots in `tools/release/m7g-variants/`. The helper `tools/release/m7g-stage-variant.py` stages a separate complete candidate directory and verifies that exactly `dec128/uint128/from_upstream_string.go` changes. **No fourth variant or new parser implementation was introduced.**

## 4. Performance: official history versus local source-snapshot diagnostics

### 4.1 User-supplied official Go 1.27.1 evidence (not reproduced here)

| Target | M7-D | M7-E | Reported effect |
|---|---:|---:|---:|
| `ParseShort` | 6.940 ns/op | 6.798 ns/op | **−2.1%**, below required **−3%** |
| `ParseShortMixed` | 7.275 ns/op | 6.973 ns/op | −4.1% |
| `ParseLong` | 39.450 ns/op | 39.545 ns/op | Essentially unchanged |

The supplied campaign used 8 ABBA/BAAB blocks, 16 runs per version and 250 ms per benchmark. The missing official raw archive prevents independent interval recomputation. The main mandatory acceptance criterion already fails based on the reported 2.1% gain.

### 4.2 Local Go 1.23.2 exploratory A/B: three short campaigns

The existing `tools/release/bench-interleaved.py` was used with `--local-upstream` and the attached **source snapshot**. Each candidate received **one** three-block ABBA/BAAB campaign with **90 ms per subbenchmark**. All campaigns included five operations, facade and direct upstream controls, matched tests and allocation checks. Figures below are the comparator's **median of within-block paired percentage changes**, not ratios of separately reported overall medians.

| Candidate vs M7-D | `ParseShort` paired change | Interpretation |
|---|---:|---|
| M7-E | −8.1% | Best *exploratory local* signal; selected for the one extended A/B |
| G2 | +2.6% | Slower; no further campaign |
| G3 | +1.8% | Slower; no further campaign |

The script historically calls any second candidate `M7-E` in its generated text and CSV. **For folders named `g2` or `g3`, that label means the folder's candidate source, not M7-E.** Source selection is documented and all raw records are preserved unchanged. These exploratory results are not statistically adequate for promotion and cannot be considered official.

### 4.3 **One final local** eight-block A/B: M7-E versus M7-D

This was the **only extended A/B** performed in M7-G after the source attachment: 8 ABBA/BAAB blocks; 16 runs per variant; **250 ms per subbenchmark**; five matched operations; direct Jokruger in every run; zero bytes/op and zero allocations/op across **320 sample rows**. Raw records, the untouched `samples.csv` and per-block report are in `reports/migration/m7g-evidence/local-jokruger-go1232/bench/m7e-final-local/`.

| Operation | M7-D facade median | M7-E facade median | **Paired** median change | 95% block-bootstrap interval | Blocks favoring E |
|---|---:|---:|---:|---|---:|
| `ParseShort` | 10.160 ns | 9.866 ns | **−1.7%** | **[−7.2%, +2.2%]** | 5/8 |
| `ParseShortMixed` | 10.110 ns | 9.689 ns | −5.4% | [−9.8%, +2.4%] | 6/8 |
| `ParseLong` | 48.860 ns | 47.030 ns | −3.5% | [−6.9%, +6.0%] | 5/8 |
| `QuoRem` | 7.898 ns | 7.771 ns | +0.0% | [−7.1%, +6.5%] | 4/8 |
| `WideQuoRem` | 11.590 ns | 11.645 ns | +3.4% | [−6.5%, +5.0%] | 3/8 |

The eight-block bootstrap is descriptive; with shared hardware and one campaign it is not a formal, portable probability bound. **`ParseShort` misses the 3% improvement gate and its interval crosses zero.** A +3.4% paired median is observed for `WideQuoRem`, but that control interval crosses zero as well; a reproducible >2% regression is **not** established. The failed primary gate alone is decisive.

Local direct-upstream `ParseShort` medians: **8.051 ns** (D-side measurements) and **8.082 ns** (E-side measurements). Using separately pooled medians, D's facade is **+26.2%** and E's facade **+22.1%** above direct upstream. These *local Go 1.23.2 source-snapshot* overheads are **not comparable** to the official Go 1.27.1 ~11.8% / ~9.5% figures; they also do not meet the preferred <=5% direct overhead. No micro-optimization was justified by a single attractive local sample.

## 5. Functional, architectural and release validation

### Completed with the attached source snapshot (exploratory Go 1.23.2)

| Check | Result |
|---|---|
| `go test ./dec128/uint128 -count=1` on M7-D | **PASS** (real attached Jokruger parser, not a stub) |
| `go test ./... -count=1` | **PASS** |
| `go test -race -p 2 ./... -count=1` | **PASS** |
| `go vet ./...` | **PASS** |
| Compatibility tag build/tests, `badassitron_m2,badassitron_m3` | **PASS** |
| Baseline primary frozen oracle | **3,518 cases; 0 unexpected differences**; 389 exact checks, zero arithmetic mismatches |
| Baseline extended frozen oracle | **1,725 cases; 0 unexpected differences**; 1,634 checks, zero mismatches |
| Three candidate parser tests within interleaved runner | **PASS** for M7-E, G2 and G3 |
| Three exploratory campaigns + final extended campaign | **PASS** as measurements; **0 added allocations** |
| Prior M7-G static guards / API audit | **PASS** (1,030 historical symbols, 66 `uint128` signatures) |
| Production equality to M7-F stable reference | **PASS** (78 source files, modules unchanged) |

The full 5,243-case oracle result was independently recomputed from frozen inputs via the real local upstream snapshot. Both `tools/migration-oracle/compare.py` and the corresponding exact checks passed; raw generated outputs are fingerprinted in the evidence directory, and the compare/check logs are included. The three variants' normal parser test runs also passed. M7-D remains the only full-suite/race/oracle validated production version; candidate status does not matter because none was promoted.

### Still not established, and explicitly **not** claimed

- **Official Go 1.27.1** compiler tests, race tests, `vet`, assembly and benchmarks.
- Authentication and `go mod verify` of published Jokruger **v1.4.0** (the supplied `main` ZIP is not the official tag).
- `GOWORK=off bash tools/release/verify.sh`: its **Go 1.27 minimum gate blocks this environment** even though exploratory tests pass with temporary metadata changes.
- An official 95% paired confidence interval for the prior Go 1.27.1 measurements (raw archive not attached).
- The <=5% preferred direct-upstream overhead on the official toolchain.

No blocked/interrupted verification is counted as passing. The earlier official M7-D release verification retained in M7-F remains historical evidence, not a newly executed M7-G certificate.

## 6. Reproduction and delivery integrity

The complete cumulative repository includes the unchanged M7-F source, original M7-G experimental snapshots, benchmark runner, staging script and new source-snapshot evidence. The attached Jokruger ZIP is intentionally **not** embedded or vendored into Badassitron. Raw local A/B measurements are preserved without postprocessing or name substitution.

To rerun **local diagnostics only**, unpack the new repository and the supplied Jokruger source into separate folders, and use:

```sh
export GOWORK=off GOTOOLCHAIN=local
python3 tools/release/m7g-stage-variant.py --baseline . --variant m7e --output /tmp/m7g-m7e
python3 tools/release/bench-interleaved.py \
  --baseline . \
  --candidate /tmp/m7g-m7e \
  --output /tmp/m7g-local-jokruger-final \
  --blocks 8 --benchtime 250ms \
  --local-upstream /path/to/dec128-main
```

The runner alone stages temporary module overrides for older Go; never edit the delivered `go.mod`. To reproduce **exploratory local full tests**, copy both projects to separate disposable directories, set `go 1.23` only in those copies of `go.mod`, add a local Jokruger `replace` only to the disposable Badassitron module, and run the exact commands in the validation table. **This is not a substitute for official release verification.**

On an official Go 1.27.1 machine, remove all diagnostic overrides and use the already included `reports/migration/m7g-evidence/exact-reproduction-commands.sh` and `tools/release/verify.sh` with the official published module. That official replay is not a pending condition for this **rejection**: no production candidate was selected.

Files **changed in M7-G** relative to M7-F are limited to documentation, inert variant snapshots, staging/benchmark evidence and test/release tools. Files **changed after the later Jokruger attachment** are exclusively the present report and `reports/migration/m7g-evidence/` records/manifests. Production files modified: **zero**. The ZIP integrity, archive traversal/symlink protection, SHA-256 manifest, absence of embedded Jokruger source and absence of temporary module overrides are checked in the final integrity evidence.

## 7. Unambiguous decision

**RESULT B — OPTIMIZATION REJECTED.** M7-E's user-reported official `ParseShort` improvement was only 2.1%, below the mandatory 3%. With the newly supplied Jokruger source, the one expanded local A/B again failed (paired −1.7%, 95% interval crossing zero); G2 and G3 produced no qualifying signal in their single exploratory campaigns. None satisfies all performance, compatibility and official verification gates. **Retain M7-D byte-for-byte, close parser optimization after the allotted three variants, and move to M7-H for documentation, licensing, provenance and publication readiness.** No publication is claimed.
