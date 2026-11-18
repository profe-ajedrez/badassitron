# GoDoc and executable examples

GoDoc is generated from package comments, exported declarations, and
`Example` functions in `_test.go` files. An example containing an
`// Output:` directive is executed and checked by `go test`.

The links below use the module path in source. A matching public pkg.go.dev
page requires the relevant module version to have been published and
indexed. The source examples work locally without hosted GoDoc.

## Decimal APIs

| Package | API links | Examples |
| --- | --- | --- |
| [optimus](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus) | [NewFromString](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus#NewFromString), [Optimus.Mul](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus#Optimus.Mul), [Optimus.DivRound](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus#Optimus.DivRound), [Optimus.RescaleRoundWithStatus](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus#Optimus.RescaleRoundWithStatus), [SQLDecimal](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/optimus#SQLDecimal) | [Source](../optimus/example_doc01_test.go) |
| [dec128](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec128) | [Dec128](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec128#Dec128), [NewFromString](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec128#NewFromString) | [Source](../dec128/example_doc01_test.go) |
| [dec256](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec256) | [Dec256](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec256#Dec256), [NewFromString](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec256#NewFromString) | [Source](../dec256/example_doc01_test.go) |
| [dec512](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec512) | [Dec512](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec512#Dec512), [NewFromString](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec512#NewFromString) | [Source](../dec512/example_doc01_test.go) |
| [uint128](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec128/uint128) | [Uint128](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec128/uint128#Uint128) | Existing unit tests |
| [uint256](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec256/uint256) | [Uint256](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec256/uint256#Uint256) | Existing unit tests |
| [uint512](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec512/uint512) | [Uint512](https://pkg.go.dev/github.com/profe-ajedrez/badassitron/dec512/uint512#Uint512) | Existing unit tests |

## Runnable GoDoc examples

Optimus has runnable examples for checked parsing, multiplication, rounding,
explicit division, loss detection, exact narrowing, and SQL NULL handling.
The fixed-width packages each have checked parsing and rounding examples.
These complement existing test examples.

```sh
go test ./optimus ./dec128 ./dec256 ./dec512 -run '^Example' -count=1 -v
go test ./... -count=1
go vet ./...
```

## Local GoDoc queries

```sh
go doc github.com/profe-ajedrez/badassitron/optimus
go doc github.com/profe-ajedrez/badassitron/optimus.NewFromString
go doc github.com/profe-ajedrez/badassitron/optimus.Optimus.DivRound
go doc github.com/profe-ajedrez/badassitron/dec512.Dec512
```

For complete programs outside GoDoc, see
[the examples directory](../examples/README.md) and the
[recipe book](RECIPES.md). Each program has a
README with usage, values, and expected output.

## Reproducibility

The source declares Go 1.27. Tests on an isolated copy with a lowered
Go declaration can establish compatibility with an older toolchain but
are not proof of a Go 1.27 release build. Keep the real module declaration
unchanged and record every test environment and command.
