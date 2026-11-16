package optimus_test

import (
	"fmt"
	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/optimus"
)

// ExampleNewFromString shows error-returning parsing for external input.
func ExampleNewFromString() {
	amount, err := optimus.NewFromString("123.4500")
	if err != nil {
		fmt.Println("parse error:", err)
		return
	}
	fmt.Println(amount.String())
	fmt.Println(amount.StringFixed())
	_, err = optimus.NewFromString("not-a-number")
	fmt.Println("invalid input:", err != nil)
	// Output:
	// 123.45
	// 123.4500
	// invalid input: true
}

// ExampleOptimus_Mul demonstrates exact decimal multiplication when the result
// fits the coefficient and scale limits.
func ExampleOptimus_Mul() {
	a, err := optimus.NewFromString("19.95")
	if err != nil {
		fmt.Println(err)
		return
	}
	b, err := optimus.NewFromString("3")
	if err != nil {
		fmt.Println(err)
		return
	}
	result := a.Mul(b)
	if result.IsNaN() {
		fmt.Println("arithmetic failed")
		return
	}
	fmt.Println(result)
	// Output: 59.85
}

// ExampleOptimus_RescaleRoundWithStatus checks whether a quantization has
// discarded nonzero digits. The boolean is not an error.
func ExampleOptimus_RescaleRoundWithStatus() {
	value := optimus.FromString("1.235")
	amount, changed := value.RescaleRoundWithStatus(2, optimus.RoundingHalfEven)
	fmt.Println(amount, changed)
	// Output: 1.24 true
}

// ExampleOptimus_Dec128Exact converts only if the exact value is representable
// at the smaller coefficient and scale limits.
func ExampleOptimus_Dec128Exact() {
	source := optimus.FromString("42.125")
	fixed, err := source.Dec128Exact()
	if err != nil {
		fmt.Println("conversion error:", err)
		return
	}
	fmt.Println(fixed.String(), optimus.FromDec128(fixed).Equal(source))
	tooManyPlaces := optimus.FromString("0.123456789012345678901")
	_, err = tooManyPlaces.Dec128Exact()
	fmt.Println("not exactly representable:", err != nil)
	// Output:
	// 42.125 true
	// not exactly representable: true
}

// ExampleOptimus_Round contrasts the two common tie-breaking policies.
func ExampleOptimus_Round() {
	pos := optimus.FromString("2.5")
	neg := optimus.FromString("-2.5")
	fmt.Println(pos.Round(0), pos.RoundBank(0))
	fmt.Println(neg.Round(0), neg.RoundBank(0))
	// Output:
	// 3 2
	// -3 -2
}

// ExampleSQLDecimal demonstrates SQL NULL distinct from numeric zero.
func ExampleSQLDecimal() {
	var value optimus.SQLDecimal
	if err := value.Scan(nil); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("null:", !value.Valid)
	if err := value.Scan("0.00"); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("valid:", value.Valid, "value:", value.Decimal.StringFixed())
	if err := value.Scan(float64(1.5)); err == nil {
		fmt.Println("unexpectedly accepted float")
		return
	}
	fmt.Println("float rejected")
	// Output:
	// null: true
	// valid: true value: 0.00
	// float rejected
}

// ExampleOptimus_DivRound demonstrates a caller-selected precision and mode.
func ExampleOptimus_DivRound() {
	result := optimus.FromInt(1).DivRound(optimus.FromInt(3), 2, dec128.RoundingHalfEven)
	if result.IsNaN() {
		fmt.Println("division failed")
		return
	}
	fmt.Println(result)
	// Output: 0.33
}
