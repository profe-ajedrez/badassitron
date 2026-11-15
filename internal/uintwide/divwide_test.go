package uintwide

import (
	"encoding/binary"
	"math/big"
	"math/rand"
	"testing"
)

func bigFromWords(v []uint64) *big.Int {
	b := make([]byte, len(v)*8)
	for i, w := range v {
		binary.BigEndian.PutUint64(b[(len(v)-1-i)*8:], w)
	}
	return new(big.Int).SetBytes(b)
}
func TestQuoRemWideDifferential(t *testing.T) {
	rnd := rand.New(rand.NewSource(20261114))
	for i := 0; i < 10000; i++ {
		var u, v, q, r [64]uint64
		for j := 0; j < 64; j++ {
			u[j] = rnd.Uint64()
			v[j] = rnd.Uint64()
		}
		// Produce significantly different effective widths and near-equal operands.
		if i%3 == 0 {
			for j := 16; j < 64; j++ {
				v[j] = 0
			}
		}
		if i%5 == 0 {
			for j := 32; j < 64; j++ {
				u[j] = 0
			}
		}
		if i%7 == 0 {
			v = u
			v[0]++
		}
		if i%11 == 0 {
			v[0] = 1
			for j := 1; j < 64; j++ {
				v[j] = 0
			}
		}
		if i%13 == 0 {
			v[0] = 0
			v[1] = 1
			for j := 2; j < 64; j++ {
				v[j] = 0
			}
		}
		if i%17 == 0 {
			v[0] = ^uint64(0)
			for j := 1; j < 64; j++ {
				v[j] = 0
			}
		}
		if !QuoRemWide(q[:], r[:], u[:], v[:]) {
			t.Fatal("unexpected division by zero")
		}
		bigU, bigV := bigFromWords(u[:]), bigFromWords(v[:])
		wantQ, wantR := new(big.Int), new(big.Int)
		wantQ.QuoRem(bigU, bigV, wantR)
		if bigFromWords(q[:]).Cmp(wantQ) != 0 || bigFromWords(r[:]).Cmp(wantR) != 0 {
			t.Fatalf("mismatch i=%d dividend=%s divisor=%s quotient=%s want=%s remainder=%s want=%s", i, bigU, bigV, bigFromWords(q[:]), wantQ, bigFromWords(r[:]), wantR)
		}
	}
}
