package withdec128

import "github.com/profe-ajedrez/badassitron/dec128"

// Stage identifies when a tax is applied in the calculation workflow.
type Stage int8

// Type identifies the method used to calculate a tax.
type Type int8

// Mode identifies a calculation input mode.
type Mode int8

const (
	// Natural selects the natural tax stage.
	Natural Stage = 0
	// Overtax selects the overtax stage.
	Overtax Stage = 1
	// Bypass selects the bypass tax stage.
	Bypass Stage = 2

	// Percentual selects percentage based tax calculation.
	Percentual Type = 0
	// Amount selects a fixed tax amount.
	Amount Type = 1
	// AmountLine selects a fixed tax amount for the line.
	AmountLine Type = 2

	// FromUV selects input from unit value.
	FromUV = 0
	// FromGross selects input from gross total.
	FromGross = 1
)

// Zero returns the decimal value zero.
func Zero() dec128.Dec128 { return dec128.Decimal0.Copy() }

// One returns the decimal value one.
func One() dec128.Dec128 { return dec128.Decimal1.Copy() }

// Two returns the decimal value two.
func Two() dec128.Dec128 { return dec128.Decimal2.Copy() }

// Ten returns the decimal value ten.
func Ten() dec128.Dec128 { return dec128.Decimal10.Copy() }

// Hundred returns the decimal value hundred.
func Hundred() dec128.Dec128 { return dec128.Decimal100.Copy() }
