package tests

import (
	"errors"
	"testing"

	"github.com/profe-ajedrez/badassitron/withdec128"
	"github.com/profe-ajedrez/badassitron/withdec128/handler"

	"github.com/profe-ajedrez/badassitron/dec128"
)

func startHandler(opts withdec128.CalculationConfiger, input withdec128.Enterable, output withdec128.Outputable) error {

	return handler.Next(
		opts,
		input,
		output,
		handler.EntryValidation,
		handler.Bootstrap,
		handler.Netter,
		handler.Taxer,
		handler.Grosser,
	)
}

func TestCalculation(t *testing.T) {

	opt := &withdec128.Options{
		Prec:             6,
		Process:          withdec128.FromUV,
		NormUV:           false,
		DetailTaxProcess: withdec128.NewDetailTaxes(),
	}

	for i, tc := range calcTestCases() {

		output := &withdec128.Output{}

		if err := startHandler(opt, tc.input, output); err != nil {

			if tc.wantError {
				if !errors.Is(err, tc.err) {
					t.Logf("[ Calculation Fails test case %v ] expecting error %v   --- Got %v", i+1, tc.err, err)
					t.FailNow()
				} else {
					t.Logf("case %2d Got error     : %v  thats correct! because we have expecting it", i+1, tc.err)
					continue
				}

			} else {
				t.Log(err)
				t.FailNow()
			}
		} else if tc.wantError {
			t.Logf("[ Calculation Fails test case %v ] expecting error %v   --- Got nil", i+1, tc.err)
			t.FailNow()
		}

		t.Logf("case %2d Net           : %.6v  ---  Expected: %.6v", i+1, output.Net().RoundHalfAwayFromZero(uint8(opt.Scale())), tc.expected.Net().RoundHalfAwayFromZero(uint8(opt.Scale())))
		t.Logf("case %2d NetWD         : %.6v  ---  Expected: %.6v", i+1, output.NetWD().RoundHalfAwayFromZero(uint8(opt.Scale())), tc.expected.NetWD().RoundHalfAwayFromZero(uint8(opt.Scale())))
		t.Logf("case %2d Gross         : %.6v  ---  Expected: %.6v", i+1, output.Gross().RoundHalfAwayFromZero(uint8(opt.Scale())), tc.expected.Gross().RoundHalfAwayFromZero(uint8(opt.Scale())))
		t.Logf("case %2d GrossWD       : %.6v  ---  Expected: %.6v", i+1, output.GrossWD().RoundHalfAwayFromZero(uint8(opt.Scale())), tc.expected.GrossWD().RoundHalfAwayFromZero(uint8(opt.Scale())))
		t.Logf("case %2d Tax           : %.6v  ---  Expected: %.6v", i+1, output.Tax().RoundHalfAwayFromZero(uint8(opt.Scale())), tc.expected.Tax().RoundHalfAwayFromZero(uint8(opt.Scale())))
		t.Logf("case %2d TaxWD         : %.6v  ---  Expected: %.6v", i+1, output.TaxWD().RoundHalfAwayFromZero(uint8(opt.Scale())), tc.expected.TaxWD().RoundHalfAwayFromZero(uint8(opt.Scale())))
		t.Logf("case %2d Discount      : %.6v  ---  Expected: %.6v", i+1, output.Discount().RoundHalfAwayFromZero(uint8(opt.Scale())), tc.expected.Discount().RoundHalfAwayFromZero(uint8(opt.Scale())))
		t.Logf("case %2d Discount Gross: %.6v  ---  Expected: %.6v", i+1, output.GrossDiscount().RoundHalfAwayFromZero(uint8(opt.Scale())), tc.expected.GrossDiscount().RoundHalfAwayFromZero(uint8(opt.Scale())))

		// if output.Unitary().String() != tc.expected.Unitary().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got Unitary %v   --- expected %v", i+1, output.Unitary(), tc.expected.Unitary())
		// 	t.FailNow()
		// }

		// if output.Qty().String() != tc.expected.Qty().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got Qty %v   --- expected %v", i+1, output.Qty(), tc.expected.Qty())
		// 	t.FailNow()
		// }

		// if output.Net().String() != tc.expected.Net().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got Net %v   --- expected %v", i+1, output.Net(), tc.expected.Net())
		// 	t.FailNow()
		// }

		// if output.Gross().String() != tc.expected.Gross().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got Gross %v   --- expected %v", i+1, output.Gross(), tc.expected.Gross())
		// 	t.FailNow()
		// }

		// if output.Tax().String() != tc.expected.Tax().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got Tax %v   --- expected %v", i+1, output.Tax(), tc.expected.Tax())
		// 	t.FailNow()
		// }

		// if output.Discount().String() != tc.expected.Discount().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got Discount %v   --- expected %v", i+1, output.Discount(), tc.expected.Discount())
		// 	t.FailNow()
		// }

		// if output.GrossDiscount().String() != tc.expected.GrossDiscount().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got GrossDiscount %v   --- expected %v", i+1, output.GrossDiscount(), tc.expected.GrossDiscount())
		// 	t.FailNow()
		// }

		// if output.DiscontedUnitary().String() != tc.expected.DiscontedUnitary().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got DiscontedUnitary %v   --- expected %v", i+1, output.DiscontedUnitary(), tc.expected.DiscontedUnitary())
		// 	t.FailNow()
		// }

		// if output.NetWD().String() != tc.expected.NetWD().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got NetWD %v   --- expected %v", i+1, output.NetWD(), tc.expected.NetWD())
		// 	t.FailNow()
		// }

		// if output.GrossWD().String() != tc.expected.GrossWD().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got GrossWD %v   --- expected %v", i+1, output.GrossWD(), tc.expected.GrossWD())
		// 	t.FailNow()
		// }

		// if output.TaxWD().String() != tc.expected.TaxWD().String() {
		// 	t.Logf("[ Calculation Fails test case %v ] Got TaxWD %v   --- expected %v", i+1, output.TaxWD(), tc.expected.TaxWD())
		// 	t.FailNow()
		// }
	}
}

func BenchmarkCalculation(b *testing.B) {

	opt := &withdec128.Options{
		Prec:             6,
		Process:          withdec128.FromUV,
		NormUV:           false,
		DetailTaxProcess: withdec128.NewDetailTaxes(),
	}

	output := &withdec128.Output{}

	for _, tc := range calcTestCases() {
		b.ResetTimer()
		b.Run(tc.name, func(b2 *testing.B) {
			for k := 0; k <= b2.N; k++ {

				_ = startHandler(opt, tc.input, output)
			}
		})
	}
}

var _ withdec128.Outputable = &test_outputed{}

type test_outputed struct {
	unitary          dec128.Dec128
	qty              dec128.Dec128
	net              dec128.Dec128
	gross            dec128.Dec128
	tax              dec128.Dec128
	discount         dec128.Dec128
	grossDiscount    dec128.Dec128
	discontedUnitary dec128.Dec128
	netWD            dec128.Dec128
	grossWD          dec128.Dec128
	taxWD            dec128.Dec128
}

// DetailDiscount implements withdec128.Outputable.
func (tout *test_outputed) DetailDiscount() []withdec128.DiscountDetailer {
	panic("unimplemented")
}

// DetailTaxes implements withdec128.Outputable.
func (tout *test_outputed) DetailTaxes() []withdec128.TaxDetailer {
	panic("unimplemented")
}

// WithTaxes implements withdec128.Outputable.
func (tout *test_outputed) WithTaxes([]withdec128.TaxDetailer) {
	panic("unimplemented")
}

func (tout *test_outputed) Unitary() dec128.Dec128 {
	return tout.unitary
}

func (tout *test_outputed) Qty() dec128.Dec128 {
	return tout.qty
}

func (tout *test_outputed) Net() dec128.Dec128 {
	return tout.net
}

func (tout *test_outputed) Gross() dec128.Dec128 {
	return tout.gross
}

func (tout *test_outputed) Tax() dec128.Dec128 {
	return tout.tax
}

func (tout *test_outputed) Discount() dec128.Dec128 {
	return tout.discount
}

func (tout *test_outputed) GrossDiscount() dec128.Dec128 {
	return tout.grossDiscount
}

func (tout *test_outputed) DiscontedUnitary() dec128.Dec128 {
	return tout.discontedUnitary
}

func (tout *test_outputed) NetWD() dec128.Dec128 {
	return tout.netWD
}

func (tout *test_outputed) GrossWD() dec128.Dec128 {
	return tout.grossWD
}

func (tout *test_outputed) WithDiscontedUnitary(value dec128.Dec128) {
	tout.discontedUnitary = value
}

func (tout *test_outputed) WithDiscount(value dec128.Dec128) {
	tout.discount = value
}

func (tout *test_outputed) TaxWD() dec128.Dec128 {
	return tout.taxWD
}

func (tout *test_outputed) WithGross(value dec128.Dec128) {
	tout.gross = value
}

// WithGrossDiscount implements d128.Output.
func (tout *test_outputed) WithGrossDiscount(v dec128.Dec128) {
	panic("unimplemented")
}

// WithGrossWD implements d128.Output.
func (tout *test_outputed) WithGrossWD(v dec128.Dec128) {
	tout.grossWD = v
}

// WithNet implements d128.Output.
func (tout *test_outputed) WithNet(v dec128.Dec128) {
	tout.net = v
}

// WithNetWD implements d128.Output.
func (tout *test_outputed) WithNetWD(v dec128.Dec128) {
	tout.netWD = v
}

// WithQty implements d128.Output.
func (tout *test_outputed) WithQty(v dec128.Dec128) {
	tout.qty = v
}

// WithTax implements d128.Output.
func (tout *test_outputed) WithTax(v dec128.Dec128) {
	tout.tax = v
}

// WithTaxWD implements d128.Output.
func (tout *test_outputed) WithTaxWD(v dec128.Dec128) {
	tout.taxWD = v
}

// WithUnitary implements d128.Output.
func (tout *test_outputed) WithUnitary(v dec128.Dec128) {
	tout.unitary = v
}

type testCase struct {
	name      string
	expected  withdec128.Outputable
	input     withdec128.Enterable
	wantError bool
	err       error
}

func calcTestCases() []testCase {
	return []testCase{
		{
			"Caso 1",
			&test_outputed{
				unitary:          withdec128.Hundred(),
				qty:              withdec128.Ten(),
				net:              dec128.FromFloat64(900.0),
				gross:            dec128.FromFloat64(1044),
				tax:              dec128.FromFloat64(144),
				discount:         withdec128.Hundred(),
				grossDiscount:    dec128.FromFloat64(116),
				discontedUnitary: dec128.FromFloat64(90),
				netWD:            dec128.FromFloat64(1000),
				grossWD:          dec128.FromFloat64(1160),
				taxWD:            dec128.FromFloat64(160),
			},
			&withdec128.Input{
				UV:      withdec128.Hundred(),
				QTY:     withdec128.Ten(),
				Disc:    withdec128.Ten(),
				TaxList: []*withdec128.InputTax{{V: dec128.FromFloat64(16), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"}},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 2",
			&test_outputed{
				unitary:          withdec128.Hundred(),
				qty:              withdec128.Ten(),
				net:              dec128.FromFloat64(900),
				gross:            dec128.FromFloat64(1158.40),
				tax:              dec128.FromFloat64(258.4),
				discount:         withdec128.Hundred(),
				grossDiscount:    dec128.FromFloat64(127.6),
				discontedUnitary: dec128.FromFloat64(90),
				netWD:            dec128.FromFloat64(1000),
				grossWD:          dec128.FromFloat64(1286),
				taxWD:            dec128.FromFloat64(286),
			},
			&withdec128.Input{
				UV:   withdec128.Hundred(),
				QTY:  withdec128.Ten(),
				Disc: withdec128.Ten(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(16), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
					{V: withdec128.Ten(), Typee: withdec128.Percentual, Stagee: 1, Id: 2, NameValue: "impuesto overtax 10%", CodeValue: "code o 10"},
					{V: withdec128.One(), Typee: withdec128.Amount, Stagee: 1, Id: 3, NameValue: "impuesto overtax 1", CodeValue: "code o 1"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 3",
			&test_outputed{
				unitary:          withdec128.Hundred(),
				discontedUnitary: dec128.FromFloat64(90),
				net:              dec128.FromFloat64(900),
				netWD:            dec128.FromFloat64(1000),
				gross:            dec128.FromFloat64(1173.4),
				grossWD:          dec128.FromFloat64(1301),
				discount:         withdec128.Hundred(),
				grossDiscount:    dec128.FromFloat64(127.6),
				tax:              dec128.FromFloat64(273.4),
				taxWD:            dec128.FromFloat64(301),
				qty:              withdec128.Ten(),
			},
			&withdec128.Input{
				UV:   withdec128.Hundred(),
				QTY:  withdec128.Ten(),
				Disc: withdec128.Ten(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(16), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
					{V: withdec128.Ten(), Typee: withdec128.Percentual, Stagee: 1, Id: 2, NameValue: "impuesto overtax 10%", CodeValue: "code o 10"},
					{V: withdec128.One(), Typee: withdec128.Amount, Stagee: 1, Id: 3, NameValue: "impuesto overtax 1", CodeValue: "code o 1"},
					{V: dec128.FromFloat64(1.5), Typee: withdec128.Amount, Stagee: 2, Id: 4, NameValue: "impuesto bypass 1", CodeValue: "code b 1"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 4",
			&test_outputed{
				unitary:          func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("8.62068965517241"),
				discontedUnitary: func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("7.758620689655169"),
				net:              func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("77.58620689655169"),
				netWD:            func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("86.2068965517241"),
				gross:            func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("124"),
				grossWD:          func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("135"),
				discount:         func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("8.62068965517241"),
				grossDiscount:    func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("11"),
				tax:              func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("46.4137931034483"),
				taxWD:            func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("48.7931034482759"),
				qty:              withdec128.Ten(),
			},
			&withdec128.Input{
				UV:   func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("8.62068965517241"),
				QTY:  withdec128.Ten(),
				Disc: withdec128.Ten(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(16), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
					{V: withdec128.Ten(), Typee: withdec128.Percentual, Stagee: 1, Id: 2, NameValue: "impuesto overtax 10%", CodeValue: "code o 10"},
					{V: withdec128.One(), Typee: withdec128.Amount, Stagee: 1, Id: 3, NameValue: "impuesto overtax 1", CodeValue: "code o 1"},
					{V: dec128.FromFloat64(1.5), Typee: withdec128.Amount, Stagee: 2, Id: 4, NameValue: "impuesto bypass 1", CodeValue: "code b 1"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 5",
			&test_outputed{},
			&withdec128.Input{
				UV:   func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("8.62068965517241"),
				QTY:  withdec128.Ten(),
				Disc: withdec128.Ten(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(-1), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural -1%", CodeValue: "code n -1"},
				},
			},
			errorWanted,
			withdec128.ErrNegativeTax,
		},
		{
			"Caso 6",
			&test_outputed{},
			&withdec128.Input{

				UV:   func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("8.62068965517241"),
				QTY:  withdec128.Ten(),
				Disc: withdec128.Ten(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(-1), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural -1%", CodeValue: "code n -1"},
				},
			},
			errorWanted,
			withdec128.ErrNegativeTax,
		},
		{
			"Caso 7",
			&test_outputed{},
			&withdec128.Input{
				UV:   func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("8.62068965517241"),
				QTY:  withdec128.Ten(),
				Disc: withdec128.Ten(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(-1), Typee: withdec128.Percentual, Stagee: 2, Id: 1, NameValue: "impuesto bypass -1%", CodeValue: "code b -1"},
				},
			},
			errorWanted,
			withdec128.ErrNegativeTax,
		},
		{
			"Caso 8",
			&test_outputed{},
			&withdec128.Input{
				UV:   func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("-1"),
				QTY:  withdec128.Ten(),
				Disc: withdec128.Ten(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(0), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 0%", CodeValue: "code n 0"},
				},
			},
			errorWanted,
			withdec128.ErrNegativeUnitary,
		},
		{
			"Caso 9",
			&test_outputed{},
			&withdec128.Input{
				UV:   withdec128.One(),
				QTY:  dec128.FromInt(-1),
				Disc: withdec128.Ten(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(0), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural -1%", CodeValue: "code n -1"},
				},
			},
			errorWanted,
			withdec128.ErrNegativeQty,
		},
		{
			"Caso 10",
			&test_outputed{},
			&withdec128.Input{
				UV:   withdec128.One(),
				QTY:  withdec128.One(),
				Disc: withdec128.Zero(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(120), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 120%", CodeValue: "code n 120"},
				},
			},
			errorWanted,
			withdec128.ErrTaxOver100,
		},
		{
			"Caso 11",
			&test_outputed{
				unitary:          withdec128.One(),
				discontedUnitary: withdec128.One(),
				net:              withdec128.Ten(),
				netWD:            withdec128.Ten(),
				gross:            func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("11.6"),
				grossWD:          func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("11.6"),
				discount:         withdec128.Zero(),
				grossDiscount:    withdec128.Zero(),
				tax:              func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("1.6"),
				taxWD:            func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("1.6"),
				qty:              withdec128.Ten(),
			},
			&withdec128.Input{
				UV:   withdec128.One(),
				QTY:  withdec128.Ten(),
				Disc: withdec128.Zero(),
				TaxList: []*withdec128.InputTax{
					{V: dec128.FromFloat64(16), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 12",
			&test_outputed{
				unitary:          withdec128.Zero(),
				discontedUnitary: withdec128.Zero(),
				net:              withdec128.Zero(),
				netWD:            withdec128.Zero(),
				gross:            withdec128.Zero(),
				grossWD:          withdec128.Zero(),
				discount:         withdec128.Zero(),
				grossDiscount:    withdec128.Zero(),
				tax:              withdec128.Zero(),
				taxWD:            withdec128.Zero(),
				qty:              withdec128.Ten(),
			},
			&withdec128.Input{
				UV:   withdec128.Zero(),
				QTY:  withdec128.One(),
				Disc: withdec128.Zero(),
				TaxList: []*withdec128.InputTax{
					{V: withdec128.Zero(), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 13",
			&test_outputed{
				unitary:          withdec128.Ten(),
				discontedUnitary: withdec128.Zero(),
				net:              withdec128.Zero(),
				netWD:            func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("100"),
				gross:            withdec128.Zero(),
				grossWD:          func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("116"),
				discount:         withdec128.Hundred(),
				grossDiscount:    func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("116"),
				tax:              withdec128.Zero(),
				taxWD:            func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("16"),
				qty:              withdec128.Ten(),
			},
			&withdec128.Input{
				UV:   withdec128.Ten(),
				QTY:  withdec128.Ten(),
				Disc: withdec128.Hundred(),
				TaxList: []*withdec128.InputTax{
					{V: func(v string) dec128.Dec128 { dd, _ := dec128.NewFromString(v); return dd }("16"), Typee: withdec128.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
				},
			},
			errorUnwanted,
			nil,
		},
	}
}

const (
	errorUnwanted = false
	errorWanted   = true
)
