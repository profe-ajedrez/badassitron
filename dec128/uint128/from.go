package uint128

import (
	"encoding/binary"
	"math/big"

	"github.com/profe-ajedrez/badassitron/dec128/errors"
)

// FromUint64 creates a new Uint128 from a uint64
func FromUint64(u uint64) Uint128 {
	return Uint128{u, 0}
}

// FromBytes creates a new Uint128 from a [16]byte
func FromBytes(bs [16]byte) Uint128 {
	return Uint128{binary.LittleEndian.Uint64(bs[:8]), binary.LittleEndian.Uint64(bs[8:])}
}

// FromBytesBigEndian creates a new Uint128 from a [16]byte in big endian
func FromBytesBigEndian(b [16]byte) Uint128 {
	return Uint128{binary.BigEndian.Uint64(b[8:]), binary.BigEndian.Uint64(b[:8])}
}

// FromBigInt creates a new Uint128 from a *big.Int
func FromBigInt(i *big.Int) (Uint128, errors.Error) {
	if i.Sign() < 0 {
		return Zero, errors.Negative
	}

	if i.BitLen() > 128 {
		return Zero, errors.Overflow
	}

	return Uint128{i.Uint64(), i.Rsh(i, 64).Uint64()}, errors.None
}
