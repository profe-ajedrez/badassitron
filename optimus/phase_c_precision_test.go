package optimus

import (
	"bytes"
	"encoding/json"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"github.com/profe-ajedrez/badassitron/dec512/uint512"
)

// phaseCPrecisionRat is deliberately independent of Optimus arithmetic.
func phaseCPrecisionRat(d Optimus) *big.Rat {
	c := d.Coefficient().BigInt()
	if d.IsNegative() {
		c.Neg(c)
	}
	return new(big.Rat).SetFrac(c, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(d.Precision())), nil))
}

// phaseCPrecisionQuantize calculates a signed integer quotient by exact
// Euclidean magnitude and remainder comparison, not by decimal formatting.
func phaseCPrecisionQuantize(v *big.Rat, scale uint8, mode RoundingMode) (*big.Int, bool) {
	n := new(big.Int).Mul(v.Num(), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil))
	neg := n.Sign() < 0
	n.Abs(n)
	den := v.Denom()
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(n, den, r)
	inexact := r.Sign() != 0
	up := false
	if inexact {
		switch mode {
		case RoundingTowardZero:
		case RoundingAwayFromZero:
			up = true
		case RoundingFloor:
			up = neg
		case RoundingCeil:
			up = !neg
		case RoundingHalfTowardZero, RoundingHalfAwayFromZero, RoundingHalfEven:
			cmp := new(big.Int).Lsh(new(big.Int).Set(r), 1).Cmp(den)
			up = cmp > 0 || cmp == 0 && (mode == RoundingHalfAwayFromZero || mode == RoundingHalfEven && q.Bit(0) == 1)
		}
	}
	if up {
		q.Add(q, big.NewInt(1))
	}
	if neg {
		q.Neg(q)
	}
	return q, inexact
}

func phaseCPrecisionExpectQuantized(t *testing.T, label string, value *big.Rat, scale uint8, mode RoundingMode, got Optimus, inexact bool) {
	t.Helper()
	coefficient, wantInexact := phaseCPrecisionQuantize(value, scale, mode)
	if inexact != wantInexact {
		t.Fatalf("%s status: got %v want %v", label, inexact, wantInexact)
	}
	if new(big.Int).Abs(new(big.Int).Set(coefficient)).BitLen() > 512 {
		if !got.IsNaN() {
			t.Fatalf("%s expected coefficient overflow, got %s", label, got)
		}
		return
	}
	if got.IsNaN() {
		t.Fatalf("%s unexpected NaN: %v", label, got.ErrorDetails())
	}
	want := new(big.Rat).SetFrac(coefficient, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil))
	if phaseCPrecisionRat(got).Cmp(want) != 0 || got.Precision() != scale {
		t.Fatalf("%s got %s(scale %d), expected %s(scale %d)", label, got.StringFixed(), got.Precision(), want.RatString(), scale)
	}
}

func phaseCPrecisionFixture(t *testing.T, c *big.Int, scale uint8, neg bool) Optimus {
	t.Helper()
	u, e := uint512.FromBigInt(c)
	if e != 0 {
		t.Fatalf("invalid coefficient %s: %v", c, e)
	}
	d := New(u, scale, neg)
	if d.IsNaN() {
		t.Fatalf("invalid fixture %s: %v", c, d.ErrorDetails())
	}
	return d
}

func TestPhaseCPrecisionRoundingWithIndependentFractions(t *testing.T) {
	for _, src := range []string{"2.5", "3.5", "-2.5", "-3.5", "2.5000001", "-2.5000001", "0.00000000000000000005", "-0.00000000000000000005", "123456789.76543210000000000001"} {
		d := FromString(src)
		if d.IsNaN() {
			t.Fatalf("bad fixture %s", src)
		}
		for _, p := range []uint8{0, 1, 2, 19, 153, 154} {
			for mode := RoundingTowardZero; mode <= RoundingHalfEven; mode++ {
				got, status := d.RescaleRoundWithStatus(p, mode)
				phaseCPrecisionExpectQuantized(t, "rescale="+src+"/"+strconv.Itoa(int(p))+"/"+strconv.Itoa(int(mode)), phaseCPrecisionRat(d), p, mode, got, status)
			}
		}
	}
	// Regression for correct exact ties and just-over-tie values.
	for _, tc := range []struct{ in, away, even string }{
		{"2.5", "3", "2"}, {"3.5", "4", "4"}, {"-2.5", "-3", "-2"}, {"-3.5", "-4", "-4"},
		{"2.5000001", "3", "3"}, {"-2.5000001", "-3", "-3"},
	} {
		d := FromString(tc.in)
		if d.Round(0).String() != tc.away || d.RoundBank(0).String() != tc.even {
			t.Fatalf("rounding %s: away %s even %s", tc.in, d.Round(0), d.RoundBank(0))
		}
	}
}

func TestPhaseCPrecisionBoundaryDivisionsAndStatus(t *testing.T) {
	rng := rand.New(rand.NewSource(0x51756c))
	maxCoefficient := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	pool := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2), big.NewInt(5), big.NewInt(10), big.NewInt(17), big.NewInt(1001), new(big.Int).Set(maxCoefficient)}
	for _, bits := range []uint{64, 128, 256, 511} {
		pivot := new(big.Int).Lsh(big.NewInt(1), bits)
		for _, offset := range []int64{-1, 0, 1} {
			pool = append(pool, new(big.Int).Add(pivot, big.NewInt(offset)))
		}
	}
	for i := 0; i < 800; i++ {
		a := phaseCPrecisionFixture(t, pool[rng.Intn(len(pool))], []uint8{0, 1, 19, 153, 154}[rng.Intn(5)], rng.Intn(2) == 0)
		b := phaseCPrecisionFixture(t, pool[1+rng.Intn(len(pool)-1)], []uint8{0, 1, 19, 153, 154}[rng.Intn(5)], rng.Intn(2) == 0)
		if b.IsZero() {
			continue
		}
		value := new(big.Rat).Quo(phaseCPrecisionRat(a), phaseCPrecisionRat(b))
		p := []uint8{0, 1, 19, 153, 154}[rng.Intn(5)]
		mode := RoundingMode(rng.Intn(int(RoundingHalfEven) + 1))
		got, status := a.DivRoundWithStatus(b, p, mode)
		phaseCPrecisionExpectQuantized(t, "div i="+strconv.Itoa(i)+" a="+a.String()+" b="+b.String(), value, p, mode, got, status)
	}
}

func TestPhaseCPrecisionMetamorphicArithmetic(t *testing.T) {
	rng := rand.New(rand.NewSource(0x9e37))
	for i := 0; i < 1100; i++ {
		var ca, cb big.Int
		bitsA := []int{1, 16, 63, 64, 65, 127, 128, 129, 255, 256, 257, 510}[rng.Intn(12)]
		bitsB := []int{1, 16, 63, 64, 65, 127, 128, 129, 255, 256, 257, 510}[rng.Intn(12)]
		ca.SetBit(&ca, bitsA-1, 1)
		cb.SetBit(&cb, bitsB-1, 1)
		if bitsA > 1 {
			ca.Add(&ca, new(big.Int).SetUint64(rng.Uint64()))
		}
		if bitsB > 1 {
			cb.Add(&cb, new(big.Int).SetUint64(rng.Uint64()))
		}
		a := phaseCPrecisionFixture(t, &ca, []uint8{0, 1, 19, 153, 154}[rng.Intn(5)], rng.Intn(2) == 0)
		b := phaseCPrecisionFixture(t, &cb, []uint8{0, 1, 19, 153, 154}[rng.Intn(5)], rng.Intn(2) == 0)
		ar, br := phaseCPrecisionRat(a), phaseCPrecisionRat(b)
		// The independent rational defines comparison even across different tiers.
		expectedCmp := ar.Cmp(br)
		if a.Compare(b) != expectedCmp || (a.Equal(b) != (expectedCmp == 0)) || b.Compare(a) != -expectedCmp {
			t.Fatalf("comparison mismatch %s %s", a, b)
		}
		canon := a.Canonical()
		if !canon.Equal(a) || !canon.Canonical().Equal(canon) || !canon.Equal(FromString(canon.String())) {
			t.Fatalf("canonical invariant %s", a)
		}
		if !a.Equal(FromString(a.String())) {
			t.Fatalf("string invariant %s", a)
		}
		text, err := a.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		var textBack Optimus
		if err = textBack.UnmarshalText(text); err != nil || !textBack.Equal(a) {
			t.Fatalf("text %s: %v", a, err)
		}
		enc, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		var back Optimus
		if err = json.Unmarshal(enc, &back); err != nil || !back.Equal(a) {
			t.Fatalf("json %s: %v", a, err)
		}
		for j, got := range []Optimus{a.Add(b), a.Sub(b), a.Mul(b)} {
			want := new(big.Rat)
			switch j {
			case 0:
				want.Add(ar, br)
			case 1:
				want.Sub(ar, br)
			case 2:
				want.Mul(ar, br)
			}
			if got.IsNaN() {
				// A result with more than 512 integer bits cannot be represented.
				if new(big.Int).Quo(new(big.Int).Abs(new(big.Int).Set(want.Num())), want.Denom()).BitLen() <= 512 {
					t.Fatalf("unexpected NaN op=%d a=%s b=%s", j, a, b)
				}
				continue
			}
			// Decimal operations are exact when the mathematical value fits at
			// its natural product/alignment scale and in 512 coefficient bits.
			scale := int(a.Precision())
			if j == 2 {
				scale += int(b.Precision())
			} else if int(b.Precision()) > scale {
				scale = int(b.Precision())
			}
			if scale <= int(MaxPrecision) {
				scaled := new(big.Rat).Mul(want, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)))
				if scaled.IsInt() && new(big.Int).Abs(new(big.Int).Set(scaled.Num())).BitLen() <= 512 && phaseCPrecisionRat(got).Cmp(want) != 0 {
					t.Fatalf("loss of exact result op=%d a=%s b=%s got=%s want=%s", j, a, b, got, want.RatString())
				}
			}
		}
		if !b.IsZero() {
			q, r := a.QuoRem(b)
			if !q.IsNaN() && !r.IsNaN() {
				combined := new(big.Rat).Add(new(big.Rat).Mul(phaseCPrecisionRat(q), br), phaseCPrecisionRat(r))
				if combined.Cmp(ar) != 0 {
					t.Fatalf("qrem equality a=%s b=%s q=%s r=%s", a, b, q, r)
				}
				if new(big.Rat).Abs(phaseCPrecisionRat(r)).Cmp(new(big.Rat).Abs(br)) >= 0 {
					t.Fatalf("remainder magnitude a=%s b=%s", a, b)
				}
				if !a.Mod(b).Equal(r) {
					t.Fatalf("mod differs from remainder a=%s b=%s", a, b)
				}
			}
		}
	}
}

func TestPhaseCPrecisionSQLAndExternalRepresentation(t *testing.T) {
	for _, src := range []string{"-0", "0.000", "1.00", "-1.000", strings.Repeat("9", 154), "0." + strings.Repeat("0", 153) + "1"} {
		original := FromString(src)
		if original.IsNaN() {
			t.Fatalf("fixture %s", src)
		}
		var dst SQLDecimal
		if err := dst.Scan(src); err != nil {
			t.Fatalf("exact SQL scan %s: %v", src, err)
		}
		if !dst.Valid || !dst.Decimal.Equal(original) {
			t.Fatalf("SQL representation %s", src)
		}
		v, err := dst.Value()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal([]byte(v.(string)), []byte(original.String())) {
			t.Fatalf("SQL value mismatch %s", src)
		}
	}
	var d Optimus
	if err := json.Unmarshal([]byte("null"), &d); err != nil || !d.IsZero() {
		t.Fatal("historical JSON null-to-zero")
	}
	if err := json.Unmarshal([]byte("-0"), &d); err != nil || !d.IsZero() {
		t.Fatal("negative zero")
	}
}

// The v7 uint128.Mul fast path could return a wrapped denominator when
// aligning a scale-zero coefficient to scale 19, yielding quotient -23
// instead of the mathematically correct zero.
func TestPhaseCPrecisionQuoRemAlignmentOverflowRegression(t *testing.T) {
	a := FromString("-8507059173023461588.1519253575114968768")
	b := FromString("34389314001497466858")
	q, r := a.QuoRem(b)
	if !q.IsZero() || !r.Equal(a) {
		t.Fatalf("alignment overflow: quotient=%s remainder=%s, expected 0 and %s", q, r, a)
	}
	if !a.Mod(b).Equal(a) {
		t.Fatalf("modulo lost exact remainder: %s", a.Mod(b))
	}
}

// Quantized sqrt is checked by rational square inequalities. The midpoint
// is exact: sqrt(x) >= k+1/2 iff 4*x >= (2*k+1)^2.
func phaseCPrecisionSqrtOracle(value *big.Rat, scale uint8, mode RoundingMode) (*big.Int, bool) {
	scaled := new(big.Rat).Mul(value, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)*2), nil)))
	lo := big.NewInt(0)
	hi := new(big.Int).Lsh(big.NewInt(1), uint((scaled.Num().BitLen()+1)/2+1))
	sq := new(big.Int)
	for new(big.Int).Sub(hi, lo).Cmp(big.NewInt(1)) > 0 {
		mid := new(big.Int).Rsh(new(big.Int).Add(lo, hi), 1)
		sq.Mul(mid, mid)
		if new(big.Rat).SetInt(sq).Cmp(scaled) <= 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	exact := new(big.Rat).SetInt(new(big.Int).Mul(new(big.Int).Set(lo), lo)).Cmp(scaled) == 0
	if !exact {
		twicePlusOne := new(big.Int).Add(new(big.Int).Lsh(new(big.Int).Set(lo), 1), big.NewInt(1))
		boundary := new(big.Int).Mul(twicePlusOne, twicePlusOne)
		four := new(big.Rat).Mul(scaled, big.NewRat(4, 1))
		cmp := four.Cmp(new(big.Rat).SetInt(boundary))
		if mode == RoundingAwayFromZero || mode == RoundingCeil || cmp > 0 && (mode == RoundingHalfEven || mode == RoundingHalfAwayFromZero || mode == RoundingHalfTowardZero) || cmp == 0 && (mode == RoundingHalfAwayFromZero || mode == RoundingHalfEven && lo.Bit(0) == 1) {
			lo.Add(lo, big.NewInt(1))
		}
	}
	return lo, !exact
}

func TestPhaseCPrecisionSqrtDirectedOracle(t *testing.T) {
	for _, input := range []string{"0", "1", "2", "2.25", "3.5", "4", "6.25", "0.000000000000000000000000000000000001", "0." + strings.Repeat("0", 153) + "1", strings.Repeat("9", 100), "340282366920938463463374607431768211456"} {
		d := FromString(input)
		if d.IsNaN() {
			t.Fatalf("invalid fixture %s", input)
		}
		for _, p := range []uint8{0, 1, 19, 153, 154} {
			for mode := RoundingTowardZero; mode <= RoundingHalfEven; mode++ {
				got, inx := d.SqrtRoundWithStatus(p, mode)
				oracle, expectedInx := phaseCPrecisionSqrtOracle(phaseCPrecisionRat(d), p, mode)
				if inx != expectedInx {
					t.Fatalf("sqrt status x=%s p=%d mode=%d got=%v want=%v", input, p, mode, inx, expectedInx)
				}
				if oracle.BitLen() > 512 {
					if !got.IsNaN() {
						t.Fatalf("expected sqrt overflow %s p=%d", input, p)
					}
					continue
				}
				if got.IsNaN() || got.Precision() != p || got.Coefficient().BigInt().Cmp(oracle) != 0 {
					t.Fatalf("sqrt x=%s p=%d mode=%d got=%s coefficient=%s expected=%s", input, p, mode, got, got.Coefficient(), oracle)
				}
			}
		}
	}
}

func TestPhaseCPrecisionPowersAndExactTypeConversions(t *testing.T) {
	for _, input := range []string{"0", "1", "-1", "1.0", "1.250", "0.2", "-0.25", "10000000000000000000000000000000000000"} {
		x := FromString(input)
		for n := -7; n <= 7; n++ {
			got := x.PowInt(n)
			if x.IsZero() && n < 0 {
				if !got.IsNaN() {
					t.Fatalf("expected zero power negative NaN")
				}
				continue
			}
			raw := phaseCPrecisionRat(x)
			num, den := new(big.Int).Set(raw.Num()), new(big.Int).Set(raw.Denom())
			if n < 0 {
				num, den = den, num
			}
			exponent := int64(n)
			if exponent < 0 {
				exponent = -exponent
			}
			num.Exp(num, big.NewInt(exponent), nil)
			den.Exp(den, big.NewInt(exponent), nil)
			expected := new(big.Rat).SetFrac(num, den)
			if got.IsNaN() {
				// More than 512 integral bits is an unrepresentable value.
				integer := new(big.Int).Quo(new(big.Int).Abs(new(big.Int).Set(expected.Num())), expected.Denom())
				if integer.BitLen() <= 512 {
					t.Fatalf("unexpected power NaN %s ^ %d", x, n)
				}
				continue
			}
			if expected.Denom().Cmp(big.NewInt(1)) == 0 || n >= 0 && int(x.exp)*n <= int(MaxPrecision) && new(big.Int).Exp(x.coefBig(), big.NewInt(int64(n)), nil).BitLen() <= 512 {
				if phaseCPrecisionRat(got).Cmp(expected) != 0 {
					t.Fatalf("power lost exactness %s ^ %d = %s, expected %s", x, n, got, expected.RatString())
				}
			}
		}
		if x.IsNaN() {
			continue
		}
		if out, err := x.Dec512Exact(); err != nil || !FromDec512(out).Equal(x) {
			t.Fatalf("512 conversion %s: %v", x, err)
		}
		if x.coefBig().BitLen() <= 128 && x.Canonical().exp <= 38 {
			if out, err := x.Dec128Exact(); err != nil || !FromDec128(out).Equal(x) {
				t.Fatalf("128 conversion %s: %v", x, err)
			}
		}
	}
}

func FuzzPhaseCPrecisionQuoRemAndDivision(f *testing.F) {
	f.Add(uint64(0), uint64(123456789), uint64(0), uint64(3), uint8(0), uint8(0))
	f.Add(uint64(0x4000000000000000), uint64(0xd98aecb578984ec0), uint64(1), uint64(0xdd3f62e1c89fa7ea), uint8(19), uint8(0))
	f.Add(uint64(0xffffffffffffffff), uint64(0xffffffffffffffff), uint64(0), uint64(1), uint8(154), uint8(0))
	f.Add(uint64(0xffffffffffffffff), uint64(0xffffffffffffffff), uint64(0xffffffffffffffff), uint64(0xffffffffffffffff), uint8(0), uint8(154))
	f.Fuzz(func(t *testing.T, aHi, aLo, bHi, bLo uint64, scaleA, scaleB uint8) {
		a := NewFromUint128(uint128.Uint128{Hi: aHi, Lo: aLo}, scaleA%155, aHi&1 != 0)
		b := NewFromUint128(uint128.Uint128{Hi: bHi, Lo: bLo}, scaleB%155, bHi&1 != 0)
		if b.IsZero() {
			return
		}
		ar, br := phaseCPrecisionRat(a), phaseCPrecisionRat(b)
		expected := new(big.Rat).Quo(ar, br)
		p := uint8(19)
		got, inx := a.DivRoundWithStatus(b, p, RoundingHalfEven)
		phaseCPrecisionExpectQuantized(t, "fuzz division a="+a.String()+" b="+b.String(), expected, p, RoundingHalfEven, got, inx)
		q, r := a.QuoRem(b)
		if q.IsNaN() || r.IsNaN() {
			return
		}
		combined := new(big.Rat).Add(new(big.Rat).Mul(phaseCPrecisionRat(q), br), phaseCPrecisionRat(r))
		if combined.Cmp(ar) != 0 || !r.Equal(a.Mod(b)) {
			t.Fatalf("inexact quo/rem a=%s b=%s q=%s r=%s", a, b, q, r)
		}
	})
}

func TestPhaseCPrecisionSpecialValuesAndDefaultPrecision(t *testing.T) {
	// In the inherited contract NaN values compare equal regardless of cause,
	// and the ordering places NaN before all ordinary numbers.
	invalid := FromString("invalid")
	divideZero := One.Div(Zero)
	if !invalid.IsNaN() || !divideZero.IsNaN() || !invalid.Equal(divideZero) || invalid.Compare(divideZero) != 0 || invalid.Compare(Zero) >= 0 {
		t.Fatal("NaN legacy comparison contract changed")
	}
	if stringMustText, err := invalid.MarshalText(); err != nil || string(stringMustText) != "NaN" {
		t.Fatalf("NaN text: %q %v", stringMustText, err)
	}
	textDestination := One
	if err := textDestination.UnmarshalText([]byte("NaN")); err == nil || !textDestination.Equal(One) {
		t.Fatal("NaN text must remain invalid on input and preserve the destination")
	}
	jsonNaN, err := json.Marshal(invalid)
	if err != nil || string(jsonNaN) != `"NaN"` {
		t.Fatalf("NaN JSON: %s %v", jsonNaN, err)
	}
	if !FromString("-0.00000").Equal(Zero) || FromString("-0.00000").StringFixed() != "0.00000" {
		t.Fatal("negative-zero normalization changed")
	}
	if !FromString("2.59").Trunc(1).Equal(FromString("2.5")) ||
		!FromString("-2.59").Rescale(1).Equal(FromString("-2.5")) {
		t.Fatal("explicit truncation must not become rounding")
	}
	SetDefaultPrecision(19)
	defer SetDefaultPrecision(19)
	if got := One.Div(FromInt(7)); got.Precision() != 19 {
		t.Fatalf("default non-terminating division precision: %d", got.Precision())
	}
	SetDefaultPrecision(7)
	if got := One.Div(FromInt(7)); got.Precision() != 7 {
		t.Fatalf("configured division precision: %d", got.Precision())
	}
	if got := One.DivRound(FromInt(7), 19, RoundingHalfEven); got.Precision() != 19 {
		t.Fatalf("explicit division precision was overridden: %d", got.Precision())
	}
}
