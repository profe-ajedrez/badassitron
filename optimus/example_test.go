package optimus_test

import (
	"fmt"

	"github.com/profe-ajedrez/badassitron/optimus"
)

func ExampleOptimus_RoundBank() {
	value := optimus.FromString("2.5")
	fmt.Println(value.Round(0))
	fmt.Println(value.RoundBank(0))
	// Output:
	// 3
	// 2
}

func ExampleOptimus_DivRoundWithStatus() {
	one := optimus.FromInt(1)
	three := optimus.FromInt(3)
	value, inexact := one.DivRoundWithStatus(three, 2, optimus.RoundingHalfEven)
	fmt.Println(value, inexact)
	// Output: 0.33 true
}

func ExampleSQLDecimal_Scan() {
	var amount optimus.SQLDecimal
	if err := amount.Scan("123.4500"); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amount.Valid, amount.Decimal.StringFixed())
	// Output: true 123.4500
}
