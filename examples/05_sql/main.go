// This program demonstrates SQL conversion without opening a database.
package main

import (
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/optimus"
)

func main() {
	// SQLDecimal keeps SQL NULL separate from numeric zero.
	var amount optimus.SQLDecimal
	if err := amount.Scan([]byte("123.4500")); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Present:", amount.Valid)
	fmt.Println("Scanned:", amount.Decimal.StringFixed())
	value, err := amount.Value()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Value for SQL:", value)

	// Strict scans reject floats. The previous destination stays unchanged.
	err = amount.Scan(float64(9.99))
	fmt.Println("Float rejected:", err != nil)
	fmt.Println("Previous value:", amount.Decimal.StringFixed())

	if err := amount.Scan(nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Null is present:", amount.Valid)
	value, err = amount.Value()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Null writes nil:", value == nil)
	if err := amount.Scan("0"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Zero is present:", amount.Valid)

	// Dec128 needs its scanner adapter to update a destination variable.
	// With a real database, use rows.Scan(dec128.ScanDestination(&fixed)).
	var fixed dec128.Dec128
	if err := dec128.ScanDestination(&fixed).Scan("42.500"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Dec128 adapter:", fixed.StringFixed())
	if err := dec128.ScanInto(&fixed, []byte("17.25")); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Dec128 direct scan:", fixed)
}
