// This program demonstrates why line rounding and total rounding can differ.
package main

import (
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
	inputs := []string{"1.005", "2.005", "3.005"}
	exactTotal := optimus.FromInt(0)
	lineTotal := optimus.FromInt(0)

	for _, text := range inputs {
		line, err := optimus.NewFromString(text)
		if err != nil {
			log.Fatal(err)
		}
		rounded, inexact := line.RescaleRoundWithStatus(2, optimus.RoundingHalfEven)
		if rounded.IsNaN() {
			log.Fatal(rounded.ErrorDetails())
		}
		exactTotal = exactTotal.Add(line)
		lineTotal = lineTotal.Add(rounded)
		if exactTotal.IsNaN() || lineTotal.IsNaN() {
			log.Fatal("total exceeded decimal limits")
		}
		fmt.Println("Line:", text, "->", rounded.StringFixed(), "inexact:", inexact)
	}
	totalRounded := exactTotal.RescaleRound(2, optimus.RoundingHalfEven)
	if totalRounded.IsNaN() {
		log.Fatal(totalRounded.ErrorDetails())
	}
	fmt.Println("Round each line:", lineTotal.StringFixed())
	fmt.Println("Round the total:", totalRounded.StringFixed())
}
