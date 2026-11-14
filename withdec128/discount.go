package withdec128

import "github.com/profe-ajedrez/badassitron/dec128"

// DetailDiscount stores the recorded values of an applied discount.
type DetailDiscount struct {
	percent    dec128.Dec128
	amount     dec128.Dec128
	rawPercent dec128.Dec128
	net        dec128.Dec128
}

// Percent returns the stored percentage.
func (d *DetailDiscount) Percent() dec128.Dec128 {
	return d.percent
}

// Amount returns the stored amount.
func (d *DetailDiscount) Amount() dec128.Dec128 {
	return d.amount
}

// Net returns the net value for the discount.
func (d *DetailDiscount) Net() dec128.Dec128 {
	return d.net
}

// RawPercent returns the original discount percentage.
func (d *DetailDiscount) RawPercent() dec128.Dec128 {
	return d.rawPercent
}

// WithPercent updates the stored percent value.
func (d *DetailDiscount) WithPercent(v dec128.Dec128) {
	d.percent = v
}

// WithAmount updates the stored amount value.
func (d *DetailDiscount) WithAmount(v dec128.Dec128) {
	d.amount = v
}

// WithRawPercent updates the stored raw percent value.
func (d *DetailDiscount) WithRawPercent(v dec128.Dec128) {
	d.rawPercent = v
}

// WithNet updates the stored net value.
func (d *DetailDiscount) WithNet(v dec128.Dec128) {
	d.net = v
}

var _ DiscountDetailer = &DetailDiscount{}
