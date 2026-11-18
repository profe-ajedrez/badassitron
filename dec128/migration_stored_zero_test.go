package dec128

import (
	"github.com/profe-ajedrez/badassitron/dec128/uint128"
	"testing"
)

func TestNewStoredZeroRetainsRawSignAtScaleZero(t *testing.T) {
	normalized := New(uint128.Zero, 0, true)
	retained := NewStoredZero(0, true)
	if normalized.StoredNegative() || !retained.StoredNegative() || !retained.IsZero() || retained.Precision() != 0 || retained.StringFixed() != "0" {
		t.Fatalf("raw-sign compatibility: normalized=%v stored=%v", normalized.StoredNegative(), retained.StoredNegative())
	}
	if !NewStoredZero(20, true).IsNaN() {
		t.Fatal("invalid stored-zero scale accepted")
	}
}
