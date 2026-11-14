// This program uses a fixed 512 bit coefficient for very large decimals.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/profe-ajedrez/badassitron/dec512"
)

func main() {
	// A one followed by 80 zeros is larger than a 256 bit coefficient can hold.
	large, err := dec512.NewFromString("1" + strings.Repeat("0", 80))
	if err != nil {
		log.Fatal(err)
	}
	shifted := large.Add(dec512.FromInt(7))
	if shifted.IsNaN() {
		log.Fatal(shifted.ErrorDetails())
	}
	fmt.Println("Large total:", shifted)

	// A high scale can also represent a very small value exactly.
	tiny := dec512.DecodeFromUint64(1, 100)
	if tiny.IsNaN() {
		log.Fatal(tiny.ErrorDetails())
	}
	fmt.Println("Tiny scale:", tiny.Precision())
	fmt.Println("Tiny coefficient:", tiny.Coefficient())

	root, inexact := dec512.FromInt(2).SqrtRoundWithStatus(30, dec512.RoundingHalfEven)
	if root.IsNaN() {
		log.Fatal(root.ErrorDetails())
	}
	fmt.Println("Square root of two:", root.StringFixed())
	fmt.Println("Square root lost digits:", inexact)

	_, err = tiny.Dec256Exact()
	fmt.Println("Exact Dec256 conversion rejected:", err != nil)
}
