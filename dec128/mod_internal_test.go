package dec128

import "testing"

// The sign of a zero remainder is not exposed by String(), but preserving
// the original bit is important for the existing Dec128 internal contract.
func TestModPreservesLegacyInternalSignedZero(t *testing.T) {
	a := FromString("-2")
	b := FromString("1")
	_, legacyRemainder := a.QuoRem(b)
	actual := a.Mod(b)
	if !legacyRemainder.IsZero() || !actual.IsZero() || legacyRemainder.StoredNegative() != actual.StoredNegative() {
		t.Fatalf("Mod changed remainder sign: QuoRem=%#v Mod=%#v", legacyRemainder, actual)
	}
}
