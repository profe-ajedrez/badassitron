// This program uses explicit scales for independent concurrent calculations.
package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
	// Each worker reads the same values and writes to its own result slot.
	one := optimus.FromInt(1)
	seven := optimus.FromInt(7)
	scales := []uint8{2, 4, 6}
	results := make([]optimus.Optimus, len(scales))
	var workers sync.WaitGroup
	for i, scale := range scales {
		workers.Add(1)
		go func(index int, places uint8) {
			defer workers.Done()
			results[index] = one.DivRound(seven, places, optimus.RoundingHalfEven)
		}(i, scale)
	}
	workers.Wait()

	// Print after all writes finish so the output order is stable.
	for i, result := range results {
		if result.IsNaN() {
			log.Fatal(result.ErrorDetails())
		}
		fmt.Printf("Scale %d: %s\n", scales[i], result.StringFixed())
	}
}
