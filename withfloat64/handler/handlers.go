package handler

import (
	"github.com/profe-ajedrez/badassitron/withfloat64"
)

func EntryValidation(opts withfloat64.CalculationConfiger, input withfloat64.Enterable, output withfloat64.Outputable, h ...withfloat64.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withfloat64.ErrNilArgument
	}

	if input.UnitValue() < 0 {
		return withfloat64.ErrNegativeUnitary
	}

	if input.Qty() < 0 {
		return withfloat64.ErrNegativeQty
	}

	if input.Qty() == 0 {
		return withfloat64.ErrZeroQty
	}

	if input.Discount() < 0 {
		input.SetDiscToZero()
	}

	if input.Discount() > 100 {
		input.SetDiscToHundred()
	}

	return Next(opts, input, output, h...)
}

func Bootstrap(opts withfloat64.CalculationConfiger, input withfloat64.Enterable, output withfloat64.Outputable, h ...withfloat64.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withfloat64.ErrNilArgument
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

func Taxer(opts withfloat64.CalculationConfiger, input withfloat64.Enterable, output withfloat64.Outputable, h ...withfloat64.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withfloat64.ErrNilArgument
	}

	stages := withfloat64.NewTaxStages()
	detailTaxes := withfloat64.NewDetailTaxes()

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

func Netter(opts withfloat64.CalculationConfiger, input withfloat64.Enterable, output withfloat64.Outputable, h ...withfloat64.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withfloat64.ErrNilArgument
	}

	netWD := output.Unitary() * output.Qty()
	r := input.Discount() / 100

	discountRatio := 1 - r
	net := netWD * discountRatio
	discount := netWD - net

	output.WithNetWD(netWD)
	output.WithNet(net)
	output.WithDiscount(discount)
	r = net / input.Qty()
	output.WithDiscontedUnitary(r)

	return Next(opts, input, output, h...)
}

func Grosser(opts withfloat64.CalculationConfiger, input withfloat64.Enterable, output withfloat64.Outputable, h ...withfloat64.HandlerFunc) error {
	if opts == nil || input == nil || output == nil {
		return withfloat64.ErrNilArgument
	}

	output.WithGross(output.Tax() + output.Net())
	output.WithGrossWD(output.TaxWD() + output.NetWD())
	output.WithGrossDiscount(output.GrossWD() - output.Gross())

	return Next(opts, input, output, h...)
}

func Next(opts withfloat64.CalculationConfiger, input withfloat64.Enterable, output withfloat64.Outputable, h ...withfloat64.HandlerFunc) error {
	if h == nil {
		return nil
	}

	if len(h) == 1 {
		return h[0](opts, input, output)
	}

	return h[0](opts, input, output, h[1:]...)
}

func totalTaxes(stages *withfloat64.Stages, taxable, qty float64) float64 {
	natural := stages.Natural.Calc(taxable, qty)
	overtax := stages.Overtax.Calc(taxable+natural, qty)
	bypass := stages.Bypass.Calc(taxable, qty)
	return natural + overtax + bypass
}
