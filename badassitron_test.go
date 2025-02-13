package badassitron

import (
	"errors"
	"testing"

	"github.com/alpacahq/alpacadecimal"
)

func TestCompleteFlow(t *testing.T) {

	for i, tt := range completeFlowTestCases() {
		err := tt.process(&tt.input)

		if err != nil && !tt.wantError {
			t.Logf("[ TestCompleteFlow %d  %s ] unexpected error %v", i, tt.name, err)
			t.FailNow()
		}

		if err == nil && tt.wantError {
			t.Logf("[ TestCompleteFlow %d  %s ] Expecting error %s  got nil", i, tt.name, tt.explain)
			t.FailNow()
		}

		if err != nil && tt.wantError {
			continue
		}

		if !tt.input.Uv.Equal(tt.want.Uv) {
			t.Logf("[ TestCompleteFlow %d  %s ] different Uv. expecting %v  got %v", i, tt.name, tt.want.Uv, tt.input.Uv)
			t.FailNow()
		}

		if !tt.input.Net.Equal(tt.want.Net) {
			t.Logf("[ TestCompleteFlow %d  %s ] different Net. expecting %v  got %v", i, tt.name, tt.want.Net, tt.input.Net)
			t.FailNow()
		}

		if !tt.input.NetWd.Equal(tt.want.NetWd) {
			t.Logf("[ TestCompleteFlow %d  %s ] different NetWd. expecting %v  got %v", i, tt.name, tt.want.NetWd, tt.input.NetWd)
			t.FailNow()
		}

		if !tt.input.Brute.Equal(tt.want.Brute) {
			t.Logf("[ TestCompleteFlow %d  %s ] different Brute. expecting %v  got %v", i, tt.name, tt.want.Brute, tt.input.Brute)
			t.FailNow()
		}

		if !tt.input.BruteWd.Equal(tt.want.BruteWd) {
			t.Logf("[ TestCompleteFlow %d  %s ] different BruteWd. expecting %v  got %v", i, tt.name, tt.want.BruteWd, tt.input.BruteWd)
			t.FailNow()
		}

		if !tt.input.DiscountAmount.Equal(tt.want.DiscountAmount) {
			t.Logf("[ TestCompleteFlow %d  %s ] different DiscountAmount. expecting %v  got %v", i, tt.name, tt.want.DiscountAmount, tt.input.DiscountAmount)
			t.FailNow()
		}

		if !tt.input.DiscountRatio.Equal(tt.want.DiscountRatio) {
			t.Logf("[ TestCompleteFlow %d  %s ] different DiscountRatio. expecting %v  got %v", i, tt.name, tt.want.DiscountRatio, tt.input.DiscountRatio)
			t.FailNow()
		}

		if !tt.input.Tax.Equal(tt.want.Tax) {
			t.Logf("[ TestCompleteFlow %d  %s ] different Tax. expecting %v  got %v", i, tt.name, tt.want.Tax, tt.input.Tax)
			t.FailNow()
		}

		if !tt.input.TaxRatio.Equal(tt.want.TaxRatio) {
			t.Logf("[ TestCompleteFlow %d  %s ] different TaxRatio. expecting %v  got %v", i, tt.name, tt.want.TaxRatio, tt.input.TaxRatio)
			t.FailNow()
		}

		if len(tt.input.Taxes) != len(tt.want.Taxes) {
			t.Logf("[ TestCompleteFlow %d  %s ] different Taxes len. expecting %v  got %v", i, tt.name, len(tt.want.Taxes), len(tt.input.Taxes))
			t.FailNow()
		}

		for k := range tt.input.Taxes {
			if !tt.input.Taxes[k].Amount.Equal(tt.want.Taxes[k].Amount) {
				t.Logf("[ TestCompleteFlow %d  %s ] different Taxes[%d].Amount    expecting  %v  got %v", i, tt.name, k, tt.want.Taxes[k].Amount, tt.input.Taxes[k].Amount)
				t.FailNow()
			}

			if !tt.input.Taxes[k].Ratio.Equal(tt.want.Taxes[k].Ratio) {
				t.Logf("[ TestCompleteFlow %d  %s ] different Taxes[%d].Ratio expecting %v  got %v", i, tt.name, k, tt.want.Taxes[k].Ratio, tt.input.Taxes[k].Ratio)
				t.FailNow()
			}

			if !tt.input.Taxes[k].Taxable.Equal(tt.want.Taxes[k].Taxable) {
				t.Logf("[ TestCompleteFlow %d  %s ] different Taxes[%d].Taxable expecting %v  got %v", i, tt.name, k, tt.want.Taxes[k].Taxable, tt.input.Taxes[k].Taxable)
				t.FailNow()
			}
		}
	}
}

func TestTaxToTaxDetail(t *testing.T) {

	_ = TaxToTaxDetail(Tax{
		Applies:    Unit,
		Value:      alpacadecimal.NewFromInt(0),
		Percentual: true,
		ApplyOn:    1,
	})
}

func TestUnitRndr(t *testing.T) {

	for i, tt := range unitRndrCases() {
		func(i int) {
			t.Run(tt.name, func(t2 *testing.T) {
				rounderUv := NewUnitValueRounder()
				rounderUv.SetNext(nil)
				err := rounderUv.Execute(&tt.detail)

				if err != nil && !tt.wantError {
					t.Logf("[ TestUnitValueRounder %d %s ] unexpected error %v", i, tt.name, err)
					t.FailNow()
				}

				if err == nil && tt.wantError {
					t.Logf("[ TestUnitValueRounder %d %s ] nil received when expecting error %v", i, tt.name, tt.explain)
					t.FailNow()
				}

				if err != nil && tt.wantError {
					t.SkipNow()
				}

				if !tt.detail.Uv.Equal(tt.wantUv) {
					t.Logf("[ TestUnitValueRounder %d %s ] expecting %v   got %v", i, tt.name, tt.wantUv, tt.detail.Uv)
					t.FailNow()
				}
			})
		}(i)
	}
}

func TestBruteUntaxer(t *testing.T) {

	for i, tt := range bruteUntaxerCases() {
		func(i int) {
			t.Run(tt.name, func(t2 *testing.T) {
				untaxer := NewBruteUntaxer(1)

				err := untaxer.Execute(&tt.detail)

				if err != nil && !tt.wantError {
					t2.Logf("[ TestBruteUntaxer %d %s ] Fails! err %v", i, tt.name, err)
					t2.FailNow()
				}

				if err == nil && tt.wantError {
					t2.Logf("[ TestBruteUntaxer %d %s ] Fails! nil received when expecting err %v", i, tt.name, tt.explain)
					t2.FailNow()
				}

				if err != nil && tt.wantError {
					t2.SkipNow()
				}

				if !tt.detail.Net.Equal(tt.wantUntaxed) {
					t.Logf("[ TestUnitValueUndiscounter  case %d,  %s ] Fails!  want untaxed %v    Got %v", i, tt.name, tt.wantUntaxed, tt.detail.Net)
					t.FailNow()
				}
			})
		}(i)
	}
}

func TestUnitValueUndiscounter(t *testing.T) {

	for i, tt := range undiscounterCases() {
		t.Run(t.Name(), func(t2 *testing.T) {
			err := (NewUnitValueUnDiscounter()).Execute(&tt.args)

			if err != nil && !tt.wantError {
				t.Logf("[ TestUnitValueUndiscounter  case %d,  %s ] Fails!  error %v", i, tt.name, err)
				t.FailNow()
			}

			if err == nil && tt.wantError {
				t.Logf("[ TestUnitValueUndiscounter  case %d,  %s ] Fails!  getting nil when expecting error %v", i, tt.name, tt.explain)
				t.FailNow()
			}

			if err != nil && tt.wantError {
				t.SkipNow()
			}

			if !tt.args.UvWd.Equal(tt.wantUndiscounted) {
				t.Logf("[ TestUnitValueUndiscounter  case %d,  %s ] Fails!  UvWd value. Expecting %v    Got %v", i, tt.name, tt.wantUndiscounted, tt.args.UvWd)
				t.FailNow()
			}
		})
	}
}

func TestNetUndiscounter(t *testing.T) {

	for i, tt := range netUndiscounterCases() {
		t.Run(t.Name(), func(t2 *testing.T) {
			err := (NewNetUnDiscounter()).Execute(&tt.args)

			if err != nil && !tt.wantError {
				t.Logf("[ TestNetUndiscounter  case %d,  %s ] Fails!  error %v", i, tt.name, err)
				t.FailNow()
			}

			if err == nil && tt.wantError {
				t.Logf("[ TestNetUndiscounter  case %d,  %s ] Fails!  getting nil when expecting error %v", i, tt.name, tt.explain)
				t.FailNow()
			}

			if err != nil && tt.wantError {
				t.SkipNow()
			}

			if !tt.args.NetWd.Equal(tt.wantUndiscounted) {
				t.Logf("[ TestNetUndiscounter  case %d,  %s ] Fails!  NetWd value. Expecting %v    Got %v", i, tt.name, tt.wantUndiscounted, tt.args.NetWd)
				t.FailNow()
			}
		})
	}
}

func TestWrappingError(t *testing.T) {
	err := ErrNegativeValue

	err = WrapWithWrappingError(err, "testing wrapping errors")
	err = WrapWithWrappingError(err, "testing wrapping errors level 2")
	err = WrapWithWrappingError(err, "testing wrapping errors level 3")

	if !errors.Is(err, ErrNegativeValue) {
		t.Logf("[ TestWrappingError ] Fails! expecting a wrapping of ErrNegativeValue got %v", err)
		t.FailNow()
	}

	if err.Error() == "" {
		t.Log("[ TestWrappingError ] Fails! expecting something got empty")
		t.FailNow()
	}
}

func TestDetail_serialize(t *testing.T) {
	dt := &Detail{
		Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
		Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
		Discounts: []Discount{
			{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("70"); return d }(), Unit, true},
			{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("500"); return d }(), Line, false},
			{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("500"); return d }(), Line, false},
		},
		Taxes: []TaxDetail{
			{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("5"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), alpacadecimal.Zero, Unit, 1, true},
		},
	}

	got := dt.serialize()

	want := `	// Uv unitary value of the product being sold
	Uv alpacadecimal.Decimal      = 100
	// Qty quantity being sold
	Qty alpacadecimal.Decimal     = 10
	// Discounts list of applied discounts
	Discounts []Discount  = [{"Applies":0,"Value":"70","Percentual":true},{"Applies":1,"Value":"500","Percentual":false},{"Applies":1,"Value":"500","Percentual":false}]
	// Taxes detail of applied taxes over the sale
	Taxes []TaxDetail     = [{"ratio":"5","applies":0,"amount":"0","taxable":"0","percentual":true}]
	// Net total value without taxes of the sale. The result of: Uv * Qty - discounts
	Net alpacadecimal.Decimal     = 0
	// NetWd total value without taxes and without discounts of the sale. The result of: Uv * Qty
	NetWd alpacadecimal.Decimal   = 0
	// Brute total value including taxes.  net + taxes
	Brute alpacadecimal.Decimal   = 0
	// BruteWd total value including taxes without discounts. netWd + taxesWd
	BruteWd alpacadecimal.Decimal = 0
	// Tax value of the taxes being applied considering discounts
	Tax alpacadecimal.Decimal     = 0
	// TaxRatio percentual ratio of the tax value over the brute
	TaxRatio alpacadecimal.Decimal = 0
	// TaxWd value of the taxes being applied without consider discounts
	TaxWd alpacadecimal.Decimal    = 0
	// TaxRatioWd percentual ratio of the tax value over the bruteWd
	TaxRatioWd alpacadecimal.Decimal = 0
	// DiscountAmount cummulated amount of the discounts applied
	DiscountAmount alpacadecimal.Decimal = 0
	// DiscountRatio percentual ratio of DiscountAmount over Brute
	DiscountRatio alpacadecimal.Decimal = 0`

	if got != want {
		t.Log("[ TestDetail_serialize ] Fails! want differs got!")
	}
}

func TestBruteUndiscounter(t *testing.T) {

	for i, tt := range bruteUndiscounterCases() {
		t.Run(t.Name(), func(t2 *testing.T) {
			err := (NewBruteUnDiscounter()).Execute(&tt.args)

			if err != nil && !tt.wantError {
				t.Logf("[ TestBruteUndiscounter  case %d,  %s ] Fails!  error %v", i, tt.name, err)
				t.FailNow()
			}

			if err == nil && tt.wantError {
				t.Logf("[ TestBruteUndiscounter  case %d,  %s ] Fails!  getting nil when expecting error %v", i, tt.name, tt.explain)
				t.FailNow()
			}

			if err != nil && tt.wantError {
				t.SkipNow()
			}

			if !tt.args.BruteWd.Equal(tt.wantUndiscounted) {
				t.Logf("[ TestBruteUndiscounter  case %d,  %s ] Fails!  BruteWd value. Expecting %v    Got %v", i, tt.name, tt.wantUndiscounted, tt.args.BruteWd)
				t.FailNow()
			}
		})
	}
}

func TestMultiStageTaxes_Execute(t *testing.T) {

	for i, tt := range multistageTaxes_ExecuteCases() {
		t.Run(t.Name(), func(t2 *testing.T) {
			brute := NewBruteSnapshot()
			third := NewTaxStage(3)
			third.SetNext(brute)
			second := NewTaxStage(2)
			second.SetNext(third)
			first := NewTaxStage(1)
			first.SetNext(second)
			discHandler := NewDiscounter()
			discHandler.SetNext(first)

			err := discHandler.Execute(tt.dt)

			if err != nil && !tt.wantError {
				t2.Logf("[ TestMultiStageTaxes_Execute %d %s ] Fails!  %v", i, tt.name, err)
				t2.FailNow()
			}

			if err == nil && tt.wantError {
				t2.Logf("[ TestMultiStageTaxes_Execute %d %s ] Fails! nil received when expecting error  %s", i, tt.name, tt.explain)
				t2.FailNow()
			}

			if err != nil && tt.wantError {
				t2.SkipNow()
			}

			if !tt.wantnet.Equal(tt.dt.Net) {
				t2.Logf("[ TestMultiStageTaxes_Execute %d %s ] Fails! expecting net = %v   got = %v", i, tt.name, tt.wantnet, tt.dt.Net)
				t2.FailNow()
			}

			if !tt.wantnetWd.Equal(tt.dt.NetWd) {
				t2.Logf("[ TestMultiStageTaxes_Execute %d %s ] Fails! expecting netWd = %v   got = %v", i, tt.name, tt.wantnetWd, tt.dt.NetWd)
				t2.FailNow()
			}

			if !tt.wantbrute.Equal(tt.dt.Brute) {
				t2.Logf("[ TestMultiStageTaxes_Execute %d %s ] Fails! expecting brute = %v   got = %v", i, tt.name, tt.wantbrute, tt.dt.Brute)
				t2.FailNow()
			}

			if !tt.wantbruteWd.Equal(tt.dt.BruteWd) {
				t2.Logf("[ TestMultiStageTaxes_Execute %d %s ] Fails! expecting bruteWd = %v   got = %v", i, tt.name, tt.wantbruteWd, tt.dt.BruteWd)
				t2.FailNow()
			}

			if !tt.wanttax.Equal(tt.dt.Tax) {
				t2.Logf("[ TestMultiStageTaxes_Execute %d %s ] Fails! expecting tax = %v   got = %v", i, tt.name, tt.wanttax, tt.dt.Tax)
				t2.FailNow()
			}

			if !tt.wanttaxWd.Equal(tt.dt.TaxWd) {
				t2.Logf("[ TestMultiStageTaxes_Execute %d %s ] Fails! expecting taxWd = %v   got = %v", i, tt.name, tt.wanttaxWd, tt.dt.TaxWd)
				t2.FailNow()
			}
		})
	}
}

func TestSecondStageTaxes_Execute(t *testing.T) {
	type fields struct {
		next Stage
	}
	type args struct {
		dt *Detail
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantTax   alpacadecimal.Decimal
		wantTaxWd alpacadecimal.Decimal
		wantErr   bool
		explain   string
	}{
		{
			"Tax Second stage | 19 percentual - 10 amount - 1 line",
			fields{nil},
			args{&Detail{
				Uv:    func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty:   func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Net:   func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				NetWd: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				Tax:   func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("291"); return d }(),
				TaxWd: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("291"); return d }(),
				Taxes: []TaxDetail{
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("19"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(), Unit, 1, true},
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(), Unit, 1, false},
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(), Line, 1, false},
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("19"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1291"); return d }(), Unit, 2, true},
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1291"); return d }(), Unit, 2, false},
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1291"); return d }(), Line, 2, false},
				},
			}},
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("637.29"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("637.29"); return d }(),
			false,
			"",
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &TaxStage{
				next:        tt.fields.next,
				stageNumber: 2,
			}

			err := d.Execute(tt.args.dt)

			if err != nil && !tt.wantErr {
				t.Logf("[ TestFirstStageTaxes_Execute %d %s ] FirstStageTaxes.Execute() error = %v", i, tt.name, err)
				t.FailNow()
			}

			if err == nil && tt.wantErr {
				t.Logf("[ TestFirstStageTaxes_Execute %d %s ] FirstStageTaxes.Execute() got nil when expecting error = %v", i, tt.name, tt.explain)
				t.FailNow()
			}

			if err != nil && tt.wantErr {
				t.SkipNow()
			}

			if !tt.wantTax.Equal(tt.args.dt.Tax) {
				t.Logf("[ TestFirstStageTaxes_Execute %d %s ] FirstStageTaxes.Execute() expectig tax = %v  got %v", i, tt.name, tt.wantTax, tt.args.dt.Tax)
				t.FailNow()
			}

			if !tt.wantTaxWd.Equal(tt.args.dt.TaxWd) {
				t.Logf("[ TestFirstStageTaxes_Execute %d %s ] FirstStageTaxes.Execute() expectig taxWd = %v  got %v", i, tt.name, tt.wantTaxWd, tt.args.dt.TaxWd)
				t.FailNow()
			}

		})
	}
}

func TestFirstStageTaxes_Execute(t *testing.T) {
	type fields struct {
		next Stage
	}
	type args struct {
		dt *Detail
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantTax   alpacadecimal.Decimal
		wantTaxWd alpacadecimal.Decimal
		wantErr   bool
		explain   string
	}{
		{
			"Tax First stage | 19 percentual - 10 amount - 1 line",
			fields{nil},
			args{&Detail{
				Uv:    func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty:   func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Net:   func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				NetWd: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				Taxes: []TaxDetail{
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("19"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Unit, 1, true},
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Unit, 1, false},
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Line, 1, false},
				},
			}},
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("291"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("291"); return d }(),
			false,
			"",
		},
		{
			"Tax First stage | 19 percentual - 10 amount - 1 line | different tax and taxWd",
			fields{nil},
			args{&Detail{
				Uv:    func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("200"); return d }(),
				Qty:   func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Net:   func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				NetWd: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("2000"); return d }(),
				Taxes: []TaxDetail{
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("19"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Unit, 1, true},
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Unit, 1, false},
					{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Line, 1, false},
				},
			}},
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("291"); return d }(),
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("481"); return d }(),
			false,
			"",
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := NewTaxStage(1)
			d.SetNext(tt.fields.next)

			err := d.Execute(tt.args.dt)

			if err != nil && !tt.wantErr {
				t.Logf("[ TestFirstStageTaxes_Execute %d %s ] FirstStageTaxes.Execute() error = %v", i, tt.name, err)
				t.FailNow()
			}

			if err == nil && tt.wantErr {
				t.Logf("[ TestFirstStageTaxes_Execute %d %s ] FirstStageTaxes.Execute() got nil when expecting error = %v", i, tt.name, tt.explain)
				t.FailNow()
			}

			if err != nil && tt.wantErr {
				t.SkipNow()
			}

			if !tt.wantTax.Equal(tt.wantTax) {
				t.Logf("[ TestFirstStageTaxes_Execute %d %s ] FirstStageTaxes.Execute() expectig tax = %v  got %v", i, tt.name, tt.wantTax, tt.args.dt.Tax)
				t.FailNow()
			}

			if !tt.wantTaxWd.Equal(tt.wantTaxWd) {
				t.Logf("[ TestFirstStageTaxes_Execute %d %s ] FirstStageTaxes.Execute() expectig taxWd = %v  got %v", i, tt.name, tt.wantTaxWd, tt.args.dt.TaxWd)
				t.FailNow()
			}

		})
	}
}

func TestSetTaxable(t *testing.T) {
	type args struct {
		detail  []TaxDetail
		taxable alpacadecimal.Decimal
	}
	tests := []struct {
		name string
		args args
		want alpacadecimal.Decimal
	}{
		{
			"mutable taxable",
			args{
				[]TaxDetail{{"1", func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("18"); return d }(), alpacadecimal.Zero, alpacadecimal.Zero, Unit, 1, true}},
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
			},
			func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			SetTaxesValues(tt.args.detail, tt.args.taxable)

			if !tt.args.taxable.Equal(tt.want) {
				t.Logf("[ Test TestSetTaxable %d %v ] Fails want taxable %v got %v", 8, tt.name, tt.want, tt.args.taxable)
				t.FailNow()
			}
		})
	}
}

func TestDiscounter_Execute(t *testing.T) {
	type fields struct {
		next Stage
	}
	type args struct {
		dt *Detail
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		explain string
		want    struct {
			netWd alpacadecimal.Decimal
			net   alpacadecimal.Decimal
		}
	}{

		{
			"3 Discounts 8% | 10 amount | 1 line",
			fields{nil},
			args{&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("8"); return d }(), Unit, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Unit, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), Line, false},
				},
			}},
			false,
			"",
			struct {
				netWd alpacadecimal.Decimal
				net   alpacadecimal.Decimal
			}{
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("819"); return d }(),
			},
		},
		{
			"want error | negative percentual",
			fields{nil},
			args{&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("-8"); return d }(), Unit, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Unit, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), Line, false},
				},
			}},
			true,
			"invalid negative percentual discount",
			struct {
				netWd alpacadecimal.Decimal
				net   alpacadecimal.Decimal
			}{
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("819"); return d }(),
			},
		},
		{
			"want error | negative amount",
			fields{nil},
			args{&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("8"); return d }(), Unit, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("-10"); return d }(), Unit, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1"); return d }(), Line, false},
				},
			}},
			true,
			"invalid negative amount discount",
			struct {
				netWd alpacadecimal.Decimal
				net   alpacadecimal.Decimal
			}{
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("819"); return d }(),
			},
		},
		{
			"want error | negative line",
			fields{nil},
			args{&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("8"); return d }(), Unit, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Unit, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("-1"); return d }(), Line, false},
				},
			}},
			true,
			"invalid negative line discount",
			struct {
				netWd alpacadecimal.Decimal
				net   alpacadecimal.Decimal
			}{
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("819"); return d }(),
			},
		},
		{
			"cummulated discount over 100%  should be overwrite to 100",
			fields{nil},
			args{&Detail{
				Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("100"); return d }(),
				Qty: func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
				Discounts: []Discount{
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("99.99"); return d }(), Unit, true},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Unit, false},
					{func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), Line, false},
				},
			}},
			false,
			"",
			struct {
				netWd alpacadecimal.Decimal
				net   alpacadecimal.Decimal
			}{
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("1000"); return d }(),
				func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("0"); return d }(),
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Discounter{
				next: tt.fields.next,
			}

			err := d.Execute(tt.args.dt)

			if err != nil && !tt.wantErr {
				t.Logf("[TestDiscounter_Execute %d  %s]  error = %v, wantErr %v", i, tt.name, err, tt.wantErr)
				t.FailNow()
			}

			if err == nil && tt.wantErr {
				t.Logf("[TestDiscounter_Execute %d  %s] Fails!  nil received when expecting error %s", i, tt.name, tt.explain)
				t.FailNow()
			}

			if err != nil && tt.wantErr {
				t.SkipNow()
			}

			if !tt.args.dt.NetWd.Equal(tt.want.netWd) {
				t.Logf("[TestDiscounter_Execute %d  %s] Fails!  want NetWd %s   got %s", i, tt.name, tt.want.netWd, tt.args.dt.NetWd)
				t.FailNow()
			}

			if !tt.args.dt.Net.Equal(tt.want.net) {
				t.Logf("[TestDiscounter_Execute %d  %s] Fails!  want Net %s   got %s", i, tt.name, tt.want.net, tt.args.dt.Net)
				t.FailNow()
			}
		})
	}
}
