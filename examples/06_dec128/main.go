// This program uses fixed 128 bit decimals with explicit calculation scales.
package main

import (
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/dec128"
)

func main() {
	price, err := dec128.NewFromString("98.00")
	if err != nil {
		log.Fatal(err)
	}
	mode := dec128.RoundingHalfEven

	// Keep 19 fractional places in each intermediate result that fits.
	unit := price.DivRound(dec128.FromString("1.18"), 19, mode)
	subtotal := unit.MulRound(dec128.FromInt(2000), 19, mode)
	net := subtotal.MulRound(dec128.FromString("0.90"), 19, mode)
	published, inexact := net.RescaleRoundWithStatus(2, mode)
	if published.IsNaN() {
		log.Fatal(published.ErrorDetails())
	}
	fmt.Println("Published:", published.StringFixed())
	fmt.Println("Final rounding lost digits:", inexact)

	root := dec128.FromString("2.25").SqrtRound(2, mode)
	if root.IsNaN() {
		log.Fatal(root.ErrorDetails())
	}
	fmt.Println("Square root:", root.StringFixed())

	// QuoRem uses an integer quotient truncated toward zero.
	quotient, remainder := dec128.FromString("-10.5").QuoRem(dec128.FromInt(3))
	if quotient.IsNaN() || remainder.IsNaN() {
		log.Fatal("quotient or remainder failed")
	}
	fmt.Println("Quotient:", quotient)
	fmt.Println("Remainder:", remainder)
}
