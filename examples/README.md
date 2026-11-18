# Runnable examples

Each folder contains a small, self-contained Go program, the command to run
it, and its expected output. Run commands from the **repository root** with
the Go version and dependencies declared by `go.mod` (Go 1.27 and imported
Jokruger v1.4.0). No live database, network API, or account is required by
these example programs.

## Start with these

| Task | Program |
| --- | --- |
| Parse, calculate and display | [01 — Optimus basics](01_optimus_basics/README.md) |
| Rounding and inexact results | [02 — Rounding](02_rounding/README.md) |
| Discount, tax, and total | [03 — Invoice illustration](03_invoice/README.md) |
| Per-line rounding versus rounding the sum | [13 — Rounding boundaries](13_rounding_boundaries/README.md) |
| JSON transport | [04 — JSON](04_json/README.md) |
| SQL DECIMAL and NULL | [05 — SQL](05_sql/README.md) |

## More numerical tasks

| Task | Program |
| --- | --- |
| Fixed 128-bit decimal | [06 — Dec128](06_dec128/README.md) |
| Fixed 256-bit decimal | [07 — Dec256](07_dec256/README.md) |
| Fixed 512-bit decimal | [08 — Dec512](08_dec512/README.md) |
| Exact widening and narrowing | [09 — Conversions](09_conversions/README.md) |
| Explicit scales in goroutines | [10 — Concurrency](10_concurrency/README.md) |
| Existing handler pipeline (legacy) | [11 — Pipeline](11_decimal_pipeline/README.md) |
| Unsigned 128-bit parsing and overflow | [12 — Uint128](12_uint128/README.md) |

## Run and verify

```sh
go run ./examples/01_optimus_basics
go run ./examples/13_rounding_boundaries
go run ./examples/12_uint128
go test ./examples/...
go vet ./examples/...
```

To run any other example, replace the folder in `go run ./examples/...`
with its directory from the table. `go test` checks compilation; `go run`
executes the program. For a task-oriented walkthrough, see the
[recipe book](../docs/RECIPES.md).

## Rules worth remembering

- Use decimal **text** or integer constructors when original digits matter;
  `float64` may already have rounded them.
- Use `NewFromString` for checked decimal input. Some historic parsers
  accept an empty string as zero; reject missing required fields yourself.
- Check `IsNaN()` after decimal arithmetic. A `WithStatus` boolean reports
  discarded digits, **not** success or failure.
- Choose the rounding **mode and stage** in your application. The invoice
  example illustrates one policy; it is not tax-compliance guidance.
- `dec128/uint128` is an integer API with `errors.Error` codes, unlike
  `optimus` decimal values.

[Documentation index](../docs/README.md) · [Get started](../docs/GETTING_STARTED.md)
