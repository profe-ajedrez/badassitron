# Calculate with Dec128

Run this from the repository root:

```sh
go run ./examples/06_dec128
```

Read [main.go](main.go) for the complete program.

Dec128 stores a fixed 128 bit coefficient and up to 19 fractional places.
Choose it when you want that fixed size. It does not automatically grow to
Dec256 if a result is too large.

The first calculation divides 98.00 by 1.18, multiplies by 2000, then
multiplies by 0.90. Each explicit operation requests 19 fractional places
and half even rounding. The result is finally rounded to two places for
publication. Each returned intermediate is a stored decimal and can already
have lost digits. A later rounding step cannot recover them.

RescaleRoundWithStatus reports that the final change to two places discards
nonzero digits. The result remains valid. Always check IsNaN separately.

The next calculation uses SqrtRound for the exact square root of 2.25.
QuoRem then divides -10.5 by 3. Its quotient is -3 because it truncates
toward zero. The remainder is -1.5 and has the dividend's sign.

The explicit methods make scale and rounding visible at each calculation
boundary. Ordinary Dec128 Div and Sqrt use truncation and a process wide
default precision, which is a different policy.

Expected output:

```text
Published: 149491.53
Final rounding lost digits: true
Square root: 1.50
Quotient: -3
Remainder: -1.5
```
