package badassitron

import "github.com/alpacahq/alpacadecimal"

// Discount is a value which could be substracted to a value to pay
type Discount struct {
	Applies    AppliesTo
	Value      alpacadecimal.Decimal
	Percentual bool
}

// GroupDiscounts returns the cummulated values of amount, percentual and line discounts
// The cummulated amount value gets multiplied by the qty.
//
// Example:
//
//	ds := []Discount{
//		{ Unit, alpacadecimal.FromString("10"), true },
//		{ Unit, alpacadecimal.FromString("7"), true },
//		{ Unit, alpacadecimal.FromString("0.8"), false },
//	}
//
//	qty := alpacadecimal.FromString("7")
//
//	amount, perc, line := GroupDiscounts(ds, qty)
//	fmt.Printf("cummulated amount * qty: %s   cummulated percentual: %s   cummulated line: %v", amount, perc, line)
//
//	Output:
//	cummulated amount * qty: 5.6   cummulated percentual: 17   cummulated line: 0
func GroupDiscounts(ds []Discount, qty alpacadecimal.Decimal) (amount, perc, lineAmount alpacadecimal.Decimal) {
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

func DiscountBelowZero(amount, percent, line alpacadecimal.Decimal) error {
	if percent.LessThan(alpacadecimal.Zero) {
		return WrapWithWrappingError(
			ErrNegativeValue,
			"percentual discount="+percent.String())
	}

	if amount.LessThan(alpacadecimal.Zero) {
		return WrapWithWrappingError(
			ErrNegativeValue,
			"amount discount="+amount.String())
	}

	if line.LessThan(alpacadecimal.Zero) {
		return WrapWithWrappingError(
			ErrNegativeValue,
			"amount line discount="+line.String())
	}

	return nil
}
