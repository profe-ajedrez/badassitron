package dec128_test

import (
	"fmt"
	"github.com/profe-ajedrez/badassitron/dec128"
)

// ExampleNewFromString shows checked decimal parsing without floating point.
func ExampleNewFromString() {
	v, err := dec128.NewFromString("12.3400")
	if err != nil {
		fmt.Println("parse error:", err)
		return
	}
	fmt.Println(v.String(), v.StringFixed())
	_, err = dec128.NewFromString("broken")
	fmt.Println("invalid:", err != nil)
	// Output:
	// 12.34 12.3400
	// invalid: true
}

// ExampleDec128_RoundBank illustrates half-even rounding at a chosen scale.
func ExampleDec128_RoundBank() {
	value := dec128.FromString("2.5")
	fmt.Println(value.Round(0), value.RoundBank(0))
	// Output: 3 2
}
