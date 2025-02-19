package internal

import (
	"errors"

	"github.com/profe-ajedrez/badassitron/dec128"
)

var (
	Hundred = dec128.NewFromInt(100)
	One     = dec128.NewFromInt(1)
	Zero    = dec128.NewFromInt(0)
)

// Part calculates the part from a total and a percentage
func Part(total, percent dec128.Dec128) dec128.Dec128 {
	return total.Mul(percent).Div(Hundred)
}

// Total calculates the total from the part and the percentage
func Total(part, percent dec128.Dec128) (dec128.Dec128, error) {
	if percent.Equal(dec128.Zero) {
		sb := GetSB()
		defer PutSB(sb)

		sb.WriteString("divisor cant be zero calculating the total. part * 100 / percent = ")
		sb.WriteString(part.String())
		sb.WriteString(" * 100 / ")
		sb.WriteString(percent.String())
		return dec128.Zero, errors.New(sb.String())
	}

	return part.Mul(Hundred).Div(percent), nil
}

// Percentage calculates the percentage from the total and the part
func Percentage(part, total dec128.Dec128) (dec128.Dec128, error) {
	if total.Equal(dec128.Zero) {
		sb := GetSB()
		defer PutSB(sb)

		sb.WriteString("divisor cant be zero calculating the percentage. part / total = ")
		sb.WriteString(part.String())
		sb.WriteString(" * 100 / ")
		sb.WriteString(total.String())
		return dec128.Zero, errors.New(sb.String())
	}

	return part.Div(total).Mul(Hundred), nil
}
