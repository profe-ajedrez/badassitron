package badassitron

import (
	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/internal"
)

func NewPercentualTax(v dec128.Dec128, stage int8) TaxDetail {
	return TaxDetail{
		Ratio:      v,
		Applies:    AppliesToUnit,
		Amount:     dec128.Zero, // because we dont know the amount yet
		Taxable:    dec128.Zero, // because we dont know the taxable yet
		Stage:      stage,
		Percentual: Percentual,
	}
}

// TaxDetail contains info about a tax to be returned to client
type TaxDetail struct {
	ID         string        `json:"id"`
	Ratio      dec128.Dec128 `json:"ratio"`
	Amount     dec128.Dec128 `json:"amount"`
	Taxable    dec128.Dec128 `json:"taxable"`
	Applies    AppliesTo     `json:"applies"`
	Stage      int8          `json:"stage"`
	Percentual bool          `json:"percentual"`
}

// Tax represents a mandatory payment or charge collected by someone
type Tax struct {
	Value      dec128.Dec128 `json:"value"`
	Applies    AppliesTo     `json:"applies"`
	Percentual bool          `json:"percentual"`
	ApplyOn    int8          `json:"applyOn"`
}

// TaxToTaxDetail converts a [Tax] into a [TaxDetail]
func TaxToTaxDetail(a Tax) TaxDetail {
	b := TaxDetail{
		Applies:    a.Applies,
		Stage:      a.ApplyOn,
		Percentual: a.Percentual,
	}

	if a.Percentual {
		b.Ratio = a.Value
	} else {
		b.Amount = a.Value
	}

	return b
}

// GroupTaxes returns the cummulated values of amount, percentual and line taxes
// The cummulated amount value gets multiplied by the qty.
//
// Example:
//
//	ds := []Tax{
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
func GroupTaxes(ds []TaxDetail, qty dec128.Dec128) (amount, perc, lineAmount dec128.Dec128) {
	for i := range ds {
		switch {
		case ds[i].Applies == Line:
			lineAmount = lineAmount.Add(ds[i].Amount)
		case !ds[i].Percentual:
			amount = amount.Add(ds[i].Amount)
			amount = amount.Mul(qty)
		default:
			perc = perc.Add(ds[i].Ratio)
		}
	}

	return
}

// Bystage returns the [][Tax] which are part of the stage parameter
//
// Example
//
//	taxes := []Tax{
//		{Unit, dec128.FromString("16"), true, 1},
//		{Unit, dec128.FromString("1"), true, 1},
//		{Unit, dec128.FromString("19"), true, 2},
//	}
//
//	taxable = dec128.FromString("100")
//
//	fmt.Println("taxes         ", taxes)
//	fmt.Println("stageredTaxes ", stagedTaxes)
//
//	OUTPUT:
//	taxes         [ { 0, 16, true, 1}, { 0, 1, true, 1 }, { 0, 19, true, 2 } ]
//	stageredTaxes [ { 0, 16, true, 1}, { 0, 1, true, 1 } ]
func TaxesByStage(txs []TaxDetail, stage int8) []TaxDetail {
	stagered := make([]TaxDetail, len(txs))
	j := 0

	for i := 0; i < len(txs); i++ {
		if txs[i].Stage == stage {
			stagered[j] = txs[i]
			j += 1
		}
	}

	if j == 0 {
		stagered = stagered[:0]
	}

	if j > 0 {
		stagered = stagered[0:j]
	}

	return stagered
}

// SetTaxesValues calculates and stores in the elements of detail
// the values of taxable, ratio and/or amount
func SetTaxesValues(stageredTaxes []TaxDetail, taxable dec128.Dec128) {
	for i := range stageredTaxes {
		stageredTaxes[i].Taxable = taxable
		// when percentual, calulate the amount and store it
		if stageredTaxes[i].Amount.Equal(dec128.Zero) && stageredTaxes[i].Percentual && stageredTaxes[i].Ratio.GreaterThan(dec128.Zero) {
			stageredTaxes[i].Amount = internal.Part(taxable, stageredTaxes[i].Ratio)
		}

		// when amount, calculate the percentage ratio from it
		if stageredTaxes[i].Ratio.Equal(dec128.Zero) && !stageredTaxes[i].Percentual && taxable.GreaterThan(dec128.Zero) {
			stageredTaxes[i].Ratio, _ = internal.Percentage(stageredTaxes[i].Amount, taxable)
		}
	}
}
