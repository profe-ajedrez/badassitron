package withfloat64

// Output stores the calculated totals and detail records.
type Output struct {
	UnitValue          float64            // Unitary value
	Quantity           float64            // Quantity
	TotalNet           float64            // Net value
	TotalGross         float64            // Gross value
	TotalTax           float64            // Tax value
	TotalDiscount      float64            // Discount value
	TotalGrossDiscount float64            // Gross discount value
	DiscontedUnitValue float64            // Discounted unitary value
	TotalNetWD         float64            // Net with discount value
	TotalGrossWD       float64            // Gross with discount value
	TotalTaxWD         float64            // Tax with discount value
	Taxes              []TaxDetailer      // Detailed taxes
	Discounts          []DiscountDetailer // Detailed discounts
}

// WithTaxes is not implemented by this Output type and panics.
func (o *Output) WithTaxes([]TaxDetailer) {
	panic("unimplemented")
}

// Unitary returns the calculated unit value.
func (o *Output) Unitary() float64 {
	return o.UnitValue
}

// Qty returns the stored quantity.
func (o *Output) Qty() float64 {
	return o.Quantity
}

// Net returns the net total.
func (o *Output) Net() float64 {
	return o.TotalNet
}

// Gross returns the gross total.
func (o *Output) Gross() float64 {
	return o.TotalGross
}

// Tax returns the calculated tax total.
func (o *Output) Tax() float64 {
	return o.TotalTax
}

// Discount returns the stored discount value.
func (o *Output) Discount() float64 {
	return o.TotalDiscount
}

// GrossDiscount returns the gross discount amount.
func (o *Output) GrossDiscount() float64 {
	return o.TotalGrossDiscount
}

// DiscontedUnitary returns the discounted unit value.
func (o *Output) DiscontedUnitary() float64 {
	return o.DiscontedUnitValue
}

// NetWD returns the net total before the line discount.
func (o *Output) NetWD() float64 {
	return o.TotalNetWD
}

// GrossWD returns the gross total before the line discount.
func (o *Output) GrossWD() float64 {
	return o.TotalGrossWD
}

// TaxWD returns the tax total before the line discount.
func (o *Output) TaxWD() float64 {
	return o.TotalTaxWD
}

// WithUnitary updates the stored unitary value.
func (o *Output) WithUnitary(uv float64) {
	o.UnitValue = uv
}

// WithQty updates the stored qty value.
func (o *Output) WithQty(qty float64) {
	o.Quantity = qty
}

// WithNet updates the stored net value.
func (o *Output) WithNet(net float64) {
	o.TotalNet = net
}

// WithGross updates the stored gross value.
func (o *Output) WithGross(gross float64) {
	o.TotalGross = gross
}

// WithTax updates the stored tax value.
func (o *Output) WithTax(tax float64) {
	o.TotalTax = tax
}

// WithDiscount updates the stored discount value.
func (o *Output) WithDiscount(discount float64) {
	o.TotalDiscount = discount
}

// WithGrossDiscount updates the stored gross discount value.
func (o *Output) WithGrossDiscount(grossDiscount float64) {
	o.TotalGrossDiscount = grossDiscount
}

// WithDiscontedUnitary updates the stored discounted unit value value.
func (o *Output) WithDiscontedUnitary(discountedUnitary float64) {
	o.DiscontedUnitValue = discountedUnitary
}

// WithNetWD updates the stored net wd value.
func (o *Output) WithNetWD(netWD float64) {
	o.TotalNetWD = netWD
}

// WithGrossWD updates the stored gross wd value.
func (o *Output) WithGrossWD(grossWD float64) {
	o.TotalGrossWD = grossWD
}

// WithTaxWD updates the stored tax wd value.
func (o *Output) WithTaxWD(taxWD float64) {
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
