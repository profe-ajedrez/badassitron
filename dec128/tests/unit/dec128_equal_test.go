package unit

import (
	"testing"

	"github.com/profe-ajedrez/badassitron/dec128"
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
)

func TestDecimalEqual(t *testing.T) {
	dec128.SetDefaultPrecision(19)

	var a, b dec128.Dec128

	a = dec128.FromString("NaN")
	b = dec128.FromString("NaN")
	if !a.Equal(b) {
		t.Errorf("expected true, got false")
	}

	a = dec128.FromString("0")
	b = dec128.FromString("NaN")
	if a.Equal(b) {
		t.Errorf("expected false, got true")
	}

	a = dec128.FromString("0")
	b = dec128.FromString("0")
	if !a.Equal(b) {
		t.Errorf("expected true, got false")
	}

	a = dec128.FromString("1")
	b = dec128.FromString("-1")
	if a.Equal(b) {
		t.Errorf("expected false, got true")
	}

	a = dec128.New(uint128.FromUint64(1000), 1, false)
	b = dec128.New(uint128.FromUint64(10000), 2, false)
	if !a.Equal(b) {
		t.Errorf("expected true, got false")
	}

	a = dec128.New(uint128.FromUint64(123456), 3, false)
	b = dec128.New(uint128.FromUint64(123456000), 6, false)
	if !a.Equal(b) {
		t.Errorf("expected true, got false")
	}

	a = dec128.FromString("123.456")
	b = dec128.FromString("123.4560000")
	if !a.Equal(b) {
		t.Errorf("expected true, got false")
	}
}
