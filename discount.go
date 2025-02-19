package badassitron

import (
	"github.com/profe-ajedrez/badassitron/dec128"
)

// Discount is a value which could be substracted to a value to pay
type Discount struct {
	Value      dec128.Dec128
	Applies    AppliesTo
	Percentual bool
}

// GroupDiscounts returns the cummulated values of amount, percentual and line discounts
// The cummulated amount value gets multiplied by the qty.
//
// Example:
//
//	ds := []Discount{
//		{ Unit, dec128.FromString("10"), true },
//		{ Unit, dec128.FromString("7"), true },
//		{ Unit, dec128.FromString("0.8"), false },
//	}
//
//	qty := dec128.FromString("7")
//
//	amount, perc, line := GroupDiscounts(ds, qty)
//	fmt.Printf("cummulated amount * qty: %s   cummulated percentual: %s   cummulated line: %v", amount, perc, line)
//
//	Output:
//	cummulated amount * qty: 5.6   cummulated percentual: 17   cummulated line: 0
func GroupDiscounts(ds []Discount, qty dec128.Dec128) (amount, perc, lineAmount dec128.Dec128) {
	for i := range ds {
		switch {
		case ds[i].Applies == Line:
			lineAmount = lineAmount.Add(ds[i].Value)
		case !ds[i].Percentual:
			amount = amount.Add(ds[i].Value)
			amount = amount.Mul(qty)
		default:
			perc = perc.Add(ds[i].Value)
		}
	}

	return
}

func DiscountBelowZero(amount, percent, line dec128.Dec128) error {
	if percent.LessThan(dec128.Zero) {
		return WrapWithWrappingError(
			ErrNegativeValue,
			"percentual discount="+percent.String())
	}

	if amount.LessThan(dec128.Zero) {
		return WrapWithWrappingError(
			ErrNegativeValue,
			"amount discount="+amount.String())
	}

	if line.LessThan(dec128.Zero) {
		return WrapWithWrappingError(
			ErrNegativeValue,
			"amount line discount="+line.String())
	}

	return nil
}
