// This program runs the existing decimal calculation handlers in order.
package main

import (
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/withdec128"
	"github.com/profe-ajedrez/badassitron/withdec128/handler"
)

func main() {
	input := &withdec128.Input{
		UV:   dec128.FromString("100.00"),
		QTY:  dec128.FromInt(10),
		Disc: dec128.FromInt(10), // This API takes 10 for ten percent.
		TaxList: []*withdec128.InputTax{
			{
				Id:        1,
				NameValue: "Example tax",
				CodeValue: "EXAMPLE",
				V:         dec128.FromInt(16), // Sixteen percent.
				Typee:     withdec128.Percentual,
				Stagee:    withdec128.Natural,
			},
		},
	}
	options := &withdec128.Options{
		Prec:    2,
		Process: withdec128.FromUV,
	}
	output := &withdec128.Output{}

	// Taxer needs the net totals. Grosser needs both the net and tax totals.
	err := handler.Next(options, input, output,
		handler.EntryValidation,
		handler.Bootstrap,
		handler.Netter,
		handler.Taxer,
		handler.Grosser,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Prec does not round these totals automatically. Choose the output scale.
	for _, item := range []struct {
		name  string
		value dec128.Dec128
	}{
		{"Before discount", output.NetWD()},
		{"Discount", output.Discount()},
		{"Net", output.Net()},
		{"Tax", output.Tax()},
		{"Gross", output.Gross()},
	} {
		published := item.value.RescaleRound(2, dec128.RoundingHalfEven)
		if published.IsNaN() {
			log.Fatal(published.ErrorDetails())
		}
		fmt.Println(item.name+":", published.StringFixed())
	}
}
