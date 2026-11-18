package uint128

import (
	"math/big"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
)

// Uint64 returns the value as uint64 if it fits, otherwise it returns an error.
func (uint128 Uint128) Uint64() (uint64, errors.Error) {
	if uint128.Hi != 0 {
		return 0, errors.Overflow
	}
	return uint128.Lo, errors.None
}

// Bytes returns the value as a [16]byte array.
func (uint128 Uint128) Bytes() [16]byte {
	bs := [16]byte{}
	_ = uint128.PutBytes(bs[:])
	return bs
}

// BytesBigEndian returns the value as a [16]byte array in big-endian order.
func (uint128 Uint128) BytesBigEndian() [16]byte {
	bs := [16]byte{}
	_ = uint128.PutBytesBigEndian(bs[:])
	return bs
}

// BigInt returns the value as a big.Int.
func (uint128 Uint128) BigInt() *big.Int {
	i := new(big.Int).SetUint64(uint128.Hi)
	i = i.Lsh(i, 64)
	i = i.Xor(i, new(big.Int).SetUint64(uint128.Lo))
	return i
}

// String returns the value as a string.
func (uint128 Uint128) String() string {
	if uint128.IsZero() {
		return ZeroStr
	}

	buf := [MaxStrLen]byte{}
	sb := uint128.StringToBuf(buf[:])

	return string(sb)
}
