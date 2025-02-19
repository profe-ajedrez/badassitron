package uint128

import "fmt"

// MarshalText implements the encoding.TextMarshaler interface.
func (uint128 Uint128) MarshalText() ([]byte, error) {
	return []byte(uint128.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (uint128 *Uint128) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*uint128 = Zero
		return nil
	}

	_, err := fmt.Sscan(string(b), uint128)
	return err
}
