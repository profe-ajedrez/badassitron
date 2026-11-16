// Package uint128 provides 128-bit unsigned integer type and basic operations.
package uint128

import (
	"fmt"
	"math/big"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
)

// Uint128 is a 128-bit unsigned integer type.
type Uint128 struct {
	Lo uint64
	Hi uint64
}

// IsZero returns true if the value is zero.
func (uint128 Uint128) IsZero() bool {
	return uint128.Lo == 0 && uint128.Hi == 0
}

// Equal returns true if the value is equal to the other value.
func (uint128 Uint128) Equal(other Uint128) bool {
	return uint128.Lo == other.Lo && uint128.Hi == other.Hi
}

// Compare returns -1 if the value is less than the other value, 0 if the value is equal to the other value, and 1 if the value is greater than the other value.
func (uint128 Uint128) Compare(other Uint128) int {
	if uint128 == other {
		return 0
	}

	if uint128.Hi < other.Hi || (uint128.Hi == other.Hi && uint128.Lo < other.Lo) {
		return -1
	}

	return 1
}

// BitLen returns the number of bits required to represent the value.
func (uint128 Uint128) BitLen() int {
	return 128 - uint128.LeadingZeroBitsCount()
}

// Scan scans the value.
func (uint128 *Uint128) Scan(s fmt.ScanState, ch rune) error {
	i := new(big.Int)

	if err := i.Scan(s, ch); err != nil {
		return errors.InvalidFormat.Value()
	}

	if i.Sign() < 0 {
		return errors.Negative.Value()
	}

	if i.BitLen() > 128 {
		return errors.Overflow.Value()
	}

	uint128.Lo = i.Uint64()
	uint128.Hi = i.Rsh(i, 64).Uint64()

	return nil
}
