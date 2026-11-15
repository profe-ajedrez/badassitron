// This program shows how to read, calculate, format, and validate decimals.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
	// Use text for decimal input. A float may already have lost decimal digits.
	text := "123.4500"
	// The library accepts empty text as zero. Reject it here if a value is required.
	if strings.TrimSpace(text) == "" {
		log.Fatal("amount is required")
	}
	amount, err := optimus.NewFromString(text)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Compact:", amount.String())
	fmt.Println("Stored scale:", amount.Precision())
	fmt.Println("Fixed:", amount.StringFixed())

	// Arithmetic returns a new value. The original amount stays the same.
	total := amount.Mul(optimus.FromInt(2)).Add(optimus.FromString("0.10"))
	if total.IsNaN() {
		log.Fatal(total.ErrorDetails())
	}
	fmt.Println("Total:", total)
	fmt.Println("Original:", amount)
	fmt.Println("Same value:", amount.Equal(optimus.FromString("123.45")))

	// The shorter parser returns NaN for invalid text instead of a Go error.
	invalid := optimus.FromString("not a number")
	fmt.Println("Invalid input:", invalid.IsNaN())
	_, err = optimus.NewFromString("not a number")
	fmt.Println("Parser returned an error:", err != nil)

	// Arithmetic failures also use NaN. Check them before publishing a result.
	invalid = amount.Div(optimus.FromInt(0))
	fmt.Println("Division by zero:", invalid.IsNaN())
	fmt.Println("Error details available:", invalid.ErrorDetails() != nil)

	// Storage grows automatically when a coefficient needs more bits.
	large := optimus.FromString("18446744073709551615").Add(optimus.FromInt(1))
	if large.IsNaN() {
		log.Fatal(large.ErrorDetails())
	}
	fmt.Println("Large integer:", large)
	fmt.Println("Coefficient bits:", large.Width())
}
