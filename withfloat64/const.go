package withfloat64

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

	FromNet   = 0
	FromGross = 1

	FromUV = 0
)

func Zero() float64    { return 0 }
func One() float64     { return 1 }
func Two() float64     { return 2 }
func Ten() float64     { return 10 }
func Hundred() float64 { return 100 }
