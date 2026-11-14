# Run the decimal calculation pipeline

Run this from the repository root:

```sh
go run ./examples/11_decimal_pipeline
```

Read [main.go](main.go) for the complete program.

The withdec128 package is an older calculation workflow built on Dec128.
It takes input, options, output, and an ordered list of handlers. This
example starts with a unit value of 100.00, a quantity of 10, a discount
of ten percent, and one natural tax of sixteen percent.

This API takes whole percentage values: Disc is 10 for ten percent and the
tax value is 16 for sixteen percent. That differs from the fractional rates
used directly in the [Optimus invoice example](../03_invoice/README.md).

The handlers run in this order:

1. EntryValidation checks the input and adjusts out of range discounts.
2. Bootstrap copies the unit value and quantity to the output.
3. Netter calculates the subtotal, discount, and net amount.
4. Taxer calculates taxes using those net amounts.
5. Grosser adds net and tax to calculate gross totals.

NetWD returns the net amount before discount. Net returns the amount after
discount. The program prints both, then prints the tax and gross total.

These handlers use the existing Dec128 arithmetic. Options.Prec does not
automatically round their totals. The example explicitly rounds each printed
amount to two places with half even rounding.

This chain works with the supplied Output type. Other older API paths are
incomplete: Output.WithTaxes panics, and this chain does not populate a full
itemized tax report. The example uses the aggregate totals and does not call
WithTaxes. For a new calculation with an explicit rounding rule at every
step, see the Optimus invoice example.

Expected output:

```text
Before discount: 1000.00
Discount: 100.00
Net: 900.00
Tax: 144.00
Gross: 1044.00
```
