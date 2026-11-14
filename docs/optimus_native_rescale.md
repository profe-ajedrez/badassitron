# BADASSITRON-OPT-01 — Optimus fixed-width rescaling

`Optimus.RescaleRoundWithStatus` previously converted every finite input to
`math/big` via `decimalwide.Quantize`, even for values already at the target
scale. It now attempts the existing Dec512 fixed-width rescaling logic first.
If the fixed-width result is valid, it is converted back into the usual
adaptive-width Optimus representation. Any failure, invalid precision,
or invalid rounding mode is handled by the original decimalwide code.

This preserves the seven rounding modes, coefficient sign/width, requested
scale, inexact status, and error semantics; no public type or function signature
has changed. There is no new dependency.

## Evidence and limits

- The test-only legacy oracle calls exactly the previous decimalwide rescale.
- A seeded random suite (8,000 inputs) and an exhaustive cross-product of
  462,672 boundary/rounding combinations compare both paths.
- The full repository tests and `go vet` passed in the available temporary
  Go 1.23.2 validation copy; the targeted differential tests also passed
  under `go test -race`.
- Normal small-operand paths can complete with zero heap allocations; wide
  values may still require the allocation used by Optimus' indirect coefficient.
- No claim is made for Go 1.27 until the project's declared toolchain is used.

To inspect allocation changes, run the new `BenchmarkNativeRescale` suite
against the original baseline and this code with the same Go version, CPU,
`GOMAXPROCS`, `-benchtime`, and `-count`. Optimize by evidence, not by
assumed behavior under other workloads.
