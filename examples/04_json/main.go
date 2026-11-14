// This program writes decimal values to JSON and reads them back.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/profe-ajedrez/badassitron/optimus"
)

type payment struct {
	Amount optimus.Optimus `json:"amount"`
}

func main() {
	original := payment{Amount: optimus.FromString("123.4500")}
	if original.Amount.IsNaN() {
		log.Fatal(original.Amount.ErrorDetails())
	}
	data, err := json.Marshal(original)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("JSON:", string(data))

	var restored payment
	if err := json.Unmarshal(data, &restored); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Same value:", restored.Amount.Equal(original.Amount))
	fmt.Println("Restored scale:", restored.Amount.Precision())

	// Read decimal text directly, without first converting through float64.
	if err := json.Unmarshal([]byte(`{"amount":"0.1234567890123456789012345"}`), &restored); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Long decimal:", restored.Amount)

	// JSON null becomes numeric zero in the existing Optimus contract.
	if err := json.Unmarshal([]byte(`{"amount":null}`), &restored); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Null becomes zero:", restored.Amount.IsZero())
}
