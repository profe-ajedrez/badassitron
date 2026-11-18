# Parse and calculate with unsigned 128-bit integers

Run from the repository root:

```sh
go run ./examples/12_uint128
```

[View the complete program](main.go).

Unlike the decimal packages, `dec128/uint128` stores a **nonnegative integer**:
no sign and no decimal point. `FromString` accepts `string` or `[]byte` and
returns `(Uint128, errors.Error)`; `errors.None` means success. The example
checks invalid text, a decimal integer larger than 128 bits, and overflowing
addition. Do not use an arithmetic result when its error code is nonzero.

Expected output:

```text
Sum: 18446744073709551641
Parsed bytes: 42 valid: true
Invalid input: true
Input overflow: true
Addition overflow: true
```

[All examples](../README.md) · [Choose a decimal type](../../docs/DECIMAL_TYPES.md)
