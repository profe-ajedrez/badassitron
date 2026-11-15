# BADASSITRON-MULROUND-OPT-01 — Dec1024 round-multiplication experiment

**Decision:** retain as the next experimental Dec1024 baseline, not a production migration approval.

## Scope and provenance

The exact input source is `badassitron_pow_opt_01_complete.zip` (BADASSITRON-POW-OPT-01). The original extracted repository is preserved as `baseline-original` in the working experiment. No changes were made to Optimus, EUCLIDS, Iacon, other decimal backends, the public API, or rounding definitions. The delivered ZIP retains the original `go 1.27` declaration and dependency graph. Verification in this environment used a distinct copy with only `go.mod` adjusted to `go 1.23`. Go 1.27 compatibility **was not tested**.

## Findings

1. `MulRoundWithStatus` first called `Uint1024.MulCarry`, then multiplied the original operands **again** with `fixedProduct` for a wide product or excessive scale.
2. `fixedQuantize` used a full fixed-width `10^n` divisor and generic wide division even when the only required operation was division by a decimal power.
3. When a 1024-bit product fitted within its width but needed a smaller output scale, the existing fast path went through `RescaleRoundWithStatus` and the slower 1024-bit bit-at-a-time divider. This accounted for the large 512-bit benchmark cost.

## Changes

- `dec1024/dec1024.go`: reuse the low and high halves from the original `MulCarry`, selecting the existing rescale path only when the desired scale is at least the exact product scale and the product is narrow.
- `dec1024/fixed_arith.go`: reconstruct the already computed full 2048-bit product in a local fixed buffer (`fixedFromProduct`), without repeating multiplication.
- Quantize to lower scales using repeated unsigned divisions by `10^19` and a final `10^k` (`1 <= k <= 19`). Track discarded lower digits with one sticky bit, and use the final remainder to classify below/at/above half for the existing seven modes. This preserves `inexact`, ties-to-even, directed rounding, coefficient overflow, errors, signed zero behavior, and output precision.
- Process only significant 64-bit limbs in the division-by-64-bit primitive; omit the now-unused `fixedPow10` helper.
- Added direct deterministic differential tests of all seven rounding modes, including large coefficients, extreme scales, exact half ties and half plus/minus one, and zero-allocation tests. Added a focused benchmark matrix.

No new dependencies, interfaces, pools, `unsafe`, or heap-based arithmetic were introduced.

## Verification

Environment: Intel Xeon Platinum 8573C, Linux/amd64, `go version go1.23.2 linux/amd64`, `GOMAXPROCS=1` for measurements.

| Verification | Result |
|---|---|
| `go test ./...` | PASS, candidate temporary copy |
| `go vet ./...` | PASS, candidate temporary copy |
| `go test -race ./dec1024 ./dec1024/uint1024 ./internal/uintwide` | PASS |
| New differential test: 600 operand pairs x 7 rounding modes | PASS (coefficient, scale, sign, error, `inexact`) |
| New isolated quantization test: 450 full-width cases x 2 signs x 7 modes | PASS against `decimalwide`/`math/big` |
| Half-tie/sticky tests (below/exact/above, both signs, 7 modes) | PASS |
| Zero-allocation checks (small + large) | PASS (`0 allocs/op`) |
| Full repository `-race` | **Not run** |
| Go 1.27 build, tests, race | **Not verified** |

The existing `decimalwide`/`math/big` implementation remains the independent **test oracle**; the candidate's Dec1024 calculation path is fixed width.

## Main CPU benchmarks

Four alternating baseline/candidate passes, `GOMAXPROCS=1`, `go test ./dec1024 -run '^$' -bench '^BenchmarkMulRoundOpt/' -benchmem -benchtime=140ms -count=1`. The table contains median ns/op; lower is better. Complete raw passes are in the evidence ZIP.

| Scenario | Baseline ns/op | Candidate ns/op | Change |
|---|---:|---:|---:|
| small-exact | 221.0 | 225.3 | +1.9% |
| small-round-2 | 761.4 | 566.0 | -25.7% |
| small-round-8 | 881.1 | 552.5 | -37.3% |
| medium-512bit-round-2 | 50,333.0 | 973.0 | -98.1% |
| wide-2000bit-round-2 | 4,534.5 | 3,869.5 | -14.7% |
| wide-2000bit-round-308 | 3,997.0 | 3,045.0 | -23.8% |
| wide-max-round-0 | 5,071.5 | 4,308.5 | -15.0% |
| wide-max-round-308 | 4,175.0 | 3,186.0 | -23.7% |
| wide-max-round-ceil | 4,683.5 | 4,207.5 | -10.2% |

All nine scenarios reported 0 B/op and 0 allocs/op for baseline and candidate. The +1.9% in `small-exact` is too small to distinguish reliably from run-to-run noise. Do not extrapolate these microbenchmarks to Iacon or other workloads.

## Isolating the two changes

A separate short two-sample experiment tested each optimization alone. These data are diagnostic only; the four-pair run above is the main comparison.

- **Product reuse only:** the wide product with target 2 fell from roughly 4.65 µs to 3.85 µs; it does not fix the slow narrow-product rescale path.
- **Power-of-ten chunking only:** the 512-bit scenario fell from roughly 50.9 µs to 1.28 µs, but wide examples sometimes lost time because they still multiplied twice.
- **Both:** obtained the most consistent results for the representative test set, combining the 512-bit improvement with a reduction of wide product preparation.

## Profiles and remaining costs

A focused CPU profile of the candidate on the 512-bit and wide-scale scenarios showed approximately 44% flat samples in `divSmallInPlace`, 20% in `uintwide.MulFull`, and 19% in the Go copy routine (`runtime.duffcopy`). Those figures apply only to this mixed profiling workload. The profile suggests that reducing copies of the large `fixedNum` temporary may be the next area to measure. No extra architecture was introduced for that hypothesis.

## Risks, limits, and next steps

- The experimental fixed arithmetic continues to use a 64-limb intermediate; returning/copying it may cost CPU even without allocations.
- Benchmarks use deterministic operands and a single test host. Repeat on a quieter production-representative machine, preferably with `benchstat` and more samples.
- Verify using **actual** Go 1.27 and the complete race suite. Do not claim Go 1.27 compatibility from this Go 1.23.2 experiment.
- Keep Optimus and the other decimal backends untouched until their independent representation experiment.
- Next: assess whether this optimization is beneficial in actual tax/rule workloads, then consider a bounded specialized operation for small powers of ten in `RescaleRoundWithStatus` **as a separate experiment**.

**Conclusion:** a significant, tested speedup for rounded multiplication of `Dec1024`, with the same public contracts and `0 allocs/op` in measured cases. Suitable as the next experimental baseline, not yet a production migration decision.
