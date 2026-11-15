# Read and write JSON

Run this from the repository root:

```sh
go run ./examples/04_json
```

Read [main.go](main.go) for the complete program.

Put an Optimus field directly in a Go struct. Its JSON encoder writes the
decimal as a quoted string, such as "123.45". The decoder reads that decimal
text without first passing it through a float64 variable.

The numeric value survives this round trip. The original display scale does
not: the encoder removes trailing fractional zeros. An amount read from
123.4500 therefore comes back with scale 2. If your format needs four places,
set that scale explicitly before displaying it. Store formatting information
separately if it must survive a JSON round trip.

The second input keeps 25 fractional digits. This shows why a string field
is useful for decimal data. The input still needs to fit Optimus capacity;
quoted text does not remove that limit.

In the current Optimus contract, JSON null becomes numeric zero. Use a
nullable application field when a missing amount must be distinct from zero.
Check values before encoding: NaN is encoded as the string "NaN", but the
decimal parser rejects that text when decoding it.

Expected output:

```text
JSON: {"amount":"123.45"}
Same value: true
Restored scale: 2
Long decimal: 0.1234567890123456789012345
Null becomes zero: true
```
