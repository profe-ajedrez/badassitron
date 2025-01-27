package badassitron

import (
	"encoding/json"

	"github.com/alpacahq/alpacadecimal"
	"github.com/profe-ajedrez/badassitron/internal"
)

// Detail contains the values calculated to make the sale
type Detail struct {
	// Uv unitary value of the product being sold
	Uv alpacadecimal.Decimal `json:"unitValue"`

	// UvWd unitary value without discount of the product being sold
	UvWd alpacadecimal.Decimal `json:"unitValueWd"`

	// Qty quantity being sold
	Qty alpacadecimal.Decimal `json:"quantity"`

	// Discounts list of applied discounts
	Discounts []Discount `json:"discounts"`

	// Taxes detail of applied taxes over the sale
	Taxes []TaxDetail `json:"taxes"`

	// Net total value without taxes of the sale. The result of: Uv * Qty - discounts
	Net alpacadecimal.Decimal `json:"net"`

	// NetWd total value without taxes and without discounts of the sale. The result of: Uv * Qty
	NetWd alpacadecimal.Decimal `json:"netWd"`

	// Brute total value including taxes.  net + taxes
	Brute alpacadecimal.Decimal `json:"brute"`

	// BruteWd total value including taxes without discounts. netWd + taxesWd
	BruteWd alpacadecimal.Decimal `json:"bruteWd"`

	// Tax value of the taxes being applied considering discounts
	Tax alpacadecimal.Decimal `json:"tax"`

	// TaxRatio percentual ratio of the tax value over the brute
	TaxRatio alpacadecimal.Decimal `json:"taxRatio"`

	// TaxWd value of the taxes being applied without consider discounts
	TaxWd alpacadecimal.Decimal `json:"taxWd"`

	// TaxRatioWd percentual ratio of the tax value over the bruteWd
	TaxRatioWd alpacadecimal.Decimal `json:"taxRatioWd"`

	// DiscountAmount cummulated amount of the discounts applied
	DiscountAmount alpacadecimal.Decimal `json:"discountAmount"`
	// DiscountRatio percentual ratio of DiscountAmount over Brute
	DiscountRatio alpacadecimal.Decimal `json:"discountRatio"`

	EntryUVScale int8 `json:"entryUvScale"`

	ValuesMaxScale int8 `json:"valuesMaxScale"`
}

func (d *Detail) serialize() string {
	sb := internal.GetSB()
	defer internal.PutSB(sb)

	sb.WriteString(`	// Uv unitary value of the product being sold
	Uv alpacadecimal.Decimal      = `)
	sb.WriteString(d.Uv.String())
	sb.WriteString(`
	// Qty quantity being sold
	Qty alpacadecimal.Decimal     = `)
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
	Net alpacadecimal.Decimal     = `)
	sb.WriteString(d.Net.String())
	sb.WriteString(`
	// NetWd total value without taxes and without discounts of the sale. The result of: Uv * Qty
	NetWd alpacadecimal.Decimal   = `)
	sb.WriteString(d.NetWd.String())
	sb.WriteString(`
	// Brute total value including taxes.  net + taxes
	Brute alpacadecimal.Decimal   = `)
	sb.WriteString(d.Brute.String())
	sb.WriteString(`
	// BruteWd total value including taxes without discounts. netWd + taxesWd
	BruteWd alpacadecimal.Decimal = `)
	sb.WriteString(d.BruteWd.String())
	sb.WriteString(`
	// Tax value of the taxes being applied considering discounts
	Tax alpacadecimal.Decimal     = `)
	sb.WriteString(d.Tax.String())
	sb.WriteString(`
	// TaxRatio percentual ratio of the tax value over the brute
	TaxRatio alpacadecimal.Decimal = `)
	sb.WriteString(d.TaxRatio.String())
	sb.WriteString(`
	// TaxWd value of the taxes being applied without consider discounts
	TaxWd alpacadecimal.Decimal    = `)
	sb.WriteString(d.TaxWd.String())
	sb.WriteString(`
	// TaxRatioWd percentual ratio of the tax value over the bruteWd
	TaxRatioWd alpacadecimal.Decimal = `)
	sb.WriteString(d.TaxRatioWd.String())
	sb.WriteString(`
	// DiscountAmount cummulated amount of the discounts applied
	DiscountAmount alpacadecimal.Decimal = `)
	sb.WriteString(d.DiscountAmount.String())
	sb.WriteString(`
	// DiscountRatio percentual ratio of DiscountAmount over Brute
	DiscountRatio alpacadecimal.Decimal = `)
	sb.WriteString(d.DiscountRatio.String())

	return sb.String()
}
