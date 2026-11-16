# Calculate an invoice line

Run this from the repository root:

```sh
go run ./examples/03_invoice
```

Read [main.go](main.go) for the complete program.

This example buys three units at 19.95 each. It applies a ten percent
discount, then an eighteen percent tax to the discounted amount.

The application chooses this policy: round the discount and tax separately
to two decimal places using half even rounding. The decimal package supplies
the arithmetic. The rates, order, and rounding points come from this program.
Change those choices to match the rules of your own application.

The calculation follows these steps:

1. Multiply 19.95 by 3 to get a subtotal of 59.85.
2. Multiply the subtotal by 0.10 and round 5.985 to 5.98.
3. Subtract that discount to get a net amount of 53.87.
4. Multiply the net amount by 0.18 and round 9.6966 to 9.70.
5. Add net and tax to get 63.57.

Rates use fractions here. The text "0.10" means ten percent. The older
withdec128 workflow uses 10 for the same percentage, as shown in the
[pipeline example](../11_decimal_pipeline/README.md).

Expected output:

```text
Subtotal: 59.85
Discount: 5.98
Net: 53.87
Tax: 9.70
Total: 63.57
```

Back to the [example index](../README.md) or the
[ten-minute guide](../../docs/GETTING_STARTED.md).
