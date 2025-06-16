package withdec128

import "github.com/profe-ajedrez/badassitron/dec128"

type Output struct {
	UnitValue          dec128.Dec128      // Unitary value
	Quantity           dec128.Dec128      // Quantity
	TotalNet           dec128.Dec128      // Net value
	TotalGross         dec128.Dec128      // Gross value
	TotalTax           dec128.Dec128      // Tax value
	TotalDiscount      dec128.Dec128      // Discount value
	TotalGrossDiscount dec128.Dec128      // Gross discount value
	DiscontedUnitValue dec128.Dec128      // Discounted unitary value
	TotalNetWD         dec128.Dec128      // Net with discount value
	TotalGrossWD       dec128.Dec128      // Gross with discount value
	TotalTaxWD         dec128.Dec128      // Tax with discount value
	Taxes              []TaxDetailer      // Detailed taxes
	Discounts          []DiscountDetailer // Detailed discounts
}

// WithTaxes implements Outputable.
func (o *Output) WithTaxes([]TaxDetailer) {
	panic("unimplemented")
}

func (o *Output) Unitary() dec128.Dec128 {
	return o.UnitValue
}

func (o *Output) Qty() dec128.Dec128 {
	return o.Quantity
}

func (o *Output) Net() dec128.Dec128 {
	return o.TotalNet
}

func (o *Output) Gross() dec128.Dec128 {
	return o.TotalGross
}

func (o *Output) Tax() dec128.Dec128 {
	return o.TotalTax
}

func (o *Output) Discount() dec128.Dec128 {
	return o.TotalDiscount
}

func (o *Output) GrossDiscount() dec128.Dec128 {
	return o.TotalGrossDiscount
}

func (o *Output) DiscontedUnitary() dec128.Dec128 {
	return o.DiscontedUnitValue
}

func (o *Output) NetWD() dec128.Dec128 {
	return o.TotalNetWD
}

func (o *Output) GrossWD() dec128.Dec128 {
	return o.TotalGrossWD
}

func (o *Output) TaxWD() dec128.Dec128 {
	return o.TotalTaxWD
}

func (o *Output) WithUnitary(uv dec128.Dec128) {
	o.UnitValue = uv
}

func (o *Output) WithQty(qty dec128.Dec128) {
	o.Quantity = qty
}

func (o *Output) WithNet(net dec128.Dec128) {
	o.TotalNet = net
}

func (o *Output) WithGross(gross dec128.Dec128) {
	o.TotalGross = gross
}

func (o *Output) WithTax(tax dec128.Dec128) {
	o.TotalTax = tax
}

func (o *Output) WithDiscount(discount dec128.Dec128) {
	o.TotalDiscount = discount
}

func (o *Output) WithGrossDiscount(grossDiscount dec128.Dec128) {
	o.TotalGrossDiscount = grossDiscount
}

func (o *Output) WithDiscontedUnitary(discountedUnitary dec128.Dec128) {
	o.DiscontedUnitValue = discountedUnitary
}

func (o *Output) WithNetWD(netWD dec128.Dec128) {
	o.TotalNetWD = netWD
}

func (o *Output) WithGrossWD(grossWD dec128.Dec128) {
	o.TotalGrossWD = grossWD
}

func (o *Output) WithTaxWD(taxWD dec128.Dec128) {
	o.TotalTaxWD = taxWD
}

// DetailTaxes returns the detailed taxes.
func (o *Output) DetailTaxes() []TaxDetailer {
	return o.Taxes
}

func (o *Output) DetailDiscount() []DiscountDetailer {
	return o.Discounts
}

var _ Outputable = (*Output)(nil)
