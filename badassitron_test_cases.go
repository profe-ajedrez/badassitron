package badassitron

import (
	"github.com/alpacahq/alpacadecimal"
	"github.com/profe-ajedrez/badassitron/internal"
)

func unitRndrCases() []struct {
	name      string
	detail    Detail
	wantUv    alpacadecimal.Decimal
	wantError bool
	explain   string
} {
	return []struct {
		name      string
		detail    Detail
		wantUv    alpacadecimal.Decimal
		wantError bool
		explain   string
	}{
		{
			name: "round to 2 decimals",
			detail: Detail{
				Uv:           func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("12.445763595553833"); return d }(),
				EntryUVScale: 2,
			},
			wantUv:    func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("12.45"); return d }(),
			wantError: false,
			explain:   "",
		},
		{
			name: "round to 8 decimals coactionated to 7",
			detail: Detail{
				Uv:           func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("12.445763595553833"); return d }(),
				EntryUVScale: 8,
			},
			wantUv:    func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("12.4457636"); return d }(),
			wantError: false,
			explain:   "",
		},
		{
			name: "round to 15 decimals",
			detail: Detail{
				Uv: func() alpacadecimal.Decimal {
					d, _ := alpacadecimal.NewFromString("12.4457635955538336432874567836")
					return d
				}(),
				EntryUVScale: 16,
			},
			wantUv:    func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("12.4457635955538336"); return d }(),
			wantError: false,
			explain:   "",
		},
	}
}

func bruteUntaxerCases() []struct {
	name        string
	detail      Detail
	taxStage    int8
	wantUntaxed alpacadecimal.Decimal
	wantError   bool
	explain     string
} {
	return []struct {
		name        string
		detail      Detail
		taxStage    int8
		wantUntaxed alpacadecimal.Decimal
		wantError   bool
		explain     string
	}{
		{
			"untax simple percentual",
			Detail{
				Brute: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1190"); return d }(),
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("19"); return d }(), Unit, internal.Zero, internal.Zero, 1, true},
				},
			},
			1,
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
			false,
			"",
		},
		{
			"untax simple percentual",
			Detail{
				Brute: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1190"); return d }(),
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("19"); return d }(), Unit, internal.Zero, internal.Zero, 1, true},
				},
			},
			1,
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
			false,
			"",
		},
	}
}

func undiscounterCases() []struct {
	name             string
	args             Detail
	wantUndiscounted alpacadecimal.Decimal
	wantError        bool
	explain          string
} {
	return []struct {
		name             string
		args             Detail
		wantUndiscounted alpacadecimal.Decimal
		wantError        bool
		explain          string
	}{
		{
			name: "undiscounting simple percentual",
			args: Detail{
				Uv: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal {
				d, _ := alpacadecimal.NewFromString("1111.1111111111111111")
				return d
			}(),
		},
		{
			name: "undiscounting simple percentual zero%",
			args: Detail{
				Uv: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
		},
		{
			name: "undiscounting simple percentual 100%",
			args: Detail{
				Uv: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(), true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
		},
	}
}

func netUndiscounterCases() []struct {
	name             string
	args             Detail
	wantUndiscounted alpacadecimal.Decimal
	wantError        bool
	explain          string
} {
	return []struct {
		name             string
		args             Detail
		wantUndiscounted alpacadecimal.Decimal
		wantError        bool
		explain          string
	}{
		{
			name: "undiscounting simple percentual",
			args: Detail{
				Net: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal {
				d, _ := alpacadecimal.NewFromString("1111.1111111111111111")
				return d
			}(),
		},
		{
			name: "undiscounting simple percentual zero%",
			args: Detail{
				Net: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
		},
		{
			name: "undiscounting simple percentual 100%",
			args: Detail{
				Net: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(), true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
		},
	}

}

func bruteUndiscounterCases() []struct {
	name             string
	args             Detail
	wantUndiscounted alpacadecimal.Decimal
	wantError        bool
	explain          string
} {
	return []struct {
		name             string
		args             Detail
		wantUndiscounted alpacadecimal.Decimal
		wantError        bool
		explain          string
	}{
		{
			name: "undiscounting simple percentual",
			args: Detail{
				Brute: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal {
				d, _ := alpacadecimal.NewFromString("1111.1111111111111111")
				return d
			}(),
		},
		{
			name: "undiscounting simple percentual zero%",
			args: Detail{
				Brute: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
		},
		{
			name: "undiscounting simple percentual 100%",
			args: Detail{
				Brute: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(), true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
		},
	}
}

func multistageTaxes_ExecuteCases() []struct {
	name        string
	dt          *Detail
	wantnet     alpacadecimal.Decimal
	wantnetWd   alpacadecimal.Decimal
	wantbrute   alpacadecimal.Decimal
	wantbruteWd alpacadecimal.Decimal
	wanttax     alpacadecimal.Decimal
	wanttaxWd   alpacadecimal.Decimal
	wantError   bool
	explain     string
} {
	return []struct {
		name        string
		dt          *Detail
		wantnet     alpacadecimal.Decimal
		wantnetWd   alpacadecimal.Decimal
		wantbrute   alpacadecimal.Decimal
		wantbruteWd alpacadecimal.Decimal
		wanttax     alpacadecimal.Decimal
		wanttaxWd   alpacadecimal.Decimal
		wantError   bool
		explain     string
	}{
		{
			"uv 100 | qty 10 | discounts 2% -- 10 unit -- 10 line | tax 18% -- 1 amount -- 0.5 line",
			&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("2"); return d }(), true},
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), false},
					{Line, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), false},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("18"); return d }(), Unit, alpacadecimal.Zero, alpacadecimal.Zero, 2, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), alpacadecimal.Zero, 1, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), Line, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0.5"); return d }(), alpacadecimal.Zero, 3, false},
				},
			},
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("870"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1038.9"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1192.3"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("168.9"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("192.3"); return d }(),
			false,
			"",
		},
		{
			"want error negative tax",
			&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("2"); return d }(), true},
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("-10"); return d }(), false},
					{Line, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), false},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("-18"); return d }(), Unit, alpacadecimal.Zero, alpacadecimal.Zero, 2, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), alpacadecimal.Zero, 1, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), Line, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0.5"); return d }(), alpacadecimal.Zero, 3, false},
				},
			},
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("870"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1037.6"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1191"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("167.6"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("191"); return d }(),
			true,
			"invalid negative discount value",
		},
		{
			"discount 100%",
			&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(), true},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("5"); return d }(), Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), alpacadecimal.Zero, 1, true},
				},
			},
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1050"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("50"); return d }(),
			false,
			"",
		},
		{
			"discount 50% + 500 amount line should be total 0",
			&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("50"); return d }(), true},
					{Line, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("500"); return d }(), false},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("5"); return d }(), Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), alpacadecimal.Zero, 1, true},
				},
			},
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1050"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("50"); return d }(),
			false,
			"",
		},
		{
			"combined discount over 100%",
			&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("70"); return d }(), true},
					{Line, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("500"); return d }(), false},
					{Line, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("500"); return d }(), false},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("5"); return d }(), Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), alpacadecimal.Zero, 1, true},
				},
			},
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1050"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("50"); return d }(),
			false,
			"",
		},
	}
}
