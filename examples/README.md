# Decimal examples

These small programs show how to use the decimal packages in Badassitron.
Start with Optimus if you do not need a fixed coefficient size. Each folder
contains a complete program, an explanation, and its expected output.

## Run an example

Use Go 1.27 or newer, as required by the repository's go.mod file. Run the
commands below from the repository root. The examples use the local source
and the Go standard library. They need no database, API key, or extra module.

```sh
go run ./examples/01_optimus_basics
```

To check that all examples compile and pass Go's static checks:

```sh
go test ./examples/...
go vet ./examples/...
```

These commands compile the programs. Use the go run command in each guide
to execute it and compare its output.

## Fastest path

Run `go run ./examples/01_optimus_basics` to get familiar with parsing,
formatting, and checking for errors. If your goal is a business calculation,
follow immediately with `go run ./examples/03_invoice` and read its rounding
policy. The [ten-minute guide](../docs/GETTING_STARTED.md) explains both
without requiring you to study the implementation first.

## Pick an example

1. [Optimus basics](01_optimus_basics/README.md): read values, calculate,
   format output, check errors, and let storage grow.
2. [Rounding](02_rounding/README.md): choose a rounding rule and detect when
   a calculation discards nonzero digits.
3. [Invoice line](03_invoice/README.md): calculate a subtotal, discount,
   tax, and total with a stated rounding policy.
4. [JSON](04_json/README.md): send decimal text and understand what happens
   to trailing zeros and null.
5. [SQL](05_sql/README.md): scan exact nullable values with SQLDecimal and
   use the Dec128 scanner adapter.
6. [Dec128](06_dec128/README.md): use a fixed 128 bit coefficient and explicit
   scales for intermediate calculations.
7. [Dec256](07_dec256/README.md): store larger coefficients, keep more
   fractional digits, and detect overflow.
8. [Dec512](08_dec512/README.md): work with very large or very small values
   and round a square root to 30 decimal places.
9. [Conversions](09_conversions/README.md): move between all four decimal
   packages and reject a conversion that would lose value.
10. [Concurrency](10_concurrency/README.md): share read only inputs and
    calculate with a separate scale in each worker.
11. [Decimal pipeline](11_decimal_pipeline/README.md): run the existing
    withdec128 handlers in the required order.

## Run commands

```sh
go run ./examples/01_optimus_basics
go run ./examples/02_rounding
go run ./examples/03_invoice
go run ./examples/04_json
go run ./examples/05_sql
go run ./examples/06_dec128
go run ./examples/07_dec256
go run ./examples/08_dec512
go run ./examples/09_conversions
go run ./examples/10_concurrency
go run ./examples/11_decimal_pipeline
```

## Terms used in the examples

A coefficient is the whole number stored inside a decimal. The scale says
how many places go after the decimal point. A coefficient of 12345 and a
scale of 2 represent 123.45. Precision() returns that scale. It does not
count significant digits.

Dec128 has a 128 bit coefficient and a maximum scale of 19. Dec256 has a
256 bit coefficient and a maximum scale of 77. Dec512 has a 512 bit
coefficient and a maximum scale of 154. A coefficient limit applies to all
stored digits, including digits before the decimal point.

Optimus chooses a 64, 128, 256, or 512 bit coefficient as needed. Its maximum
scale is 154. It can still run out of capacity. Ordinary Optimus operations
can round when an exact result does not fit. Explicit rounding operations
use the scale you request and return NaN if the result cannot fit.

NaN means the decimal is invalid. Check IsNaN() after arithmetic and read
ErrorDetails() when it is true. NewFromString returns a Go error for invalid
text. A WithStatus method returns a boolean called inexact. True means
nonzero information was discarded. Check IsNaN() separately from this flag.

Use decimal text such as "19.95" or integer constructors for exact input.
Converting from float64 cannot recover decimal digits already lost in the
float. String() removes trailing fractional zeros. StringFixed() keeps the
stored scale. RescaleRound can set a display scale, subject to capacity.

The unsigned uint128, uint256, and uint512 packages store coefficients.
They do not store a decimal point or a sign. The Dec256 example uses
uint256.Max to construct a decimal at the coefficient limit.

The withdec128 package adds an older calculation workflow on top of Dec128.
The parallel withfloat64 package uses binary floating point. The examples
here focus on the decimal components.

See the [Optimus guide](../docs/OPTIMUS_STAGE.md) for more details about
precision and capacity, and the [main README](../README.md) for the package
overview.
