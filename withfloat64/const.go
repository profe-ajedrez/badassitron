package withfloat64

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

	// FromNet selects input from net total.
	FromNet = 0
	// FromGross selects input from gross total.
	FromGross = 1

	// FromUV selects input from unit value.
	FromUV = 0
)

// Zero returns the floating point value zero.
func Zero() float64 { return 0 }

// One returns the floating point value one.
func One() float64 { return 1 }

// Two returns the floating point value two.
func Two() float64 { return 2 }

// Ten returns the floating point value ten.
func Ten() float64 { return 10 }

// Hundred returns the floating point value hundred.
func Hundred() float64 { return 100 }
