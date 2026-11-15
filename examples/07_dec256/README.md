# Calculate with Dec256

Run this from the repository root:

```sh
go run ./examples/07_dec256
```

Read [main.go](main.go) for the complete program.

Dec256 stores a fixed 256 bit coefficient and up to 77 fractional places.
The 41 digit integer in this example exceeds a 128 bit coefficient but fits
in Dec256. Adding 25 keeps the value exact.

The next value has a stored scale of 30. StringFixed keeps its final zero.
Scale and coefficient capacity are separate limits. A scale of 77 does not
mean every number with 77 fractional places can fit: the whole coefficient
must still fit in 256 bits.

DivRoundWithStatus rounds two thirds to six fractional places using half
even rounding. Ordinary Div truncates instead, so use the explicit operation
when a specific rounding rule is required.

The last operation constructs a decimal from uint256.Max at scale zero.
New takes a coefficient, scale, and negative flag. Adding one to this maximum
value overflows. Fixed width arithmetic returns NaN instead of growing to
512 bits.

Expected output:

```text
Large total: 10000000000000000000000000000000000000025
Stored decimal: 0.123456789012345678901234567890
Stored scale: 30
Two thirds: 0.666667 inexact: true
Overflow detected: true
```
