# Optimus basics

Run this from the repository root:

```sh
go run ./examples/01_optimus_basics
```

Read [main.go](main.go) for the complete program.

The program reads 123.4500 with NewFromString. This constructor returns a
Go error when the text is invalid. Empty text is accepted as zero by the
library, so the example checks for a missing amount before parsing it.

String prints 123.45. StringFixed prints 123.4500 because the stored scale
is four. These are two ways of printing the same numeric value. Equal
compares numeric values even when their scales differ.

Mul and Add return new values. The original amount is still 123.45 after
the calculation. Use the returned value to keep a result.

FromString is a shorter parser that returns NaN for invalid text. Division
by zero also returns NaN. IsNaN checks for this state and ErrorDetails
explains the failure. Use NewFromString when a Go error is easier to handle.

The final addition crosses the largest unsigned 64 bit integer. Optimus
increases its coefficient storage to 128 bits. Width describes storage,
so application rules should use the numeric value rather than this width.

Expected output:

```text
Compact: 123.45
Stored scale: 4
Fixed: 123.4500
Total: 247
Original: 123.45
Same value: true
Invalid input: true
Parser returned an error: true
Division by zero: true
Error details available: true
Large integer: 18446744073709551616
Coefficient bits: 128
```
