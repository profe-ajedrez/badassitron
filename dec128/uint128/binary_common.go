package uint128

import (
	"encoding/binary"
	"github.com/profe-ajedrez/badassitron/dec128/errors"
	"math/bits"
)

// PutBytes writes the Uint128 to the byte slice bs in little-endian order.
func (uint128 Uint128) PutBytes(bs []byte) errors.Error {
	if len(bs) < 16 {
		return errors.NotEnoughBytes
	}

	binary.LittleEndian.PutUint64(bs[:8], uint128.Lo)
	binary.LittleEndian.PutUint64(bs[8:], uint128.Hi)

	return errors.None
}

// PutBytesBigEndian writes the Uint128 to the byte slice bs in big-endian order.
func (uint128 Uint128) PutBytesBigEndian(bs []byte) errors.Error {
	if len(bs) < 16 {
		return errors.NotEnoughBytes
	}

	binary.BigEndian.PutUint64(bs[:8], uint128.Hi)
	binary.BigEndian.PutUint64(bs[8:], uint128.Lo)

	return errors.None
}

// AppendBytes appends the Uint128 to the byte slice bs in little-endian order.
func (uint128 Uint128) AppendBytes(bs []byte) []byte {
	bs = binary.LittleEndian.AppendUint64(bs, uint128.Lo)
	bs = binary.LittleEndian.AppendUint64(bs, uint128.Hi)
	return bs
}

// AppendBytesBigEndian appends the Uint128 to the byte slice bs in big-endian order.
func (uint128 Uint128) AppendBytesBigEndian(bs []byte) []byte {
	bs = binary.BigEndian.AppendUint64(bs, uint128.Hi)
	bs = binary.BigEndian.AppendUint64(bs, uint128.Lo)
	return bs
}

// ReverseBytes returns the Uint128 with the byte order reversed.
func (uint128 Uint128) ReverseBytes() Uint128 {
	return Uint128{bits.ReverseBytes64(uint128.Hi), bits.ReverseBytes64(uint128.Lo)}
}
