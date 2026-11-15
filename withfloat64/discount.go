package withfloat64

// DetailDiscount stores the recorded values of an applied discount.
type DetailDiscount struct {
	percent    float64
	amount     float64
	rawPercent float64
	net        float64
}

// Percent returns the stored percentage.
func (d *DetailDiscount) Percent() float64 {
	return d.percent
}

// Amount returns the stored amount.
func (d *DetailDiscount) Amount() float64 {
	return d.amount
}

// Net returns the net value for the discount.
func (d *DetailDiscount) Net() float64 {
	return d.net
}

// RawPercent returns the original discount percentage.
func (d *DetailDiscount) RawPercent() float64 {
	return d.rawPercent
}

// WithPercent updates the stored percent value.
func (d *DetailDiscount) WithPercent(v float64) {
	d.percent = v
}

// WithAmount updates the stored amount value.
func (d *DetailDiscount) WithAmount(v float64) {
	d.amount = v
}

// WithRawPercent updates the stored raw percent value.
func (d *DetailDiscount) WithRawPercent(v float64) {
	d.rawPercent = v
}

// WithNet updates the stored net value.
func (d *DetailDiscount) WithNet(v float64) {
	d.net = v
}

var _ DiscountDetailer = &DetailDiscount{}
