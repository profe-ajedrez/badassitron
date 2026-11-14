# Convert between decimal packages

Run this from the repository root:

```sh
go run ./examples/09_conversions
```

Read [main.go](main.go) for the complete program.

The four decimal types are separate Go types. Use their conversion methods
instead of going through float64.

The program widens a Dec128 value to Dec256, then to Dec512, then imports it
into Optimus. Dec128Exact returns it to the original package. Equal confirms
that the numeric value is unchanged. Optimus also has FromDec128 and
FromDec256 constructors, plus Dec256Exact and Dec512Exact output methods.

Exact conversion can remove trailing zeros when that makes a smaller scale
possible. The second value has many fractional zeros after 1.2, so it can
still fit Dec128 exactly. Exact here refers to numeric value; it does not
promise to keep display scale or storage width.

The third value has a nonzero digit at fractional place 20. Dec128 supports
at most 19 places, so Dec128Exact returns an error.

When the application permits loss, make that decision visible: call
RescaleRoundWithStatus with a chosen scale and mode, check IsNaN, inspect
inexact, and then call the exact conversion. In this case the rounded value
is zero. A successful conversion after rounding does not mean the original
value was preserved.

Expected output:

```text
Round trip: 123.45
Same value: true
Trailing zeros removed: 1.2
Lossy conversion rejected: true
Rounded conversion: 0
Rounding lost digits: true
```
