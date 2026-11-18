package uint128

import ju "github.com/jokruger/dec128/uint128"

// StringToBuf formats using the imported Jokruger backend; its zero-value
// behavior and caller-owned buffer contract match the historical API.
func (u Uint128) StringToBuf(buf []byte) []byte {
	return ju.Uint128(u).StringToBuf(buf)
}
