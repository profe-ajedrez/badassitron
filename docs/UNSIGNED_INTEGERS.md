# Unsigned integer helpers

The `dec128/uint128`, `dec256/uint256`, and `dec512/uint512` packages hold
**unsigned integers** (no decimal point, no negative sign). Use them when an
integer coefficient or bit operation is needed. For ordinary decimal values,
start with [Optimus](OPTIMUS.md).

## Read and add a 128-bit integer

```go
large, code := uint128.FromString("18446744073709551616")
if code != errors.None {
    log.Fatal(code.Error())
}
sum, code := large.Add(uint128.FromUint64(25))
if code != errors.None {
    log.Fatal(code.Error())
}
fmt.Println(sum) // 18446744073709551641
```

This fragment uses `fmt`, `log`,
`github.com/profe-ajedrez/badassitron/dec128/errors`, and
`github.com/profe-ajedrez/badassitron/dec128/uint128` as shown in the
[complete runnable example](../examples/12_uint128/main.go).

The parser accepts decimal strings or byte slices and returns an `errors.Error`
**code**, not the standard Go `error` interface. `errors.None` indicates
success; invalid decimal input and integer overflow have distinct codes.
Check operation codes rather than assuming arithmetic wraps.

## Widths and limitations

| Package | Coefficient width | Decimal scale? |
| --- | ---: | --- |
| `dec128/uint128` | 128 bits | No |
| `dec256/uint256` | 256 bits | No |
| `dec512/uint512` | 512 bits | No |

The three types have distinct Go representations and are not source-compatible
aliases. Fixed-width arithmetic uses checked error codes where exposed;
bit shifts and bitwise operations follow each package's documented rules.
Some historical conversion methods have side effects: in particular,
`uint128.FromBigInt` may mutate its `*big.Int` argument. Avoid passing a
shared or reusable `big.Int` without copying it first.

[Run the Uint128 example](../examples/12_uint128/README.md) ·
[Decimal type selection](DECIMAL_TYPES.md) · [GoDoc](GODOC_AND_EXAMPLES.md)
