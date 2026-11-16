# Dec128 FASTPATH V4 (uint16 state)

This change optimizes the ordinary `dec128.Add`, `Sub`, `Mul`, and `Compare`
paths without changing the exported Go declarations or the `uint128` numeric
primitives. The fallback paths retain the original arithmetic semantics.

## Internal representation

`Dec128` stores the 128-bit coefficient, decimal scale, and a 16-bit state.
Bit 0 stores the sign (including signed zero); bits 1 through 8 preserve all
256 values of `errors.Error`. The value remains 24 bytes on amd64.
The private field layout changed: unsafe consumers and code inspecting
private fields are outside the exported-API compatibility guarantee.

## Validation

The verified candidate was tested on AMD Ryzen 7 7735HS using Go 1.27.1.
The target's `go test -race ./... -count=1 -timeout=20m -v` and `go vet ./...`
were reported passing. The differential oracle and 256-error-code checks passed.
The corrected uint16 candidate improved 18 of 20 benchmark scenarios relative
to V3. Two regression scenarios were `Compare/DifferentScale` (about 1.9%)
and `Compare/OppositeSign` (about 8.4%); these are retained as known costs.
For the common-case fixed-operand joint benchmark, Badassitron V4 was
approximately 2.413 ns Add, 2.414 ns Sub, 2.622 ns Compare, 2.851 ns Mul;
jokruger v1.4.0 was 2.206, 2.211, 2.842, and 2.633 ns respectively.
These timings are environment-specific and not API guarantees.

## Verification commands

Run on Go 1.27.1 after applying to the current repository:

```sh
GOWORK=off go test ./... -count=1 -timeout=20m
GOWORK=off go test -race ./... -count=1 -timeout=20m
GOWORK=off go vet ./...
GOWORK=off GOMAXPROCS=1 go test ./dec128 -run '^$' \
  -bench '^BenchmarkV4Common/' -benchtime=250ms -count=5 -benchmem
```

Do not overwrite unrelated packages or downgrade `go.mod`.
