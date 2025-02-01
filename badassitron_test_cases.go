package badassitron

import (
	"github.com/alpacahq/alpacadecimal"
)

var (
	ivaCl   = func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("19"); return d }()
	ivaMx   = alpacadecimal.NewFromInt(16)
	ten     = func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }()
	hundred = func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }()
	z       = alpacadecimal.Zero
)

func completeFlowTestCases() []struct {
	name      string
	input     Detail
	want      Detail
	process   func(*Detail) error
	wantError bool
	explain   string
} {
	return []struct {
		name      string
		input     Detail
		want      Detail
		process   func(*Detail) error
		wantError bool
		explain   string
	}{
		{
			name: "Calculate From unit value",
			input: Detail{
				Qty:       alpacadecimal.NewFromInt(10),
				Uv:        alpacadecimal.NewFromInt(100),
				Discounts: []Discount{{alpacadecimal.NewFromInt(10), AppliesToUnit, Percentual}},
				Taxes:     []TaxDetail{{ivaMx, z, z, AppliesToUnit, FirstStage, Percentual}},
			},
			want: Detail{
				Uv:             alpacadecimal.NewFromInt(100),
				Net:            alpacadecimal.NewFromInt(900),
				NetWd:          alpacadecimal.NewFromInt(1000),
				Brute:          alpacadecimal.NewFromInt(1044),
				BruteWd:        alpacadecimal.NewFromInt(1160),
				Taxes:          []TaxDetail{{ivaMx, alpacadecimal.NewFromInt(144), alpacadecimal.NewFromInt(900), 0, 1, true}},
				DiscountAmount: alpacadecimal.NewFromInt(116),
				DiscountRatio:  alpacadecimal.NewFromInt(10),
				Tax:            alpacadecimal.NewFromInt(144),
				TaxRatio:       ivaMx,
			},
			process: func(det *Detail) error {

				t3 := NewTaxStage(3)
				t2 := NewTaxStage(2)
				t1 := NewTaxStage(1)

				d1 := NewDiscounter()
				br := NewBruteSnapshot()
				db := NewBruteDiscounter()
				tr := NewTaxRater()

				d1.SetNext(t1)
				t1.SetNext(t2)
				t2.SetNext(t3)
				t3.SetNext(br)
				br.SetNext(db)
				db.SetNext(tr)

				return d1.Execute(det)

			},
			wantError: false,
			explain:   "",
		},
		{
			name: "Calculate From Brute",
			input: Detail{
				Qty:       alpacadecimal.NewFromInt(10),
				Brute:     alpacadecimal.NewFromInt(1044),
				Discounts: []Discount{{alpacadecimal.NewFromInt(10), AppliesToUnit, Percentual}},
				Taxes:     []TaxDetail{{ivaMx, z, z, AppliesToUnit, FirstStage, Percentual}},
			},
			want: Detail{
				Uv:             alpacadecimal.NewFromInt(100),
				Net:            alpacadecimal.NewFromInt(900),
				NetWd:          alpacadecimal.NewFromInt(1000),
				Brute:          alpacadecimal.NewFromInt(1044),
				BruteWd:        alpacadecimal.NewFromInt(1160),
				Taxes:          []TaxDetail{{ivaMx, alpacadecimal.NewFromInt(144), alpacadecimal.NewFromInt(900), 0, 1, true}},
				DiscountAmount: alpacadecimal.NewFromInt(116),
				DiscountRatio:  alpacadecimal.NewFromInt(10),
				Tax:            alpacadecimal.NewFromInt(144),
				TaxRatio:       ivaMx,
			},
			process: func(det *Detail) error {
				// Para calcular desde el bruto, generalmente, repito, generalmente se debe primero
				// remove los impuestos de cada stage
				// t stands for taxHandler
				ut3 := NewBruteUntaxer(3)
				ut2 := NewBruteUntaxer(2)
				ut1 := NewBruteUntaxer(1)

				// Lo anterior nos dejara al net, al cual se le debe agregar el monto indicado por los descuentos
				unv := NewNetUnDiscounter()

				// Lo anterior nos dejara al net without discounts, a partir del cual podemos
				// obtener el valor unitario
				unq := NewUnquantifier()

				// De aquí en adelante, solo basta realizar el proceso a patir del valor unitario

				t3 := NewTaxStage(3)
				t2 := NewTaxStage(2)
				t1 := NewTaxStage(1)

				d1 := NewDiscounter()
				br := NewBruteSnapshot()
				db := NewBruteDiscounter()
				tr := NewTaxRater()

				ut3.SetNext(ut2)
				ut2.SetNext(ut1)
				ut1.SetNext(unv)
				unv.SetNext(unq)
				unq.SetNext(d1)
				d1.SetNext(t1)
				t1.SetNext(t2)
				t2.SetNext(t3)
				t3.SetNext(br)
				br.SetNext(db)
				db.SetNext(tr)

				return ut3.Execute(det)

			},
			wantError: false,
			explain:   "",
		},
		{
			name: "Calculate From Net with net zero and discounts",
			input: Detail{
				Qty:       alpacadecimal.NewFromInt(4),
				Brute:     z,
				Net:       z,
				Discounts: []Discount{{hundred, 0, true}},
				Taxes:     []TaxDetail{{ivaMx, z, z, 0, 1, true}},
			},
			want: Detail{},
			process: func(det *Detail) error {
				// t stands for taxHandler
				un := NewNetUnDiscounter()

				return un.Execute(det)

			},
			wantError: true,
			explain:   "should fail because we cant calculate values without discount only from brute",
		},
		{
			name: "Calculate From Brute with brute zero and discounts",
			input: Detail{
				Qty:       alpacadecimal.NewFromInt(4),
				Brute:     z,
				Net:       z,
				Discounts: []Discount{{hundred, 0, true}},
				Taxes:     []TaxDetail{{ivaMx, z, z, 1, 0, true}},
			},
			want: Detail{},
			process: func(det *Detail) error {
				// t stands for taxHandler
				t3 := NewBruteUntaxer(3)
				t2 := NewBruteUntaxer(2)
				t1 := NewBruteUntaxer(1)

				un := NewNetUnDiscounter()

				t3.SetNext(t2)
				t2.SetNext(t1)
				t1.SetNext(un)

				return t3.Execute(det)

			},
			wantError: true,
			explain:   "should fail because we cant calculate values without discount only from brute",
		},
	}

}

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
					NewPercentualTax(ivaCl, 1),
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
					NewPercentualTax(ivaCl, 1),
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
					{ten, Unit, true},
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
					{z, Unit, true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
		},
		{
			name: "undiscounting simple percentual 100%",
			args: Detail{
				Uv: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
				Discounts: []Discount{
					{hundred, Unit, true},
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
					{ten, Unit, true},
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
					{ten, Unit, true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
		},
		{
			name: "undiscounting simple percentual 100%",
			args: Detail{
				Net: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
				Discounts: []Discount{
					{ten, Unit, true},
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
					{ten, Unit, true},
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
					{z, Unit, true},
				},
			},
			wantUndiscounted: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
		},
		{
			name: "undiscounting simple percentual 100%",
			args: Detail{
				Brute: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
				Discounts: []Discount{
					{hundred, Unit, true},
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
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("2"); return d }(), Unit, true},
					{ten, Unit, false},
					{ten, Line, false},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("18"); return d }(), alpacadecimal.Zero, alpacadecimal.Zero, Unit, SecondStage, Percentual},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), alpacadecimal.Zero, Unit, FirstStage, Amount},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0.5"); return d }(), alpacadecimal.Zero, Line, ThirdStage, Amount},
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
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("2"); return d }(), Unit, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("-10"); return d }(), Unit, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Line, false},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("-18"); return d }(), alpacadecimal.Zero, alpacadecimal.Zero, Unit, 2, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), alpacadecimal.Zero, Unit, 1, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0.5"); return d }(), alpacadecimal.Zero, Line, 3, false},
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
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(), Unit, true},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("5"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), alpacadecimal.Zero, Unit, 1, true},
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
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("50"); return d }(), Unit, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("500"); return d }(), Line, false},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("5"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), alpacadecimal.Zero, Unit, 1, true},
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
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("70"); return d }(), Unit, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("500"); return d }(), Line, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("500"); return d }(), Line, false},
				},
				Taxes: []TaxDetail{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("5"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), alpacadecimal.Zero, Unit, 1, true},
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
