package optimus

// The slow Mul path works on one exact 1024-bit product. It retains the
// compact Optimus representation and the old auto() scale-selection rules.
// In particular, trailing decimal zeroes are removed only when an exact
// result is otherwise unrepresentable; rounding always uses the original
// (never an already-rounded) product.

import (
	"math"
	"math/big"
	"math/bits"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"github.com/profe-ajedrez/badassitron/internal/decimalwide"
	"github.com/profe-ajedrez/badassitron/internal/uintwide"
)

// mulLegacy is the original general implementation, retained as a behavioral
// oracle for differential tests and as an available fallback.
func (d Optimus) mulLegacy(b Optimus) Optimus {
	a, c := d.core(), b.core()
	return auto(decimalwide.Number{
		Coef:  new(big.Int).Mul(a.Coef, c.Coef),
		Scale: a.Scale + c.Scale,
		Neg:   a.Neg != c.Neg,
	})
}

// mulFixed02B is only invoked after the existing zero, small and fitting
// fixed-width fast paths. Both operands are therefore finite and nonzero.
func (d Optimus) mulFixed02B(b Optimus) Optimus {
	a, na := fixed01Words(d)
	c, nb := fixed01Words(b)
	var product [16]uint64
	n := na + nb
	uintwide.MulFull(product[:n], a[:na], c[:nb])
	n = fixed01Trim(&product, n)
	scale := int(d.exp) + int(b.exp)
	neg := d.neg != b.neg

	// auto() first strips *exact* decimal trailing zeros, but only as long
	// as the value cannot otherwise fit into the available scale/width.
	// The first nonzero remainder ends exact stripping; the candidate must
	// then be rounded directly from the exact coefficient.
	for scale > 0 && (scale > int(MaxPrecision) || n > 8) {
		trial := product
		trialN, rem := fixed01DivChunk(&trial, n, 10)
		if rem != 0 {
			break
		}
		product, n = trial, trialN
		scale--
	}
	if scale <= int(MaxPrecision) && n <= 8 {
		return fixed01Pack(&product, n, uint8(scale), neg)
	}

	// Divide the remaining *exact* coefficient in place. Keep the most
	// significant discarded chunk and a sticky bit for all lower chunks:
	// these are sufficient to perform one final half-even rounding without
	// constructing a big.Int or a power of ten.
	sticky := false
	var discarded, divisor uint64
	for scale > 0 {
		step := 1
		switch {
		case scale > int(MaxPrecision):
			// No scale above 154 is a representable candidate.
			step = min(19, scale-int(MaxPrecision))
		case n > 9:
			// At least 577 bits. Dividing by <= 10^19 (< 2^64)
			// cannot produce a <= 512-bit coefficient at any skipped scale.
			step = min(19, scale)
		}
		if discarded != 0 {
			sticky = true
		}
		divisor = fixed01Power10(step)
		n, discarded = fixed01DivChunk(&product, n, divisor)
		scale -= step
		if scale > int(MaxPrecision) || n > 8 {
			continue
		}
		// The quotient fits. Rounding may still carry into bit 513; in
		// that exceptional case lower the target scale once more.
		up := discarded > divisor/2 ||
			(discarded == divisor/2 && (sticky || product[0]&1 != 0))
		if up && n == 8 {
			allMax := true
			for i := 0; i < 8; i++ {
				if product[i] != math.MaxUint64 {
					allMax = false
					break
				}
			}
			if allMax {
				continue
			}
		}
		if up {
			carry := uint64(1)
			for i := 0; i < n && carry != 0; i++ {
				product[i], carry = bits.Add64(product[i], 0, carry)
			}
			if carry != 0 {
				product[n] = carry
				n++
			}
		}
		return fixed01Pack(&product, n, uint8(scale), neg && n != 0)
	}
	// Nothing at or above scale zero can represent an integer larger than
	// 512 bits. Never saturate or silently lose integer digits.
	return NaN(errors.Overflow)
}
