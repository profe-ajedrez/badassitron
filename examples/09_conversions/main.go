// This program moves decimal values between packages without silent rounding.
package main

import (
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec256"
	"github.com/profe-ajedrez/badassitron/dec512"
	"github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
	source := dec128.FromString("123.4500")
	wide := dec256.FromDec128(source)
	wider := dec512.FromDec256(wide)
	adaptive := optimus.FromDec512(wider)
	if adaptive.IsNaN() {
		log.Fatal(adaptive.ErrorDetails())
	}
	back, err := adaptive.Dec128Exact()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Round trip:", back)
	fmt.Println("Same value:", back.Equal(source))

	// Removing trailing zeros can allow an exact conversion to a smaller scale.
	reducible := optimus.FromString("1.2000000000000000000000000")
	narrow, err := reducible.Dec128Exact()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Trailing zeros removed:", narrow)

	// This nonzero digit needs 20 fractional places. Dec128 permits at most 19.
	tiny := optimus.FromString("0.00000000000000000001")
	_, err = tiny.Dec128Exact()
	fmt.Println("Lossy conversion rejected:", err != nil)

	// If the application permits loss, round explicitly before converting.
	rounded, inexact := tiny.RescaleRoundWithStatus(19, optimus.RoundingHalfEven)
	if rounded.IsNaN() {
		log.Fatal(rounded.ErrorDetails())
	}
	narrow, err = rounded.Dec128Exact()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Rounded conversion:", narrow)
	fmt.Println("Rounding lost digits:", inexact)
}
