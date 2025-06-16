package withdec128

import "github.com/profe-ajedrez/badassitron/dec128"

type Stage int8
type Type int8
type Mode int8

const (
	Natural Stage = 0
	Overtax Stage = 1
	Bypass  Stage = 2

	Percentual Type = 0
	Amount     Type = 1
	AmountLine Type = 2

	FromUV    = 0
	FromGross = 1
)

func Zero() dec128.Dec128    { return dec128.Decimal0.Copy() }
func One() dec128.Dec128     { return dec128.Decimal1.Copy() }
func Two() dec128.Dec128     { return dec128.Decimal2.Copy() }
func Ten() dec128.Dec128     { return dec128.Decimal10.Copy() }
func Hundred() dec128.Dec128 { return dec128.Decimal100.Copy() }
