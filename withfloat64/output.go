package withfloat64

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

// WithTaxes implements Outputable.
func (o *Output) WithTaxes([]TaxDetailer) {
	panic("unimplemented")
}

func (o *Output) Unitary() float64 {
	return o.UnitValue
}

func (o *Output) Qty() float64 {
	return o.Quantity
}

func (o *Output) Net() float64 {
	return o.TotalNet
}

func (o *Output) Gross() float64 {
	return o.TotalGross
}

func (o *Output) Tax() float64 {
	return o.TotalTax
}

func (o *Output) Discount() float64 {
	return o.TotalDiscount
}

func (o *Output) GrossDiscount() float64 {
	return o.TotalGrossDiscount
}

func (o *Output) DiscontedUnitary() float64 {
	return o.DiscontedUnitValue
}

func (o *Output) NetWD() float64 {
	return o.TotalNetWD
}

func (o *Output) GrossWD() float64 {
	return o.TotalGrossWD
}

func (o *Output) TaxWD() float64 {
	return o.TotalTaxWD
}

func (o *Output) WithUnitary(uv float64) {
	o.UnitValue = uv
}

func (o *Output) WithQty(qty float64) {
	o.Quantity = qty
}

func (o *Output) WithNet(net float64) {
	o.TotalNet = net
}

func (o *Output) WithGross(gross float64) {
	o.TotalGross = gross
}

func (o *Output) WithTax(tax float64) {
	o.TotalTax = tax
}

func (o *Output) WithDiscount(discount float64) {
	o.TotalDiscount = discount
}

func (o *Output) WithGrossDiscount(grossDiscount float64) {
	o.TotalGrossDiscount = grossDiscount
}

func (o *Output) WithDiscontedUnitary(discountedUnitary float64) {
	o.DiscontedUnitValue = discountedUnitary
}

func (o *Output) WithNetWD(netWD float64) {
	o.TotalNetWD = netWD
}

func (o *Output) WithGrossWD(grossWD float64) {
	o.TotalGrossWD = grossWD
}

func (o *Output) WithTaxWD(taxWD float64) {
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
