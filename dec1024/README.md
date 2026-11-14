# Dec1024 (experimental additive backend)

`dec1024` is a signed, fixed-width decimal backend with a **1024-bit unsigned
coefficient** and a **base-10 fractional scale from 0 to 308**. It is a
comparable, self-contained value type; the coefficient is stored inline.

`dec1024/uint1024` supplies 1024-bit unsigned integer arithmetic and exact
2048-bit multiplication via `MulCarry`. Integer arithmetic uses the existing
`internal/uintwide` primitives. Like Dec256 and Dec512, some decimal operations
still use `math/big` through `internal/decimalwide` for exact intermediates,
including out-of-width multiplication, division, square root and powers.
**This is not yet a zero-allocation decimal backend.**

## Example

```go
package main

import (
    "fmt"
    "github.com/profe-ajedrez/badassitron/dec1024"
)

func main() {
    a := dec1024.FromString("123456789.001")
    b := dec1024.FromString("0.18")
    tax, inexact := a.MulRoundWithStatus(b, 2, dec1024.RoundingHalfEven)
    if tax.IsNaN() { panic(tax.ErrorDetails()) }
    fmt.Println(tax, inexact)
}
```

## API notes

- Public methods closely follow `dec512`, with the principal intentional
  difference that scale arguments, `New`'s exponent, `Precision()` and
  `Exponent()` use **`uint16`**, not `uint8`. This is required for scales
  exceeding 255. `MaxPrecision` is 308.
- Exact widening from Dec128, Dec256 and Dec512 is provided via `FromDec128`,
  `FromDec256` and `FromDec512`. Exact narrowing is available through
  `Dec128Exact`, `Dec256Exact` and `Dec512Exact`; narrowing first canonicalizes,
  so trailing fractional zeros may be removed.
- Numeric errors remain represented as NaNs with explicit error details.
  Check `IsNaN()` and `ErrorDetails()` before publishing a result.
- `SetDefaultPrecision` uses a process-wide atomic default, as in Dec512.
  For isolation across workers, prefer explicit rounding methods.
- The existing `optimus`, `dec128`, `dec256`, `dec512` and Iacon APIs are
  unchanged. Dec1024 is not automatically selected by Optimus or Iacon.
- The supplied module retains `go 1.27`. The test results included with this
  experiment were obtained with Go 1.23.2 in a **temporary copy** after
  changing only the module's go directive. Go 1.27 compatibility is unverified.

## Current limits and next steps

1. Compare coefficient storage alternatives for Optimus independently of
   this new backend, preserving original APIs.
2. Profile and implement fixed-width 1024/2048-bit quantization paths where
   exactness and CPU/heap improvements are demonstrated.
3. Add Dec1024 adapters to Iacon only after checking the generic decimal
   contracts and economic goldens.

Use `go test ./dec1024/...`, `go vet ./...` and the benchmarks in
`dec1024/precision_1024_test.go` and
`dec1024/uint1024/precision_1024_test.go`.

## Fixed-width experimental core (DEC1024-FIXED-01)

The experimental calculation path now uses bounded 4096-bit internal limb
arithmetic instead of `math/big`/`internal/decimalwide`. This does not change
`Dec1024`'s public size, scale or methods. The `Uint1024` API deliberately keeps
explicit `big.Int` conversion methods. See `docs/DEC1024_FIXED_01.md` for the
measured CPU regressions, property testing, and why this version should not yet
replace production deployments without workload-specific measurements.
