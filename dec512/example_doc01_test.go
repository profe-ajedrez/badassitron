package dec512_test

import (
	"fmt"
	"github.com/profe-ajedrez/badassitron/dec512"
)

// ExampleNewFromString shows checked decimal parsing without floating point.
func ExampleNewFromString() {
	v, err := dec512.NewFromString("12.3400")
	if err != nil {
		fmt.Println("parse error:", err)
		return
	}
	fmt.Println(v.String(), v.StringFixed())
	_, err = dec512.NewFromString("broken")
	fmt.Println("invalid:", err != nil)
	// Output:
	// 12.34 12.3400
	// invalid: true
}

// ExampleDec512_RoundBank illustrates half-even rounding at a chosen scale.
func ExampleDec512_RoundBank() {
	value := dec512.FromString("2.5")
	fmt.Println(value.Round(0), value.RoundBank(0))
	// Output: 3 2
}
