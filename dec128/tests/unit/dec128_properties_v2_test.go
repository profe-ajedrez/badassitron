package unit

import (
	"encoding/json"
	"math/big"
	"math/rand"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func v2DecimalRat(t *testing.T, d dec128.Dec128) *big.Rat {
	t.Helper()
	r, ok := new(big.Rat).SetString(d.String())
	if !ok {
		t.Fatalf("cannot parse decimal as rational: %s", d)
	}
	return r
}

func v2RandomDecimal(rng *rand.Rand) dec128.Dec128 {
	return dec128.New(uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), rng.Intn(2) == 1)
}

func TestV2QuoRemOverflowRegression(t *testing.T) {
	max := dec128.FromString(uint128.MaxUint128Str)
	for _, a := range []string{"1.2345678901234567890", "-1.2345678901234567890"} {
		v := dec128.FromString(a)
		q, r := v.QuoRem(max)
		if q.IsNaN() || r.IsNaN() || !q.IsZero() || !r.Equal(v) {
			t.Errorf("%s / max: q=%s r=%s, want zero and %s", a, q, r, a)
		}
		if mod := v.Mod(max); !mod.Equal(v) {
			t.Errorf("%s %% max: got %s", a, mod)
		}
	}
}

func TestV2QuoRemBigRatProperties(t *testing.T) {
	rng := rand.New(rand.NewSource(20261112))
	const cases = 3500
	for i := 0; i < cases; i++ {
		a, b := v2RandomDecimal(rng), v2RandomDecimal(rng)
		if b.IsZero() {
			b = dec128.One
		}
		ar, br := v2DecimalRat(t, a), v2DecimalRat(t, b)
		ratio := new(big.Rat).Quo(ar, br)
		qRef := new(big.Int).Quo(ratio.Num(), ratio.Denom()) // toward zero
		rRef := new(big.Rat).Sub(ar, new(big.Rat).Mul(new(big.Rat).SetInt(qRef), br))

		// The integer quotient and both operands' common-scale remainder must
		// fit the representation, or QuoRem is entitled to return NaN.
		commonScale := int(a.Exponent())
		if s := int(b.Exponent()); s > commonScale {
			commonScale = s
		}
		rScaled := new(big.Rat).Mul(rRef, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(commonScale)), nil)))
		if qRef.BitLen() > 128 || !rScaled.IsInt() || new(big.Int).Abs(rScaled.Num()).BitLen() > 128 {
			continue
		}
		q, r := a.QuoRem(b)
		if q.IsNaN() || r.IsNaN() {
			t.Fatalf("case %d: representable %s / %s => NaN (%v)", i, a, b, q.ErrorDetails())
		}
		if v2DecimalRat(t, q).Cmp(new(big.Rat).SetInt(qRef)) != 0 || v2DecimalRat(t, r).Cmp(rRef) != 0 {
			t.Fatalf("case %d: %s / %s => q=%s r=%s; want q=%s r=%s", i, a, b, q, r, qRef, rRef)
		}
		product := q.Mul(b)
		if !product.IsNaN() {
			reconstructed := product.Add(r)
			// An exact result can be representable even if an intermediate
			// scale-alignment overflows; skip only that intermediate case.
			if !reconstructed.IsNaN() && !reconstructed.Equal(a) {
				t.Fatalf("case %d: reconstruction failed for %s / %s", i, a, b)
			}
		}
		if !r.IsZero() && r.Sign() != a.Sign() {
			t.Fatalf("case %d: remainder sign differs from dividend", i)
		}
	}
}

func TestV2SqrtFullWidthBigIntOracle(t *testing.T) {
	oldPrecision := dec128.MaxPrecision // suite convention restores to max
	defer dec128.SetDefaultPrecision(oldPrecision)
	rng := rand.New(rand.NewSource(2026111202))
	max := dec128.New(uint128.Max, 0, false)
	cases := []dec128.Dec128{
		max, dec128.FromString("2"), dec128.FromString("0.0000000000000000001"),
		dec128.FromString("12345678901234567890123456789012345678"),
	}
	for i := 0; i < 450; i++ {
		cases = append(cases, dec128.New(uint128.Uint128{Lo: rng.Uint64(), Hi: rng.Uint64()}, uint8(rng.Intn(20)), false))
	}
	for _, prec := range []uint8{0, 1, 2, 6, 9, 10, 19} {
		dec128.SetDefaultPrecision(prec)
		for i, input := range cases {
			c := new(big.Int)
			if _, ok := c.SetString(input.Coefficient().String(), 10); !ok {
				t.Fatal("invalid coefficient")
			}
			diff := 2*int(prec) - int(input.Exponent())
			pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(absIntV2(diff))), nil)
			if diff >= 0 {
				c.Mul(c, pow)
			} else {
				c.Quo(c, pow)
			}
			want := new(big.Int).Sqrt(c)
			got := input.Sqrt()
			if got.IsNaN() {
				t.Fatalf("scale %d case %d: sqrt(%s) => %v", prec, i, input, got.ErrorDetails())
			}
			// Check exactly the coefficient after adjusting early-return cases
			// (sqrt(0) and sqrt(1) return canonical constants).
			if got.Exponent() != prec {
				got = got.Rescale(prec)
			}
			if got.IsNaN() || got.Coefficient().String() != want.String() {
				t.Fatalf("scale %d case %d: sqrt(%s) => %s, want coefficient %s / 10^%d", prec, i, input, got, want, prec)
			}
		}
	}
}

func absIntV2(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func TestV2ArithmeticAndRoundtripProperties(t *testing.T) {
	rng := rand.New(rand.NewSource(2026111203))
	for i := 0; i < 2500; i++ {
		// Bound coefficients and scales so the exact results always fit 128 bits.
		a := dec128.New(uint128.FromUint64(rng.Uint64()>>24), uint8(rng.Intn(9)), rng.Intn(2) == 1)
		b := dec128.New(uint128.FromUint64(rng.Uint64()>>24), uint8(rng.Intn(9)), rng.Intn(2) == 1)
		ar, br := v2DecimalRat(t, a), v2DecimalRat(t, b)
		for _, op := range []struct {
			name string
			got  dec128.Dec128
			want *big.Rat
		}{
			{"Add", a.Add(b), new(big.Rat).Add(ar, br)},
			{"Sub", a.Sub(b), new(big.Rat).Sub(ar, br)},
			{"Mul", a.Mul(b), new(big.Rat).Mul(ar, br)},
		} {
			if op.got.IsNaN() || v2DecimalRat(t, op.got).Cmp(op.want) != 0 {
				t.Fatalf("case %d: %s(%s,%s) => %s, want %s", i, op.name, a, b, op.got, op.want)
			}
		}
		d := v2RandomDecimal(rng)
		var textDecoded dec128.Dec128
		encoded, err := d.MarshalText()
		if err != nil || textDecoded.UnmarshalText(encoded) != nil || !d.Equal(textDecoded) {
			t.Fatalf("case %d: text roundtrip failure: %s -> %q -> %s", i, d, encoded, textDecoded)
		}
		encodedJSON, err := json.Marshal(d)
		var jsonDecoded dec128.Dec128
		if err != nil || json.Unmarshal(encodedJSON, &jsonDecoded) != nil || !d.Equal(jsonDecoded) {
			t.Fatalf("case %d: JSON roundtrip failure: %s -> %q -> %s", i, d, encodedJSON, jsonDecoded)
		}
	}
}
