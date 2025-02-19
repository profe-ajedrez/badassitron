package badassitron

import (
	"github.com/profe-ajedrez/badassitron/dec128"
)

var (
	ivaCl   = func() dec128.Dec128 { d, _ := dec128.NewFromString("19"); return d }()
	ivaMx   = dec128.NewFromInt(16)
	ten     = func() dec128.Dec128 { d, _ := dec128.NewFromString("10"); return d }()
	hundred = func() dec128.Dec128 { d, _ := dec128.NewFromString("100"); return d }()
	z       = dec128.Zero
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
				Qty:       dec128.NewFromInt(10),
				Uv:        dec128.NewFromInt(100),
				Discounts: []Discount{{dec128.NewFromInt(10), AppliesToUnit, Percentual}},
				Taxes:     []TaxDetail{{"1", ivaMx, z, z, AppliesToUnit, FirstStage, Percentual}},
			},
			want: Detail{
				Uv:             dec128.NewFromInt(100),
				Net:            dec128.NewFromInt(900),
				NetWd:          dec128.NewFromInt(1000),
				Brute:          dec128.NewFromInt(1044),
				BruteWd:        dec128.NewFromInt(1160),
				Taxes:          []TaxDetail{{"1", ivaMx, dec128.NewFromInt(144), dec128.NewFromInt(900), 0, 1, true}},
				DiscountAmount: dec128.NewFromInt(116),
				DiscountRatio:  dec128.NewFromInt(10),
				Tax:            dec128.NewFromInt(144),
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
				Qty:       dec128.NewFromInt(10),
				Brute:     dec128.NewFromInt(1044),
				Discounts: []Discount{{dec128.NewFromInt(10), AppliesToUnit, Percentual}},
				Taxes:     []TaxDetail{{"1", ivaMx, z, z, AppliesToUnit, FirstStage, Percentual}},
			},
			want: Detail{
				Uv:             dec128.NewFromInt(100),
				Net:            dec128.NewFromInt(900),
				NetWd:          dec128.NewFromInt(1000),
				Brute:          dec128.NewFromInt(1044),
				BruteWd:        dec128.NewFromInt(1160),
				Taxes:          []TaxDetail{{"1", ivaMx, dec128.NewFromInt(144), dec128.NewFromInt(900), 0, 1, true}},
				DiscountAmount: dec128.NewFromInt(116),
				DiscountRatio:  dec128.NewFromInt(10),
				Tax:            dec128.NewFromInt(144),
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
				Qty:       dec128.NewFromInt(4),
				Brute:     z,
				Net:       z,
				Discounts: []Discount{{hundred, 0, true}},
				Taxes:     []TaxDetail{{"1", ivaMx, z, z, 0, 1, true}},
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
				Qty:       dec128.NewFromInt(4),
				Brute:     z,
				Net:       z,
				Discounts: []Discount{{hundred, 0, true}},
				Taxes:     []TaxDetail{{"1", ivaMx, z, z, 1, 0, true}},
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
	wantUv    dec128.Dec128
	wantError bool
	explain   string
} {
	return []struct {
		name      string
		detail    Detail
		wantUv    dec128.Dec128
		wantError bool
		explain   string
	}{
		{
			name: "round to 2 decimals",
			detail: Detail{
				Uv:           func() dec128.Dec128 { d, _ := dec128.NewFromString("12.445763595553833"); return d }(),
				EntryUVScale: 2,
			},
			wantUv:    func() dec128.Dec128 { d, _ := dec128.NewFromString("12.45"); return d }(),
			wantError: false,
			explain:   "",
		},
		{
			name: "round to 8 decimals coactionated to 7",
			detail: Detail{
				Uv:           func() dec128.Dec128 { d, _ := dec128.NewFromString("12.445763595553833"); return d }(),
				EntryUVScale: 8,
			},
			wantUv:    func() dec128.Dec128 { d, _ := dec128.NewFromString("12.4457636"); return d }(),
			wantError: false,
			explain:   "",
		},
		{
			name: "round to 15 decimals",
			detail: Detail{
				Uv: func() dec128.Dec128 {
					d, _ := dec128.NewFromString("12.4457635955538336432")
					return d
				}(),
				EntryUVScale: 16,
			},
			wantUv:    func() dec128.Dec128 { d, _ := dec128.NewFromString("12.4457635955538336"); return d }(),
			wantError: false,
			explain:   "",
		},
	}
}

func bruteUntaxerCases() []struct {
	name        string
	detail      Detail
	taxStage    int8
	wantUntaxed dec128.Dec128
	wantError   bool
	explain     string
} {
	return []struct {
		name        string
		detail      Detail
		taxStage    int8
		wantUntaxed dec128.Dec128
		wantError   bool
		explain     string
	}{
		{
			"untax simple percentual",
			Detail{
				Brute: func() dec128.Dec128 { d, _ := dec128.NewFromString("1190"); return d }(),
				Taxes: []TaxDetail{
					NewPercentualTax(ivaCl, 1),
				},
			},
			1,
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
			false,
			"",
		},
		{
			"untax simple percentual",
			Detail{
				Brute: func() dec128.Dec128 { d, _ := dec128.NewFromString("1190"); return d }(),
				Taxes: []TaxDetail{
					NewPercentualTax(ivaCl, 1),
				},
			},
			1,
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
			false,
			"",
		},
	}
}

func undiscounterCases() []struct {
	name             string
	args             Detail
	wantUndiscounted dec128.Dec128
	wantError        bool
	explain          string
} {
	return []struct {
		name             string
		args             Detail
		wantUndiscounted dec128.Dec128
		wantError        bool
		explain          string
	}{
		{
			name: "undiscounting simple percentual",
			args: Detail{
				Uv: func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{ten, Unit, true},
				},
			},
			wantUndiscounted: func() dec128.Dec128 {
				d, _ := dec128.NewFromString("1111.1111111111111111111")
				return d
			}(),
		},
		{
			name: "undiscounting simple percentual zero%",
			args: Detail{
				Uv: func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{z, Unit, true},
				},
			},
			wantUndiscounted: func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
		},
		{
			name: "undiscounting simple percentual 100%",
			args: Detail{
				Uv: func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
				Discounts: []Discount{
					{hundred, Unit, true},
				},
			},
			wantUndiscounted: func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
		},
	}
}

func netUndiscounterCases() []struct {
	name             string
	args             Detail
	wantUndiscounted dec128.Dec128
	wantError        bool
	explain          string
} {
	return []struct {
		name             string
		args             Detail
		wantUndiscounted dec128.Dec128
		wantError        bool
		explain          string
	}{
		{
			name: "undiscounting simple percentual",
			args: Detail{
				Net: func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{ten, Unit, true},
				},
			},
			wantUndiscounted: func() dec128.Dec128 {
				d, _ := dec128.NewFromString("1111.1111111111111111111")
				return d
			}(),
		},
		{
			name: "undiscounting simple percentual 100%",
			args: Detail{
				Net: func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
				Discounts: []Discount{
					{hundred, Unit, true},
				},
			},
			wantUndiscounted: func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			wantError:        true,
			explain:          "",
		},
	}

}

func bruteUndiscounterCases() []struct {
	name             string
	args             Detail
	wantUndiscounted dec128.Dec128
	wantError        bool
	explain          string
} {
	return []struct {
		name             string
		args             Detail
		wantUndiscounted dec128.Dec128
		wantError        bool
		explain          string
	}{
		{
			name: "undiscounting simple percentual",
			args: Detail{
				Brute: func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{ten, Unit, true},
				},
			},
			wantUndiscounted: func() dec128.Dec128 {
				d, _ := dec128.NewFromString("1111.1111111111111111111")
				return d
			}(),
		},
		{
			name: "undiscounting simple percentual zero%",
			args: Detail{
				Brute: func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
				Discounts: []Discount{
					{z, Unit, true},
				},
			},
			wantUndiscounted: func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
		},
		{
			name: "undiscounting simple percentual 100%",
			args: Detail{
				Brute: func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
				Discounts: []Discount{
					{hundred, Unit, true},
				},
			},
			wantUndiscounted: func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
		},
	}
}

func multistageTaxes_ExecuteCases() []struct {
	name        string
	dt          *Detail
	wantnet     dec128.Dec128
	wantnetWd   dec128.Dec128
	wantbrute   dec128.Dec128
	wantbruteWd dec128.Dec128
	wanttax     dec128.Dec128
	wanttaxWd   dec128.Dec128
	wantError   bool
	explain     string
} {
	return []struct {
		name        string
		dt          *Detail
		wantnet     dec128.Dec128
		wantnetWd   dec128.Dec128
		wantbrute   dec128.Dec128
		wantbruteWd dec128.Dec128
		wanttax     dec128.Dec128
		wanttaxWd   dec128.Dec128
		wantError   bool
		explain     string
	}{
		{
			"uv 100 | qty 10 | discounts 2% -- 10 unit -- 10 line | tax 18% -- 1 amount -- 0.5 line",
			&Detail{
				Uv:  func() dec128.Dec128 { d, _ := dec128.NewFromString("100"); return d }(),
				Qty: func() dec128.Dec128 { d, _ := dec128.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("2"); return d }(), Unit, true},
					{ten, Unit, false},
					{ten, Line, false},
				},
				Taxes: []TaxDetail{
					{"1", func() dec128.Dec128 { d, _ := dec128.NewFromString("18"); return d }(), dec128.Zero, dec128.Zero, Unit, SecondStage, Percentual},
					{"1", func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(), func() dec128.Dec128 { d, _ := dec128.NewFromString("1"); return d }(), dec128.Zero, Unit, FirstStage, Amount},
					{"1", func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(), func() dec128.Dec128 { d, _ := dec128.NewFromString("0.5"); return d }(), dec128.Zero, Line, ThirdStage, Amount},
				},
			},
			func() dec128.Dec128 { d, _ := dec128.NewFromString("870"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1038.9"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1192.3"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("168.9"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("192.3"); return d }(),
			false,
			"",
		},
		{
			"want error negative tax",
			&Detail{
				Uv:  func() dec128.Dec128 { d, _ := dec128.NewFromString("100"); return d }(),
				Qty: func() dec128.Dec128 { d, _ := dec128.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("2"); return d }(), Unit, true},
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("-10"); return d }(), Unit, false},
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("10"); return d }(), Line, false},
				},
				Taxes: []TaxDetail{
					{"1", func() dec128.Dec128 { d, _ := dec128.NewFromString("-18"); return d }(), dec128.Zero, dec128.Zero, Unit, 2, true},
					{"1", func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(), func() dec128.Dec128 { d, _ := dec128.NewFromString("1"); return d }(), dec128.Zero, Unit, 1, false},
					{"1", func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(), func() dec128.Dec128 { d, _ := dec128.NewFromString("0.5"); return d }(), dec128.Zero, Line, 3, false},
				},
			},
			func() dec128.Dec128 { d, _ := dec128.NewFromString("870"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1037.6"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1191"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("167.6"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("191"); return d }(),
			true,
			"invalid negative discount value",
		},
		{
			"discount 100%",
			&Detail{
				Uv:  func() dec128.Dec128 { d, _ := dec128.NewFromString("100"); return d }(),
				Qty: func() dec128.Dec128 { d, _ := dec128.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("100"); return d }(), Unit, true},
				},
				Taxes: []TaxDetail{
					{"1", func() dec128.Dec128 { d, _ := dec128.NewFromString("5"); return d }(), func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(), dec128.Zero, Unit, 1, true},
				},
			},
			func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1050"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("50"); return d }(),
			false,
			"",
		},
		{
			"discount 50% + 500 amount line should be total 0",
			&Detail{
				Uv:  func() dec128.Dec128 { d, _ := dec128.NewFromString("100"); return d }(),
				Qty: func() dec128.Dec128 { d, _ := dec128.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("50"); return d }(), Unit, true},
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("500"); return d }(), Line, false},
				},
				Taxes: []TaxDetail{
					{"1", func() dec128.Dec128 { d, _ := dec128.NewFromString("5"); return d }(), func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(), dec128.Zero, Unit, 1, true},
				},
			},
			func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1050"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("50"); return d }(),
			false,
			"",
		},
		{
			"combined discount over 100%",
			&Detail{
				Uv:  func() dec128.Dec128 { d, _ := dec128.NewFromString("100"); return d }(),
				Qty: func() dec128.Dec128 { d, _ := dec128.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("70"); return d }(), Unit, true},
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("500"); return d }(), Line, false},
					{func() dec128.Dec128 { d, _ := dec128.NewFromString("500"); return d }(), Line, false},
				},
				Taxes: []TaxDetail{
					{"1", func() dec128.Dec128 { d, _ := dec128.NewFromString("5"); return d }(), func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(), dec128.Zero, Unit, 1, true},
				},
			},
			func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1000"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("1050"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("0"); return d }(),
			func() dec128.Dec128 { d, _ := dec128.NewFromString("50"); return d }(),
			false,
			"",
		},
	}
}
