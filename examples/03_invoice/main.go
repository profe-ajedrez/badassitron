// This program applies an example discount and tax policy to one invoice line.
package main

import (
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
	unitPrice := optimus.FromString("19.95")
	quantity := optimus.FromInt(3)
	discountRate := optimus.FromString("0.10")
	taxRate := optimus.FromString("0.18")
	mode := optimus.RoundingHalfEven

	// Policy for this example: round each line amount to two decimal places.
	// Rates are fractions: 0.10 means ten percent.
	subtotal := unitPrice.Mul(quantity)
	discount := subtotal.MulRound(discountRate, 2, mode)
	net := subtotal.Sub(discount)
	tax := net.MulRound(taxRate, 2, mode)
	total := net.Add(tax)

	for _, value := range []optimus.Optimus{subtotal, discount, net, tax, total} {
		if value.IsNaN() {
			log.Fatal(value.ErrorDetails())
		}
	}
	fmt.Println("Subtotal:", subtotal.StringFixed())
	fmt.Println("Discount:", discount.StringFixed())
	fmt.Println("Net:", net.StringFixed())
	fmt.Println("Tax:", tax.StringFixed())
	fmt.Println("Total:", total.StringFixed())
}
