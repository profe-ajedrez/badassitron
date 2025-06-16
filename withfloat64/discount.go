package withfloat64

type DetailDiscount struct {
	percent    float64
	amount     float64
	rawPercent float64
	net        float64
}

func (d *DetailDiscount) Percent() float64 {
	return d.percent
}

func (d *DetailDiscount) Amount() float64 {
	return d.amount
}

// Net returns the net value for the discount.
func (d *DetailDiscount) Net() float64 {
	return d.net
}

func (d *DetailDiscount) RawPercent() float64 {
	return d.rawPercent
}

func (d *DetailDiscount) WithPercent(v float64) {
	d.percent = v
}

func (d *DetailDiscount) WithAmount(v float64) {
	d.amount = v
}

func (d *DetailDiscount) WithRawPercent(v float64) {
	d.rawPercent = v
}

func (d *DetailDiscount) WithNet(v float64) {
	d.net = v
}

var _ DiscountDetailer = &DetailDiscount{}
