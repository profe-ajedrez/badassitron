package tests

import (
	"errors"
	"testing"

	"github.com/profe-ajedrez/badassitron/internal"
	"github.com/profe-ajedrez/badassitron/withfloat64"
	"github.com/profe-ajedrez/badassitron/withfloat64/handler"
)

func startHandler(opts withfloat64.CalculationConfiger, input withfloat64.Enterable, output withfloat64.Outputable) error {

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

	opt := &withfloat64.Options{
		Prec:             6,
		Process:          withfloat64.FromUV,
		NormUV:           false,
		DetailTaxProcess: withfloat64.NewDetailTaxes(),
	}

	for i, tc := range calcTestCases() {

		output := &withfloat64.Output{}

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

		t.Logf("case %2d Net           : %.6v  ---  Expected: %.6v", i+1, internal.RoundHalfUp(output.Net(), opt.Scale()), internal.RoundHalfUp(tc.expected.Net(), opt.Scale()))
		t.Logf("case %2d NetWD         : %.6v  ---  Expected: %.6v", i+1, internal.RoundHalfUp(output.NetWD(), opt.Scale()), internal.RoundHalfUp(tc.expected.NetWD(), (opt.Scale())))
		t.Logf("case %2d Gross         : %.6v  ---  Expected: %.6v", i+1, internal.RoundHalfUp(output.Gross(), opt.Scale()), internal.RoundHalfUp(tc.expected.Gross(), (opt.Scale())))
		t.Logf("case %2d GrossWD       : %.6v  ---  Expected: %.6v", i+1, internal.RoundHalfUp(output.GrossWD(), opt.Scale()), internal.RoundHalfUp(tc.expected.GrossWD(), (opt.Scale())))
		t.Logf("case %2d Tax           : %.6v  ---  Expected: %.6v", i+1, internal.RoundHalfUp(output.Tax(), opt.Scale()), internal.RoundHalfUp(tc.expected.Tax(), (opt.Scale())))
		t.Logf("case %2d TaxWD         : %.6v  ---  Expected: %.6v", i+1, internal.RoundHalfUp(output.TaxWD(), opt.Scale()), internal.RoundHalfUp(tc.expected.TaxWD(), (opt.Scale())))
		t.Logf("case %2d Discount      : %.6v  ---  Expected: %.6v", i+1, internal.RoundHalfUp(output.Discount(), opt.Scale()), internal.RoundHalfUp(tc.expected.Discount(), (opt.Scale())))
		t.Logf("case %2d Discount Gross: %.6v  ---  Expected: %.6v", i+1, internal.RoundHalfUp(output.GrossDiscount(), opt.Scale()), internal.RoundHalfUp(tc.expected.GrossDiscount(), (opt.Scale())))

	}
}

func BenchmarkCalculation(b *testing.B) {

	opt := &withfloat64.Options{
		Prec:             6,
		Process:          withfloat64.FromUV,
		NormUV:           false,
		DetailTaxProcess: withfloat64.NewDetailTaxes(),
	}

	output := &withfloat64.Output{}

	for _, tc := range calcTestCases() {
		b.ResetTimer()
		b.Run(tc.name, func(b2 *testing.B) {
			for k := 0; k <= b2.N; k++ {

				_ = startHandler(opt, tc.input, output)
			}
		})
	}
}

var _ withfloat64.Outputable = &test_outputed{}

type test_outputed struct {
	unitary          float64
	qty              float64
	net              float64
	gross            float64
	tax              float64
	discount         float64
	grossDiscount    float64
	discontedUnitary float64
	netWD            float64
	grossWD          float64
	taxWD            float64
}

// DetailDiscount implements withfloat64.Outputable.
func (tout *test_outputed) DetailDiscount() []withfloat64.DiscountDetailer {
	panic("unimplemented")
}

// DetailTaxes implements withfloat64.Outputable.
func (tout *test_outputed) DetailTaxes() []withfloat64.TaxDetailer {
	panic("unimplemented")
}

// WithTaxes implements withfloat64.Outputable.
func (tout *test_outputed) WithTaxes([]withfloat64.TaxDetailer) {
	panic("unimplemented")
}

func (tout *test_outputed) Unitary() float64 {
	return tout.unitary
}

func (tout *test_outputed) Qty() float64 {
	return tout.qty
}

func (tout *test_outputed) Net() float64 {
	return tout.net
}

func (tout *test_outputed) Gross() float64 {
	return tout.gross
}

func (tout *test_outputed) Tax() float64 {
	return tout.tax
}

func (tout *test_outputed) Discount() float64 {
	return tout.discount
}

func (tout *test_outputed) GrossDiscount() float64 {
	return tout.grossDiscount
}

func (tout *test_outputed) DiscontedUnitary() float64 {
	return tout.discontedUnitary
}

func (tout *test_outputed) NetWD() float64 {
	return tout.netWD
}

func (tout *test_outputed) GrossWD() float64 {
	return tout.grossWD
}

func (tout *test_outputed) WithDiscontedUnitary(value float64) {
	tout.discontedUnitary = value
}

func (tout *test_outputed) WithDiscount(value float64) {
	tout.discount = value
}

func (tout *test_outputed) TaxWD() float64 {
	return tout.taxWD
}

func (tout *test_outputed) WithGross(value float64) {
	tout.gross = value
}

// WithGrossDiscount implements d128.Output.
func (tout *test_outputed) WithGrossDiscount(v float64) {
	panic("unimplemented")
}

// WithGrossWD implements d128.Output.
func (tout *test_outputed) WithGrossWD(v float64) {
	tout.grossWD = v
}

// WithNet implements d128.Output.
func (tout *test_outputed) WithNet(v float64) {
	tout.net = v
}

// WithNetWD implements d128.Output.
func (tout *test_outputed) WithNetWD(v float64) {
	tout.netWD = v
}

// WithQty implements d128.Output.
func (tout *test_outputed) WithQty(v float64) {
	tout.qty = v
}

// WithTax implements d128.Output.
func (tout *test_outputed) WithTax(v float64) {
	tout.tax = v
}

// WithTaxWD implements d128.Output.
func (tout *test_outputed) WithTaxWD(v float64) {
	tout.taxWD = v
}

// WithUnitary implements d128.Output.
func (tout *test_outputed) WithUnitary(v float64) {
	tout.unitary = v
}

type testCase struct {
	name      string
	expected  withfloat64.Outputable
	input     withfloat64.Enterable
	wantError bool
	err       error
}

func calcTestCases() []testCase {
	return []testCase{
		{
			"Caso 1",
			&test_outputed{
				unitary:          100,
				qty:              10,
				net:              900.0,
				gross:            1044,
				tax:              144,
				discount:         100,
				grossDiscount:    116,
				discontedUnitary: 90,
				netWD:            1000,
				grossWD:          1160,
				taxWD:            160,
			},
			&withfloat64.Input{
				UV:      100,
				QTY:     10,
				Disc:    10,
				TaxList: []*withfloat64.InputTax{{V: 16, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"}},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 2",
			&test_outputed{
				unitary:          100,
				qty:              10,
				net:              900,
				gross:            1158.40,
				tax:              258.4,
				discount:         100,
				grossDiscount:    127.6,
				discontedUnitary: 90,
				netWD:            1000,
				grossWD:          1286,
				taxWD:            286,
			},
			&withfloat64.Input{
				UV:   100,
				QTY:  10,
				Disc: 10,
				TaxList: []*withfloat64.InputTax{
					{V: 16, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
					{V: 10, Typee: withfloat64.Percentual, Stagee: 1, Id: 2, NameValue: "impuesto overtax 10%", CodeValue: "code o 10"},
					{V: 1, Typee: withfloat64.Amount, Stagee: 1, Id: 3, NameValue: "impuesto overtax 1", CodeValue: "code o 1"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 3",
			&test_outputed{
				unitary:          100,
				discontedUnitary: 90,
				net:              900,
				netWD:            1000,
				gross:            1173.4,
				grossWD:          1301,
				discount:         100,
				grossDiscount:    127.6,
				tax:              273.4,
				taxWD:            301,
				qty:              10,
			},
			&withfloat64.Input{
				UV:   100,
				QTY:  10,
				Disc: 10,
				TaxList: []*withfloat64.InputTax{
					{V: 16, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
					{V: 10, Typee: withfloat64.Percentual, Stagee: 1, Id: 2, NameValue: "impuesto overtax 10%", CodeValue: "code o 10"},
					{V: 1, Typee: withfloat64.Amount, Stagee: 1, Id: 3, NameValue: "impuesto overtax 1", CodeValue: "code o 1"},
					{V: 1.5, Typee: withfloat64.Amount, Stagee: 2, Id: 4, NameValue: "impuesto bypass 1", CodeValue: "code b 1"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 4",
			&test_outputed{
				unitary:          8.62068965517241,
				discontedUnitary: 7.758620689655169,
				net:              77.58620689655169,
				netWD:            86.2068965517241,
				gross:            124,
				grossWD:          135,
				discount:         8.62068965517241,
				grossDiscount:    11,
				tax:              46.4137931034483,
				taxWD:            48.7931034482759,
				qty:              10,
			},
			&withfloat64.Input{
				UV:   8.62068965517241,
				QTY:  10,
				Disc: 10,
				TaxList: []*withfloat64.InputTax{
					{V: 16, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
					{V: 10, Typee: withfloat64.Percentual, Stagee: 1, Id: 2, NameValue: "impuesto overtax 10%", CodeValue: "code o 10"},
					{V: 1, Typee: withfloat64.Amount, Stagee: 1, Id: 3, NameValue: "impuesto overtax 1", CodeValue: "code o 1"},
					{V: 1.5, Typee: withfloat64.Amount, Stagee: 2, Id: 4, NameValue: "impuesto bypass 1", CodeValue: "code b 1"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 5",
			&test_outputed{},
			&withfloat64.Input{
				UV:   8.62068965517241,
				QTY:  10,
				Disc: 10,
				TaxList: []*withfloat64.InputTax{
					{V: -1, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural -1%", CodeValue: "code n -1"},
				},
			},
			errorWanted,
			withfloat64.ErrNegativeTax,
		},
		{
			"Caso 6",
			&test_outputed{},
			&withfloat64.Input{

				UV:   8.62068965517241,
				QTY:  10,
				Disc: 10,
				TaxList: []*withfloat64.InputTax{
					{V: -1, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural -1%", CodeValue: "code n -1"},
				},
			},
			errorWanted,
			withfloat64.ErrNegativeTax,
		},
		{
			"Caso 7",
			&test_outputed{},
			&withfloat64.Input{
				UV:   8.62068965517241,
				QTY:  10,
				Disc: 10,
				TaxList: []*withfloat64.InputTax{
					{V: -1, Typee: withfloat64.Percentual, Stagee: 2, Id: 1, NameValue: "impuesto bypass -1%", CodeValue: "code b -1"},
				},
			},
			errorWanted,
			withfloat64.ErrNegativeTax,
		},
		{
			"Caso 8",
			&test_outputed{},
			&withfloat64.Input{
				UV:   -1,
				QTY:  10,
				Disc: 10,
				TaxList: []*withfloat64.InputTax{
					{V: 0, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 0%", CodeValue: "code n 0"},
				},
			},
			errorWanted,
			withfloat64.ErrNegativeUnitary,
		},
		{
			"Caso 9",
			&test_outputed{},
			&withfloat64.Input{
				UV:   1,
				QTY:  -1,
				Disc: 10,
				TaxList: []*withfloat64.InputTax{
					{V: 0, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural -1%", CodeValue: "code n -1"},
				},
			},
			errorWanted,
			withfloat64.ErrNegativeQty,
		},
		{
			"Caso 10",
			&test_outputed{},
			&withfloat64.Input{
				UV:   1,
				QTY:  1,
				Disc: 0,
				TaxList: []*withfloat64.InputTax{
					{V: 120, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 120%", CodeValue: "code n 120"},
				},
			},
			errorWanted,
			withfloat64.ErrTaxOver100,
		},
		{
			"Caso 11",
			&test_outputed{
				unitary:          1,
				discontedUnitary: 1,
				net:              10,
				netWD:            10,
				gross:            11.6,
				grossWD:          11.6,
				discount:         0,
				grossDiscount:    0,
				tax:              1.6,
				taxWD:            1.6,
				qty:              10,
			},
			&withfloat64.Input{
				UV:   1,
				QTY:  10,
				Disc: 0,
				TaxList: []*withfloat64.InputTax{
					{V: 16, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 12",
			&test_outputed{
				unitary:          0,
				discontedUnitary: 0,
				net:              0,
				netWD:            0,
				gross:            0,
				grossWD:          0,
				discount:         0,
				grossDiscount:    0,
				tax:              0,
				taxWD:            0,
				qty:              10,
			},
			&withfloat64.Input{
				UV:   0,
				QTY:  1,
				Disc: 0,
				TaxList: []*withfloat64.InputTax{
					{V: 0, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
				},
			},
			errorUnwanted,
			nil,
		},
		{
			"Caso 13",
			&test_outputed{
				unitary:          10,
				discontedUnitary: 0,
				net:              0,
				netWD:            100,
				gross:            0,
				grossWD:          116,
				discount:         100,
				grossDiscount:    116,
				tax:              0,
				taxWD:            16,
				qty:              10,
			},
			&withfloat64.Input{
				UV:   10,
				QTY:  10,
				Disc: 100,
				TaxList: []*withfloat64.InputTax{
					{V: 16, Typee: withfloat64.Percentual, Stagee: 0, Id: 1, NameValue: "impuesto natural 16%", CodeValue: "code n 16"},
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
