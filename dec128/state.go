package dec128

import "github.com/profe-ajedrez/badassitron/dec128/errors"

// Each Dec128 is a four-scalar value for Go's amd64 register ABI. The
// low state bit preserves the sign (including signed zero), and the remaining
// bits preserve the full uint8 error code. Use uint16 to avoid truncating
// codes >= 128. No exported API changes.
func signState(neg bool) uint16 {
	if neg {
		return 1
	}
	return 0
}
func errorState(e errors.Error) uint16              { return uint16(e) << 1 }
func combinedState(neg bool, e errors.Error) uint16 { return errorState(e) | signState(neg) }
func (d Dec128) negative() bool                     { return d.state&1 != 0 }
func (d Dec128) errorCode() errors.Error            { return errors.Error(d.state >> 1) }
func (d Dec128) hasError() bool                     { return d.state >= 2 }
