// This program uses decimal values that need more than 128 coefficient bits.
package main

import (
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/dec256"
	"github.com/profe-ajedrez/badassitron/dec256/uint256"
)

func main() {
	// This 41 digit integer is too large for a 128 bit coefficient.
	large, err := dec256.NewFromString("10000000000000000000000000000000000000000")
	if err != nil {
		log.Fatal(err)
	}
	total := large.Add(dec256.FromInt(25))
	if total.IsNaN() {
		log.Fatal(total.ErrorDetails())
	}
	fmt.Println("Large total:", total)

	// The coefficient and the scale have separate limits.
	small := dec256.FromString("0.123456789012345678901234567890")
	if small.IsNaN() {
		log.Fatal(small.ErrorDetails())
	}
	fmt.Println("Stored decimal:", small.StringFixed())
	fmt.Println("Stored scale:", small.Precision())

	part, inexact := dec256.FromInt(2).DivRoundWithStatus(
		dec256.FromInt(3), 6, dec256.RoundingHalfEven,
	)
	if part.IsNaN() {
		log.Fatal(part.ErrorDetails())
	}
	fmt.Println("Two thirds:", part.StringFixed(), "inexact:", inexact)

	// Fixed width arithmetic cannot grow to the next coefficient tier.
	maximum := dec256.New(uint256.Max, 0, false)
	overflow := maximum.Add(dec256.FromInt(1))
	fmt.Println("Overflow detected:", overflow.IsNaN())
}
