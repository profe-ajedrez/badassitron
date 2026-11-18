# Badassitron

**Decimal arithmetic for Go with explicit rounding and bounded precision.**

Badassitron provides an adaptive decimal (`optimus`), fixed-width decimal
types (`dec128`, `dec256`, `dec512`), and unsigned integer helpers. Decimal
arithmetic uses base-10 coefficients rather than converting operands to binary
`float64`.

**[Get started](docs/GETTING_STARTED.md) · [Recipes](docs/RECIPES.md) ·
[Runnable examples](examples/README.md) · [Documentation index](docs/README.md)**

## Install

The source declares **Go 1.27** and depends on
`github.com/jokruger/dec128 v1.4.0`. In an existing Go module:

```sh
go get github.com/profe-ajedrez/badassitron
```

This installs the version Go resolves from the module registry, **not** this
unpublished documentation snapshot. To build this snapshot, extract its source
and work in its directory.

## Calculate a total

```go
package main

import (
    "fmt"
    "log"

    "github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
    price, err := optimus.NewFromString("12.50")
    if err != nil {
        log.Fatal(err)
    }

    total := price.Mul(optimus.FromInt(3)).
        RescaleRound(2, optimus.RoundingHalfEven)
    if total.IsNaN() {
        log.Fatal(total.ErrorDetails())
    }
    fmt.Println(total.StringFixed())
}
```

Output:

```text
37.50
```

`NewFromString` returns a Go error for invalid input; arithmetic operations
usually report failures as `NaN`. Check `IsNaN()` before using a result.
`RescaleRound` sets the stored number of decimal places and chooses how to
round; `StringFixed()` displays that scale.

## Choose a type

| Package | Decimal scale limit | Use it for |
| --- | ---: | --- |
| [`optimus`](optimus) | 154 | Most applications; adaptive 64–512-bit coefficients |
| [`dec128`](dec128) | 19 | Existing fixed-width 128-bit integrations |
| [`dec256`](dec256) | 77 | Larger fixed-width coefficients |
| [`dec512`](dec512) | 154 | The largest fixed-width coefficients supported |

The decimal *scale* is the count of fractional digits stored, not significant
digits. All types have finite coefficient capacity. `optimus` may round some
ordinary arithmetic when an exact result does not fit; fixed-width types
have different overflow and division rules. See [Choosing a type](docs/DECIMAL_TYPES.md)
and [Decimal behavior](docs/DECIMAL_BEHAVIOR.md).

## Learn by example

| I want to… | Example |
| --- | --- |
| Parse text, add, multiply, and format | [Optimus basics](examples/01_optimus_basics/README.md) |
| Round a value or divide with a chosen scale | [Rounding](examples/02_rounding/README.md) |
| Apply a stated discount and tax policy | [Invoice illustration](examples/03_invoice/README.md) |
| Compare per-line rounding with total rounding | [Rounding boundaries](examples/13_rounding_boundaries/README.md) |
| Read or write JSON | [JSON](examples/04_json/README.md) |
| Handle nullable SQL decimals | [SQL](examples/05_sql/README.md) |
| Convert between widths without silent loss | [Conversions](examples/09_conversions/README.md) |
| Parse unsigned 128-bit integers | [Uint128](examples/12_uint128/README.md) |

[All thirteen runnable examples](examples/README.md) cover the remaining
fixed-width types, concurrency, and the older calculation pipeline.

Badassitron is a **numeric library**, not an accounting or tax-compliance
engine. Rates, rounding points, and regulatory rules are supplied by the
calling application. Input originating as `float64` may already have lost
its original decimal digits; prefer text or integer inputs.

## Reference and project information

- [Documentation index](docs/README.md) — guides organized by task
- [GoDoc and API](docs/GODOC_AND_EXAMPLES.md) — exported methods and examples
- [Architecture](docs/ARCHITECTURE.md) — implementation and concurrency limits
- [Third-party notices](THIRD_PARTY_NOTICES.md) and [license](LICENSE)
- [Maintainer release review](docs/RELEASE_STATUS.md) — current verification
  and unresolved historical source-provenance approval

**Source status:** This snapshot is not a published release. Its historical
source provenance and redistribution permissions have not completed the
independent review needed for publication. See [NOTICE](NOTICE.md).
