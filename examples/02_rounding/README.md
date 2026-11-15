# Choose a rounding rule

Run this from the repository root:

```sh
go run ./examples/02_rounding
```

Read [main.go](main.go) for the complete program.

Round chooses the nearest value and sends an exact tie away from zero.
RoundBank chooses the nearest value and sends an exact tie to the even
neighbor. This is why 2.5 becomes 3 with Round and 2 with RoundBank.

Trunc discards digits toward zero. RoundDown moves toward negative infinity.
RoundUp moves toward positive infinity. Negative values make the difference
easy to see: truncating -1.239 to two places gives -1.23, while rounding
down gives -1.24.

Explicit rounding methods accept these modes:

- RoundingTowardZero: discard the fraction.
- RoundingAwayFromZero: increase the magnitude if a nonzero fraction remains.
- RoundingFloor: move toward negative infinity.
- RoundingCeil: move toward positive infinity.
- RoundingHalfTowardZero: choose the nearest value, with ties toward zero.
- RoundingHalfAwayFromZero: choose the nearest value, with ties away from zero.
- RoundingHalfEven: choose the nearest value, with ties to an even last digit.

DivRoundWithStatus calculates at the requested scale and reports whether
nonzero information was discarded. One third needs rounding. One quarter
fits exactly at two places. The status is not an error flag. Check IsNaN
before using either result.

Round does not add trailing zeros when the requested scale is larger than
the stored scale. Use RescaleRound followed by StringFixed to print a
specific number of fractional places. Increasing the scale can overflow
the coefficient, so check that result too.

SqrtRoundWithStatus applies the same choice of scale and mode to a square
root. These methods do not require a change to the process wide default
precision.

Expected output:

```text
Half away from zero: 3
Half even: 2
Truncate: -1.23
Floor: -1.24
Ceiling: -1.23
One third: 0.33 inexact: true
One quarter: 0.25 inexact: false
Two decimal places: 12.00
Square root of two: 1.414214 inexact: true
```
