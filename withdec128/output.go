package withdec128

import "github.com/profe-ajedrez/badassitron/dec128"

// Output stores the calculated totals and detail records.
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

// WithTaxes is not implemented by this Output type and panics.
func (o *Output) WithTaxes([]TaxDetailer) {
	panic("unimplemented")
}

// Unitary returns the calculated unit value.
func (o *Output) Unitary() dec128.Dec128 {
	return o.UnitValue
}

// Qty returns the stored quantity.
func (o *Output) Qty() dec128.Dec128 {
	return o.Quantity
}

// Net returns the net total.
func (o *Output) Net() dec128.Dec128 {
	return o.TotalNet
}

// Gross returns the gross total.
func (o *Output) Gross() dec128.Dec128 {
	return o.TotalGross
}

// Tax returns the calculated tax total.
func (o *Output) Tax() dec128.Dec128 {
	return o.TotalTax
}

// Discount returns the stored discount value.
func (o *Output) Discount() dec128.Dec128 {
	return o.TotalDiscount
}

// GrossDiscount returns the gross discount amount.
func (o *Output) GrossDiscount() dec128.Dec128 {
	return o.TotalGrossDiscount
}

// DiscontedUnitary returns the discounted unit value.
func (o *Output) DiscontedUnitary() dec128.Dec128 {
	return o.DiscontedUnitValue
}

// NetWD returns the net total before the line discount.
func (o *Output) NetWD() dec128.Dec128 {
	return o.TotalNetWD
}

// GrossWD returns the gross total before the line discount.
func (o *Output) GrossWD() dec128.Dec128 {
	return o.TotalGrossWD
}

// TaxWD returns the tax total before the line discount.
func (o *Output) TaxWD() dec128.Dec128 {
	return o.TotalTaxWD
}

// WithUnitary updates the stored unitary value.
func (o *Output) WithUnitary(uv dec128.Dec128) {
	o.UnitValue = uv
}

// WithQty updates the stored qty value.
func (o *Output) WithQty(qty dec128.Dec128) {
	o.Quantity = qty
}

// WithNet updates the stored net value.
func (o *Output) WithNet(net dec128.Dec128) {
	o.TotalNet = net
}

// WithGross updates the stored gross value.
func (o *Output) WithGross(gross dec128.Dec128) {
	o.TotalGross = gross
}

// WithTax updates the stored tax value.
func (o *Output) WithTax(tax dec128.Dec128) {
	o.TotalTax = tax
}

// WithDiscount updates the stored discount value.
func (o *Output) WithDiscount(discount dec128.Dec128) {
	o.TotalDiscount = discount
}

// WithGrossDiscount updates the stored gross discount value.
func (o *Output) WithGrossDiscount(grossDiscount dec128.Dec128) {
	o.TotalGrossDiscount = grossDiscount
}

// WithDiscontedUnitary updates the stored discounted unit value value.
func (o *Output) WithDiscontedUnitary(discountedUnitary dec128.Dec128) {
	o.DiscontedUnitValue = discountedUnitary
}

// WithNetWD updates the stored net wd value.
func (o *Output) WithNetWD(netWD dec128.Dec128) {
	o.TotalNetWD = netWD
}

// WithGrossWD updates the stored gross wd value.
func (o *Output) WithGrossWD(grossWD dec128.Dec128) {
	o.TotalGrossWD = grossWD
}

// WithTaxWD updates the stored tax wd value.
func (o *Output) WithTaxWD(taxWD dec128.Dec128) {
	o.TotalTaxWD = taxWD
}

// DetailTaxes returns the detailed taxes.
func (o *Output) DetailTaxes() []TaxDetailer {
	return o.Taxes
}

// DetailDiscount returns the recorded individual discount details.
func (o *Output) DetailDiscount() []DiscountDetailer {
	return o.Discounts
}

var _ Outputable = (*Output)(nil)
