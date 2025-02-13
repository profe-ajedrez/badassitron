package badassitron

import (
	"testing"

	"github.com/alpacahq/alpacadecimal"
)

func BenchmarkCompleteFlow(b *testing.B) {
	for i, tt := range completeFlowTestCases() {
		func(i int) {
			b.Run(tt.name, func(b2 *testing.B) {
				_ = tt.process(&tt.input)
			})
		}(i)
	}
}

func BenchmarkUnitRndr(b *testing.B) {
	for i, tt := range unitRndrCases() {
		func(i int) {
			b.Run(tt.name, func(b2 *testing.B) {
				rounderUv := NewUnitValueRounder()
				rounderUv.SetNext(nil)
				_ = rounderUv.Execute(&tt.detail)
			})
		}(i)
	}
}

func BenchmarkBruteUntaxer(b *testing.B) {
	for i, tt := range bruteUntaxerCases() {
		func(i int) {
			b.Run(tt.name, func(b2 *testing.B) {
				untaxer := NewBruteUntaxer(1)
				_ = untaxer.Execute(&tt.detail)
			})
		}(i)
	}
}

func BenchmarkValueUndiscounter(b *testing.B) {
	for i, tt := range undiscounterCases() {
		func(i int) {
			b.Run(tt.name, func(b2 *testing.B) {
				_ = (NewUnitValueUnDiscounter()).Execute(&tt.args)
			})
		}(i)
	}
}

func BenchmarkNetUndiscounter(b *testing.B) {
	for i, tt := range netUndiscounterCases() {
		func(i int) {
			b.Run(tt.name, func(b2 *testing.B) {
				_ = (NewNetUnDiscounter()).Execute(&tt.args)
			})
		}(i)
	}
}

func BenchmarkDetail_serialize(b *testing.B) {
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

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		_ = dt.serialize()
	}
}

func BenchmarkBruteUndiscounter(b *testing.B) {
	for i, tt := range bruteUndiscounterCases() {
		func(i int) {
			b.Run(tt.name, func(b2 *testing.B) {
				_ = (NewBruteUnDiscounter()).Execute(&tt.args)
			})
		}(i)
	}
}

func BenchmarkMultistageTaxes_execute(b *testing.B) {
	for i, tt := range multistageTaxes_ExecuteCases() {
		func(i int) {
			b.Run(tt.name, func(b2 *testing.B) {
				brute := NewBruteSnapshot()
				third := NewTaxStage(3)
				third.SetNext(brute)
				second := NewTaxStage(2)
				second.SetNext(third)
				first := NewTaxStage(1)
				first.SetNext(second)
				discHandler := NewDiscounter()
				discHandler.SetNext(first)

				_ = discHandler.Execute(tt.dt)
			})
		}(i)
	}
}
