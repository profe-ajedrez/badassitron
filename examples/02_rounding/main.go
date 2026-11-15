// This program compares rounding rules and shows how to detect lost digits.
package main

import (
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
	tie := optimus.FromString("2.5")
	fmt.Println("Half away from zero:", tie.Round(0))
	fmt.Println("Half even:", tie.RoundBank(0))

	negative := optimus.FromString("-1.239")
	fmt.Println("Truncate:", negative.Trunc(2))
	fmt.Println("Floor:", negative.RoundDown(2))
	fmt.Println("Ceiling:", negative.RoundUp(2))

	// A true status means that nonzero digits were discarded. It is not an error.
	third, inexact := optimus.FromInt(1).DivRoundWithStatus(
		optimus.FromInt(3), 2, optimus.RoundingHalfEven,
	)
	if third.IsNaN() {
		log.Fatal(third.ErrorDetails())
	}
	fmt.Println("One third:", third.StringFixed(), "inexact:", inexact)

	quarter, inexact := optimus.FromInt(1).DivRoundWithStatus(
		optimus.FromInt(4), 2, optimus.RoundingHalfEven,
	)
	if quarter.IsNaN() {
		log.Fatal(quarter.ErrorDetails())
	}
	fmt.Println("One quarter:", quarter.StringFixed(), "inexact:", inexact)

	// Round removes digits. RescaleRound can also add zeros for display.
	fixed := optimus.FromInt(12).RescaleRound(2, optimus.RoundingHalfEven)
	if fixed.IsNaN() {
		log.Fatal(fixed.ErrorDetails())
	}
	fmt.Println("Two decimal places:", fixed.StringFixed())

	root, inexact := optimus.FromInt(2).SqrtRoundWithStatus(6, optimus.RoundingHalfEven)
	if root.IsNaN() {
		log.Fatal(root.ErrorDetails())
	}
	fmt.Println("Square root of two:", root.StringFixed(), "inexact:", inexact)
}
