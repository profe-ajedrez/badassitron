package dec128_test

import (
	"fmt"

	"github.com/profe-ajedrez/badassitron/dec128"
)

func ExampleDec128_RescaleRound() {
	value := dec128.FromString("12.5")
	fixed := value.RescaleRound(2, dec128.RoundingHalfEven)
	fmt.Println(fixed.StringFixed())
	// Output: 12.50
}

func ExampleDec128_DivRoundWithStatus() {
	value, inexact := dec128.FromInt(1).DivRoundWithStatus(
		dec128.FromInt(3), 2, dec128.RoundingHalfEven,
	)
	fmt.Println(value.String(), inexact)
	// Output: 0.33 true
}

func ExampleScanInto() {
	var amount dec128.Dec128
	if err := dec128.ScanInto(&amount, "42.500"); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amount.StringFixed())
	// Output: 42.500
}
