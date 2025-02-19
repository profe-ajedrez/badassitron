package badassitron

import (
	"encoding/json"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/internal"
)

// Detail contains the values calculated to make the sale
type Detail struct {
	TaxRatioWd        dec128.Dec128 `json:"taxRatioWd"`
	DiscountNetAmount dec128.Dec128 `json:"discountNetAmount"`
	BruteWd           dec128.Dec128 `json:"bruteWd"`
	Tax               dec128.Dec128 `json:"tax"`
	DiscountRatio     dec128.Dec128 `json:"discountRatio"`
	Net               dec128.Dec128 `json:"net"`
	NetWd             dec128.Dec128 `json:"netWd"`
	Brute             dec128.Dec128 `json:"brute"`
	Qty               dec128.Dec128 `json:"quantity"`
	DiscountAmount    dec128.Dec128 `json:"discountAmount"`
	TaxWd             dec128.Dec128 `json:"taxWd"`
	TaxRatio          dec128.Dec128 `json:"taxRatio"`
	Uv                dec128.Dec128 `json:"unitValue"`
	UvWd              dec128.Dec128 `json:"unitValueWd"`
	DiscountNetRatio  dec128.Dec128 `json:"discountNetRatio"`
	Taxes             []TaxDetail   `json:"taxes"`
	Discounts         []Discount    `json:"discounts"`
	ValuesMaxScale    int8          `json:"valuesMaxScale"`
	EntryUVScale      int8          `json:"entryUvScale"`
}

func (d *Detail) serialize() string {
	sb := internal.GetSB()
	defer internal.PutSB(sb)

	sb.WriteString(`	// Uv unitary value of the product being sold
	Uv dec128.Dec128      = `)
	sb.WriteString(d.Uv.String())
	sb.WriteString(`
	// Qty quantity being sold
	Qty dec128.Dec128     = `)
	sb.WriteString(d.Qty.String())
	sb.WriteString(`
	// Discounts list of applied discounts
	Discounts []Discount  = `)

	if ds, err := json.Marshal(d.Discounts); err != nil {
		sb.WriteString(" coudnt masrhall discounts ")
		sb.WriteString(err.Error())
	} else {
		sb.WriteString(string(ds))
	}

	sb.WriteString(`
	// Taxes detail of applied taxes over the sale
	Taxes []TaxDetail     = `)

	if tx, err := json.Marshal(d.Taxes); err != nil {
		sb.WriteString(" coudnt masrhall taxes ")
		sb.WriteString(err.Error())
	} else {
		sb.WriteString(string(tx))
	}

	sb.WriteString(`
	// Net total value without taxes of the sale. The result of: Uv * Qty - discounts
	Net dec128.Dec128     = `)
	sb.WriteString(d.Net.String())
	sb.WriteString(`
	// NetWd total value without taxes and without discounts of the sale. The result of: Uv * Qty
	NetWd dec128.Dec128   = `)
	sb.WriteString(d.NetWd.String())
	sb.WriteString(`
	// Brute total value including taxes.  net + taxes
	Brute dec128.Dec128   = `)
	sb.WriteString(d.Brute.String())
	sb.WriteString(`
	// BruteWd total value including taxes without discounts. netWd + taxesWd
	BruteWd dec128.Dec128 = `)
	sb.WriteString(d.BruteWd.String())
	sb.WriteString(`
	// Tax value of the taxes being applied considering discounts
	Tax dec128.Dec128     = `)
	sb.WriteString(d.Tax.String())
	sb.WriteString(`
	// TaxRatio percentual ratio of the tax value over the brute
	TaxRatio dec128.Dec128 = `)
	sb.WriteString(d.TaxRatio.String())
	sb.WriteString(`
	// TaxWd value of the taxes being applied without consider discounts
	TaxWd dec128.Dec128    = `)
	sb.WriteString(d.TaxWd.String())
	sb.WriteString(`
	// TaxRatioWd percentual ratio of the tax value over the bruteWd
	TaxRatioWd dec128.Dec128 = `)
	sb.WriteString(d.TaxRatioWd.String())
	sb.WriteString(`
	// DiscountAmount cummulated amount of the discounts applied
	DiscountAmount dec128.Dec128 = `)
	sb.WriteString(d.DiscountAmount.String())
	sb.WriteString(`
	// DiscountRatio percentual ratio of DiscountAmount over Brute
	DiscountRatio dec128.Dec128 = `)
	sb.WriteString(d.DiscountRatio.String())

	return sb.String()
}

func (d *Detail) Add(d2 *Detail) {

	d.TaxRatioWd = d.TaxRatioWd.Add(d2.TaxRatioWd)
	d.DiscountNetAmount = d.DiscountNetAmount.Add(d2.DiscountNetAmount)
	d.BruteWd = d.BruteWd.Add(d2.BruteWd)
	d.Tax = d.Tax.Add(d2.Tax)
	d.DiscountRatio = d.DiscountRatio.Add(d2.DiscountRatio)
	d.Net = d.Net.Add(d2.Net)
	d.NetWd = d.NetWd.Add(d2.NetWd)
	d.Brute = d.Brute.Add(d2.Brute)
	d.Qty = d.Qty.Add(d2.Qty)
	d.DiscountAmount = d.DiscountAmount.Add(d2.DiscountAmount)
	d.TaxWd = d.TaxWd.Add(d2.TaxWd)
	d.TaxRatio = d.TaxRatio.Add(d2.TaxRatio)
	d.Uv = d.Uv.Add(d2.Uv)
	d.UvWd = d.UvWd.Add(d2.UvWd)
	d.DiscountNetRatio = d.DiscountNetRatio.Add(d2.DiscountNetRatio)

	found := false
	index := -1

	for i, tx2 := range d2.Taxes {

		for k, tx1 := range d.Taxes {
			if tx2.ID == tx1.ID {
				index = k
				found = true
				break
			}
		}

		if found {
			d.Taxes[index].Amount = d.Taxes[index].Amount.Add(d2.Taxes[i].Amount)
			d.Taxes[index].Taxable = d.Taxes[index].Taxable.Add(d2.Taxes[i].Taxable)
		} else {
			d.Taxes = append(d.Taxes, tx2)
		}

		found = false
		index = -1
	}

}

func (d *Detail) Reset() {
	d.TaxRatioWd = z
	d.DiscountNetAmount = z
	d.BruteWd = z
	d.Tax = z
	d.DiscountRatio = z
	d.Net = z
	d.NetWd = z
	d.Brute = z
	d.Qty = z
	d.DiscountAmount = z
	d.TaxWd = z
	d.TaxRatio = z
	d.Uv = z
	d.UvWd = z
	d.DiscountNetRatio = z
	d.Taxes = nil
	d.Discounts = nil
}
