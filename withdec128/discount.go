package withdec128

import "github.com/profe-ajedrez/badassitron/dec128"

type DetailDiscount struct {
	percent    dec128.Dec128
	amount     dec128.Dec128
	rawPercent dec128.Dec128
	net        dec128.Dec128
}

func (d *DetailDiscount) Percent() dec128.Dec128 {
	return d.percent
}

func (d *DetailDiscount) Amount() dec128.Dec128 {
	return d.amount
}

// Net returns the net value for the discount.
func (d *DetailDiscount) Net() dec128.Dec128 {
	return d.net
}

func (d *DetailDiscount) RawPercent() dec128.Dec128 {
	return d.rawPercent
}

func (d *DetailDiscount) WithPercent(v dec128.Dec128) {
	d.percent = v
}

func (d *DetailDiscount) WithAmount(v dec128.Dec128) {
	d.amount = v
}

func (d *DetailDiscount) WithRawPercent(v dec128.Dec128) {
	d.rawPercent = v
}

func (d *DetailDiscount) WithNet(v dec128.Dec128) {
	d.net = v
}

var _ DiscountDetailer = &DetailDiscount{}
