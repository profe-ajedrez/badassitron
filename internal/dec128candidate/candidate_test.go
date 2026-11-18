package dec128candidate_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"testing"

	jk "github.com/jokruger/dec128"
	legacy "github.com/profe-ajedrez/badassitron/dec128"
	lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
	candidate "github.com/profe-ajedrez/badassitron/internal/dec128candidate"
)

func same(t *testing.T, label string, want legacy.Dec128, got candidate.Dec128) {
	t.Helper()
	e1, e2 := "", ""
	if e := want.ErrorDetails(); e != nil {
		e1 = e.Error()
	}
	if e := got.ErrorDetails(); e != nil {
		e2 = e.Error()
	}
	if want.String() != got.String() || want.StringFixed() != got.StringFixed() || want.Precision() != got.Precision() || want.IsNaN() != got.IsNaN() || want.IsNegative() != got.IsNegative() || want.StoredNegative() != got.StoredNegative() || e1 != e2 || want.Coefficient().Hi != got.Coefficient().Hi || want.Coefficient().Lo != got.Coefficient().Lo {
		t.Fatalf("%s mismatch: want %s scale=%d storedNeg=%t code=%s coef=%v; got %s scale=%d storedNeg=%t code=%s coef=%v", label, want.StringFixed(), want.Precision(), want.StoredNegative(), e1, want.Coefficient(), got.StringFixed(), got.Precision(), got.StoredNegative(), e2, got.Coefficient())
	}
}

// These assertions exercise concrete Go interface satisfaction, complementing
// the declaration-only AST checker; they do not certify external consumers.
var (
	_ fmt.Stringer             = candidate.Dec128{}
	_ encoding.TextMarshaler   = candidate.Dec128{}
	_ encoding.TextUnmarshaler = (*candidate.Dec128)(nil)
	_ json.Marshaler           = candidate.Dec128{}
	_ json.Unmarshaler         = (*candidate.Dec128)(nil)
	_ driver.Valuer            = candidate.Dec128{}
	_ sql.Scanner              = candidate.Dec128{} // Historical value receiver.
)

func TestCandidatePublicCodecInterfaces(t *testing.T) {
	value := candidate.FromString("123.450")
	text, err := value.MarshalText()
	if err != nil || string(text) != "123.45" {
		t.Fatalf("text codec: %q, %v", text, err)
	}
	var restored candidate.Dec128
	if err := restored.UnmarshalText(text); err != nil || restored.String() != "123.45" {
		t.Fatalf("text unmarshal: %v, %v", restored, err)
	}
	encoded, err := json.Marshal(value)
	if err != nil || string(encoded) != `"123.45"` {
		t.Fatalf("json codec: %q, %v", encoded, err)
	}
	if err := json.Unmarshal(encoded, &restored); err != nil || restored.String() != "123.45" {
		t.Fatalf("json unmarshal: %v, %v", restored, err)
	}
	if err := candidate.ScanDestination(&restored).Scan("44.20"); err != nil || restored.StringFixed() != "44.20" {
		t.Fatalf("SQL scanner: %v, %v", restored, err)
	}
	if err := candidate.ScanDestination(nil).Scan("1"); err == nil {
		t.Fatal("nil SQL scan destination accepted")
	}
}

func TestCandidateSignedZeroAndRounding(t *testing.T) {
	for scale := uint8(0); scale <= 19; scale++ {
		for _, neg := range []bool{false, true} {
			a := legacy.NewStoredZero(scale, neg)
			b := candidate.NewStoredZero(scale, neg)
			same(t, "signed-zero", a, b)
			for target := uint8(0); target <= 19; target++ {
				same(t, "zero rescale", a.Rescale(target), b.Rescale(target))
				for m := uint8(0); m < 7; m++ {
					w, wi := a.RescaleRoundWithStatus(target, legacy.RoundingMode(m))
					g, gi := b.RescaleRoundWithStatus(target, candidate.RoundingMode(m))
					same(t, "zero rescale-round", w, g)
					if wi != gi {
						t.Fatal("zero inexact mismatch")
					}
				}
			}
		}
	}
}
func TestCandidateRandomizedSqrt(t *testing.T) {
	r := rand.New(rand.NewSource(20261117))
	samples := []string{"0", "-0", "0.00", "-0.000", "1", "1.00", "4", "2", "0.25", "0.0001", "0.125", "9.99", "10", "-1", "-1.23", "340282366920938463463374607431768211455", "0.0000000000000000001"}
	for i := 0; i < 1500; i++ {
		hi, lo := r.Uint64(), r.Uint64()
		if r.Intn(2) == 0 {
			hi = 0
		}
		p := uint8(r.Intn(20))
		neg := r.Intn(2) == 0
		c := lu.Uint128{Hi: hi, Lo: lo}
		v := legacy.New(c, p, neg)
		samples = append(samples, v.StringFixed())
	}
	for _, s := range samples {
		want := legacy.FromString(s)
		got := candidate.FromString(s)
		same(t, "parse:"+s, want, got)
		for _, p := range []uint8{0, 1, 2, 9, 10, 19, 20} {
			for m := uint8(0); m < 7; m++ {
				w, wi := want.SqrtRoundWithStatus(p, legacy.RoundingMode(m))
				g, gi := got.SqrtRoundWithStatus(p, candidate.RoundingMode(m))
				same(t, fmt.Sprintf("sqrtRound %s p=%d mode=%d", s, p, m), w, g)
				if wi != gi {
					t.Fatalf("sqrt status %q p=%d mode=%d: want %t got %t", s, p, m, wi, gi)
				}
			}
		}
		same(t, "sqrt:"+s, want.Sqrt(), got.Sqrt())
	}
}
func TestCandidateRandomizedArithmetic(t *testing.T) {
	r := rand.New(rand.NewSource(20261118))
	for i := 0; i < 5000; i++ {
		a := legacy.New(lu.Uint128{Hi: r.Uint64(), Lo: r.Uint64()}, uint8(r.Intn(20)), r.Intn(2) == 0)
		b := legacy.New(lu.Uint128{Hi: r.Uint64(), Lo: r.Uint64()}, uint8(r.Intn(20)), r.Intn(2) == 0)
		ac := candidate.New(a.Coefficient(), a.Precision(), a.StoredNegative())
		bc := candidate.New(b.Coefficient(), b.Precision(), b.StoredNegative())
		same(t, "add", a.Add(b), ac.Add(bc))
		same(t, "sub", a.Sub(b), ac.Sub(bc))
		same(t, "mul", a.Mul(b), ac.Mul(bc))
		same(t, "div", a.Div(b), ac.Div(bc))
		same(t, "mod", a.Mod(b), ac.Mod(bc))
		wq, wr := a.QuoRem(b)
		gq, gr := ac.QuoRem(bc)
		same(t, "quorem/q", wq, gq)
		same(t, "quorem/r", wr, gr)
		if i%5 == 0 {
			p := uint8(r.Intn(20))
			m := uint8(r.Intn(7))
			w, wi := a.MulRoundWithStatus(b, p, legacy.RoundingMode(m))
			g, gi := ac.MulRoundWithStatus(bc, p, candidate.RoundingMode(m))
			same(t, "mulround", w, g)
			if wi != gi {
				t.Fatal("mul inexact mismatch")
			}
			w, wi = a.DivRoundWithStatus(b, p, legacy.RoundingMode(m))
			g, gi = ac.DivRoundWithStatus(bc, p, candidate.RoundingMode(m))
			same(t, "divround", w, g)
			if wi != gi {
				t.Fatal("div inexact mismatch")
			}
		}
	}
}

var benchLegacy legacy.Dec128
var benchCandidate candidate.Dec128
var benchJK jk.Dec128

func BenchmarkCandidateVsLegacy(b *testing.B) {
	for _, op := range []string{"Add", "Mul", "Div", "Parse"} {
		for _, v := range []string{"legacy", "candidate", "upstream"} {
			b.Run(op+"/"+v, func(b *testing.B) {
				la, lb := legacy.FromString("123456789.012345"), legacy.FromString("6789.010020")
				ca, cb := candidate.FromString("123456789.012345"), candidate.FromString("6789.010020")
				ja, jb := jk.FromString("123456789.012345"), jk.FromString("6789.010020")
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					switch v {
					case "legacy":
						switch op {
						case "Add":
							benchLegacy = la.Add(lb)
						case "Mul":
							benchLegacy = la.Mul(lb)
						case "Div":
							benchLegacy = la.Div(lb)
						case "Parse":
							benchLegacy = legacy.FromString("123456789.012345")
						}
					case "candidate":
						switch op {
						case "Add":
							benchCandidate = ca.Add(cb)
						case "Mul":
							benchCandidate = ca.Mul(cb)
						case "Div":
							benchCandidate = ca.Div(cb)
						case "Parse":
							benchCandidate = candidate.FromString("123456789.012345")
						}
					case "upstream":
						switch op {
						case "Add":
							benchJK = ja.Add(jb)
						case "Mul":
							benchJK = ja.Mul(jb)
						case "Div":
							benchJK = ja.Div(jb)
						case "Parse":
							benchJK = jk.FromString("123456789.012345")
						}
					}
				}
			})
		}
	}
}
func TestCandidateFloatSpecial(t *testing.T) {
	same(t, "float nan", legacy.FromFloat64(math.NaN()), candidate.FromFloat64(math.NaN()))
	same(t, "float inf", legacy.FromFloat64(math.Inf(1)), candidate.FromFloat64(math.Inf(1)))
}

// Run serially: upstream configuration is process global and unsafe to mutate
// while unrelated goroutines use upstream decimal arithmetic.
func TestCandidateIndependentOfUpstreamProcessPolicy(t *testing.T) {
	oldScale := jk.DefaultScale()
	oldPolicy := jk.CurrentLossPolicy()
	oldRounding := jk.ArithmeticRounding()
	oldTrim := jk.TrimOutput()
	defer func() {
		jk.SetDefaultScale(oldScale)
		jk.SetArithmeticRounding(oldRounding)
		jk.SetLossPolicy(oldPolicy)
		jk.SetTrimOutput(oldTrim)
	}()
	cases := []struct{ a, b string }{
		{"1.2350000000", "2.1500000000"},
		{"-0.0000000000000000001", "3.1234567890"},
		{"123456789012345678.0123456789012345678", "1.2345678900000"},
		{"9999999999999999999.9000", "0.0001234567"},
	}
	base := make([][]candidate.Dec128, len(cases))
	for i, c := range cases {
		a, b := candidate.FromString(c.a), candidate.FromString(c.b)
		base[i] = []candidate.Dec128{a.Add(b), a.Sub(b), a.Mul(b), a.Div(b), a.DivRound(b, 6, candidate.RoundingHalfEven), a.MulRound(b, 6, candidate.RoundingHalfEven), a.RescaleRound(2, candidate.RoundingHalfAwayFromZero), a.SqrtRound(4, candidate.RoundingHalfEven)}
	}
	for _, policy := range []jk.LossPolicy{jk.LossRound, jk.LossNaNOnUnderflow, jk.LossNaNOnInexact} {
		jk.SetArithmeticRounding(jk.ROUND_UP)
		jk.SetLossPolicy(policy)
		jk.SetDefaultScale(2)
		jk.SetTrimOutput(true)
		for i, c := range cases {
			a, b := candidate.FromString(c.a), candidate.FromString(c.b)
			actual := []candidate.Dec128{a.Add(b), a.Sub(b), a.Mul(b), a.Div(b), a.DivRound(b, 6, candidate.RoundingHalfEven), a.MulRound(b, 6, candidate.RoundingHalfEven), a.RescaleRound(2, candidate.RoundingHalfAwayFromZero), a.SqrtRound(4, candidate.RoundingHalfEven)}
			for j := range actual {
				got, want := actual[j], base[i][j]
				if got.StringFixed() != want.StringFixed() || got.ErrorCode() != want.ErrorCode() || got.Precision() != want.Precision() || got.StoredNegative() != want.StoredNegative() {
					t.Fatalf("global upstream settings altered candidate: policy=%v item=%d op=%d: got %s want %s", policy, i, j, got.StringFixed(), want.StringFixed())
				}
			}
			if s := a.StringFixed(); s == "" {
				t.Fatal("empty candidate string")
			}
			_, _ = a.Value()
		}
	}
}

func TestCandidateDivisionSignedZeroAtScaleZero(t *testing.T) {
	legacy.SetDefaultPrecision(0)
	candidate.SetDefaultPrecision(0)
	defer legacy.SetDefaultPrecision(19)
	defer candidate.SetDefaultPrecision(19)
	w := legacy.FromString("-1").Div(legacy.FromString("1000"))
	g := candidate.FromString("-1").Div(candidate.FromString("1000"))
	same(t, "underflow negative zero", w, g)
	if !g.IsZero() || !g.StoredNegative() {
		t.Fatal("candidate must retain raw sign on zero quotient")
	}
}
