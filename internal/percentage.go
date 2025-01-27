package internal

import (
	"errors"

	"github.com/alpacahq/alpacadecimal"
)

var (
	Hundred = alpacadecimal.NewFromInt(100)
	One     = alpacadecimal.NewFromInt(1)
	Zero    = alpacadecimal.NewFromInt(0)
)

// Part calculates the part from a total and a percentage
func Part(total, percent alpacadecimal.Decimal) alpacadecimal.Decimal {
	return total.Mul(percent).Div(Hundred)
}

// Total calculates the total from the part and the percentage
func Total(part, percent alpacadecimal.Decimal) (alpacadecimal.Decimal, error) {
	if percent.Equal(alpacadecimal.Zero) {
		sb := GetSB()
		defer PutSB(sb)

		sb.WriteString("divisor cant be zero calculating the total. part * 100 / percent = ")
		sb.WriteString(part.String())
		sb.WriteString(" * 100 / ")
		sb.WriteString(percent.String())
		return alpacadecimal.Zero, errors.New(sb.String())
	}

	return part.Mul(Hundred).Div(percent), nil
}

// Percentage calculates the percentage from the total and the part
func Percentage(part, total alpacadecimal.Decimal) (alpacadecimal.Decimal, error) {
	if total.Equal(alpacadecimal.Zero) {
		sb := GetSB()
		defer PutSB(sb)

		sb.WriteString("divisor cant be zero calculating the percentage. part / total = ")
		sb.WriteString(part.String())
		sb.WriteString(" * 100 / ")
		sb.WriteString(total.String())
		return alpacadecimal.Zero, errors.New(sb.String())
	}

	return part.Div(total).Mul(Hundred), nil
}
