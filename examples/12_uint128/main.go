// This program parses and calculates with unsigned 128-bit integers.
package main

import (
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func main() {
	// FromString returns an error code, not a Go error or a decimal value.
	large, code := uint128.FromString("18446744073709551616")
	if code != errors.None {
		log.Fatal(code.Error())
	}
	sum, code := large.Add(uint128.FromUint64(25))
	if code != errors.None {
		log.Fatal(code.Error())
	}
	fmt.Println("Sum:", sum)

	// A byte slice uses the same parser and error contract.
	number, code := uint128.FromString([]byte("42"))
	fmt.Println("Parsed bytes:", number, "valid:", code == errors.None)

	_, code = uint128.FromString("12x")
	fmt.Println("Invalid input:", code == errors.InvalidFormat)
	_, code = uint128.FromString("340282366920938463463374607431768211456")
	fmt.Println("Input overflow:", code == errors.Overflow)

	// Checked addition reports overflow. Do not use the discarded result.
	_, code = uint128.Max.Add(uint128.One)
	fmt.Println("Addition overflow:", code == errors.Overflow)
}
