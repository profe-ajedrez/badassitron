# Badassitron

**Decimal arithmetic in Go, with explicit rounding when it matters.** Badassitron
provides an adaptive decimal (`optimus`), three fixed-width decimals, and a
separate, older calculation pipeline. Decimal calculations do not use binary
floating-point arithmetic internally.

**New here?** Run the example below, then follow the
[10-minute guide](docs/GETTING_STARTED.md). The [example index](examples/README.md)
contains eleven runnable programs with expected output.

## Install

Go **1.27 or newer** is required by this source tree's `go.mod`.

```sh
go get github.com/profe-ajedrez/badassitron
```

The command above is for an existing Go module. To work on this repository,
clone or unpack it and run commands from its root.

## First calculation

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

`NewFromString` reports parse failures as Go errors; arithmetic reports
failures through `NaN` (`IsNaN` and `ErrorDetails`). For required input,
**check for empty text yourself**: the historical parser treats an empty
string as zero.

`RescaleRound(2, RoundingHalfEven)` explicitly sets two fractional places.
`RoundBank(2)` rounds existing places but does **not** add trailing zero places
when the current scale is smaller. `StringFixed()` preserves the stored scale;
`String()` formats without insignificant trailing fractional zeros.

## Choose a package

| Package | Coefficient | Max fractional scale | Choose it when… |
| --- | ---: | ---: | --- |
| [`optimus`](optimus) | Adaptive 64–512 bits | 154 | You want a general-purpose decimal (recommended starting point). |
| [`dec128`](dec128) | 128 bits | 19 | Fixed coefficient width or existing `Dec128` integration matters. |
| [`dec256`](dec256) | 256 bits | 77 | You need a wider fixed-width value. |
| [`dec512`](dec512) | 512 bits | 154 | You need the widest fixed-width value in this library. |

A stored decimal is a coefficient and decimal *scale*: coefficient `12345` at
scale `2` represents `123.45`. `Precision()` returns the **fractional scale**,
not the number of significant digits. Even `optimus` has finite capacity.

`optimus` can keep exact results when representable and can round ordinary
arithmetic when a result does not fit exactly. The fixed-width types instead
return `NaN` when an exact ordinary arithmetic result cannot fit. Division,
square roots, and explicit rounding have additional policies. See
[decimal behavior and trade-offs](docs/DECIMAL_BEHAVIOR.md).

## Common tasks

| I want to… | Start here |
| --- | --- |
| Parse, calculate, and display a decimal | [Getting started](docs/GETTING_STARTED.md) · [Optimus basics](examples/01_optimus_basics/README.md) |
| Decide where and how to round | [Rounding example](examples/02_rounding/README.md) · [behavior guide](docs/DECIMAL_BEHAVIOR.md) |
| Calculate an invoice line with discount and tax | [Invoice example](examples/03_invoice/README.md) |
| Encode JSON without unintended floating-point conversion | [JSON example](examples/04_json/README.md) |
| Read SQL `DECIMAL`/`NUMERIC` exactly, including NULL | [SQL example](examples/05_sql/README.md) |
| Work with `dec128` | [Dec128 example](examples/06_dec128/README.md) · [precision](dec128/PRECISION.md) |
| Convert between decimal widths | [Conversions](examples/09_conversions/README.md) |
| Use shared decimals from goroutines | [Concurrency example](examples/10_concurrency/README.md) |
| Run the historical tax/discount handler chain | [Pipeline example](examples/11_decimal_pipeline/README.md) |

**Two important boundaries:** start with decimal strings or integers rather
than `float64` when original decimal digits must be preserved. `WithStatus`
methods report whether **nonzero digits were discarded**; the boolean is not
an error indicator. Always check the result for `NaN` as well.

## API reference and deeper reading

- [Documentation map](docs/README.md) — quick answers versus technical detail.
- [Optimus guide](docs/OPTIMUS_STAGE.md) — representation, precision, errors,
  conversions, SQL, and concurrency.
- [GoDoc for Optimus](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus)
  and [GoDoc for Dec128](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec128).
- [Dec128 SQL adapter](dec128/SQL_SCANNING.md) — why its historical `Scan`
  method is not a direct destination for `rows.Scan`.
- [Calculation pipeline](docs/CALCULATION_PIPELINE.md) — separate from the
  decimal arithmetic APIs; some older paths are incomplete.
- [Dec128 FASTPATH V4](docs/FASTPATH_V4.md) — implementation notes and measured
  results, **not** additional API requirements.
- [Changelog](CHANGELOG.md) — source history; this snapshot does not itself
  establish a published release tag.

## Verify a checkout

```sh
GOWORK=off go test ./... -count=1
GOWORK=off go test -race ./... -count=1 -timeout=20m
GOWORK=off go vet ./...
python3 tools/check_doc_links.py
```

These commands test the current checkout; the race detector exercises only
paths covered by tests. To run the examples, use `go run ./examples/03_invoice`
(or another folder listed in the [examples index](examples/README.md)).

Badassitron supplies numeric primitives, **not** a jurisdiction-specific money
or tax policy. Your application must define rounding points, rates, and
acceptable precision. The project is distributed under [LICENSE](LICENSE).
