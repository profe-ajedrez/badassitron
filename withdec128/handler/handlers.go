package handler

import (
	"github.com/profe-ajedrez/badassitron/withdec128"

	"github.com/profe-ajedrez/badassitron/dec128"
)

// EntryValidation validates required input and normalizes out of range discounts before the next step.
func EntryValidation(opts withdec128.CalculationConfiger, input withdec128.Enterable, output withdec128.Outputable, h ...withdec128.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withdec128.ErrNilArgument
	}

	if input.UnitValue().IsNegative() {
		return withdec128.ErrNegativeUnitary
	}

	if input.Qty().IsNegative() {
		return withdec128.ErrNegativeQty
	}

	if input.Qty().Equal(withdec128.Zero()) {
		return withdec128.ErrZeroQty
	}

	if input.Discount().IsNegative() {
		input.SetDiscToZero()
	}

	if input.Discount().GreaterThan(dec128.Decimal100) {
		input.SetDiscToHundred()
	}

	return Next(opts, input, output, h...)
}

// Bootstrap copies starting input values into the calculation output.
func Bootstrap(opts withdec128.CalculationConfiger, input withdec128.Enterable, output withdec128.Outputable, h ...withdec128.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withdec128.ErrNilArgument
	}

	unitValue := input.UnitValue()

	if opts.NormalizeUnitValue() {
		unitValue = opts.UVNormalizer(unitValue)
	}

	output.WithUnitary(unitValue)
	output.WithQty(input.Qty())
	output.WithDiscount(input.Discount())

	return Next(opts, input, output, h...)
}

// Taxer binds input taxes to stages and computes tax totals after net values are ready.
func Taxer(opts withdec128.CalculationConfiger, input withdec128.Enterable, output withdec128.Outputable, h ...withdec128.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withdec128.ErrNilArgument
	}

	stages := withdec128.NewTaxStages()
	detailTaxes := withdec128.NewDetailTaxes()

	for _, tax := range input.Taxes() {
		err := stages.Bind(input.Qty(), tax)
		if err != nil {
			return err
		}

		detailTaxes.Bind(input.Qty(), tax)
	}

	output.WithTaxWD(
		totalTaxes(stages, output.NetWD(), input.Qty()),
	)
	output.WithTax(
		totalTaxes(stages, output.Net(), input.Qty()),
	)

	opts.WithDetailTaxProcessor(detailTaxes)

	return Next(opts, input, output, h...)
}

// Netter calculates net values and line discounts from quantity and unit value.
func Netter(opts withdec128.CalculationConfiger, input withdec128.Enterable, output withdec128.Outputable, h ...withdec128.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withdec128.ErrNilArgument
	}

	netWD := output.Unitary().Mul(output.Qty())
	r := input.Discount().Div(dec128.Decimal100)

	discountRatio := dec128.Decimal1.Sub(r)
	net := netWD.Mul(discountRatio)
	discount := netWD.Sub(net)

	output.WithNetWD(netWD)
	output.WithNet(net)
	output.WithDiscount(discount)
	r = net.Div(input.Qty())
	output.WithDiscontedUnitary(r)

	return Next(opts, input, output, h...)
}

// Grosser combines net and tax values to produce gross totals.
func Grosser(opts withdec128.CalculationConfiger, input withdec128.Enterable, output withdec128.Outputable, h ...withdec128.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withdec128.ErrNilArgument
	}

	output.WithGross(output.Tax().Add(output.Net()))
	output.WithGrossWD(output.TaxWD().Add(output.NetWD()))
	output.WithGrossDiscount(output.GrossWD().Sub(output.Gross()))

	return Next(opts, input, output, h...)
}

// Next calls the next handler with the remaining chain.
func Next(opts withdec128.CalculationConfiger, input withdec128.Enterable, output withdec128.Outputable, h ...withdec128.HandlerFunc) error {
	if h == nil {
		return nil
	}

	if len(h) == 1 {
		return h[0](opts, input, output)
	}

	return h[0](opts, input, output, h[1:]...)
}

func totalTaxes(stages *withdec128.Stages, taxable, qty dec128.Dec128) dec128.Dec128 {
	natural := stages.Natural.Calc(taxable, qty)
	overtax := stages.Overtax.Calc(taxable.Add(natural), qty)
	bypass := stages.Bypass.Calc(taxable, qty)
	return natural.Add(overtax).Add(bypass)
}
