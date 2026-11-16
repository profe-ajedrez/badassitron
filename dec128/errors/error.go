// Package errors provides custom error type and error codes for uint128 and dec128 packages.
package errors

import "errors"

// Error is the decimal arithmetic error code type.
type Error uint8

const (
	// None identifies no error.
	None Error = iota
	// NotANumber identifies an invalid decimal.
	NotANumber
	// DivisionByZero identifies division by zero.
	DivisionByZero
	// Overflow identifies overflow of a representable coefficient.
	Overflow
	// Underflow identifies unsigned integer underflow.
	Underflow
	// Negative identifies an unexpected negative operand.
	Negative
	// NotEnoughBytes identifies insufficient binary input.
	NotEnoughBytes
	// InvalidFormat identifies invalid input format.
	InvalidFormat
	// PrecisionOutOfRange identifies a decimal scale outside supported bounds.
	PrecisionOutOfRange
	// RescaleToLessPrecision identifies rescaling to a less precise representation.
	RescaleToLessPrecision
	// SqrtNegative identifies the square root of a negative number.
	SqrtNegative
)

var code2err = [...]error{
	None:                   nil,
	NotANumber:             errors.New("not a number"),
	DivisionByZero:         errors.New("division by zero"),
	Overflow:               errors.New("overflow"),
	Underflow:              errors.New("underflow"),
	Negative:               errors.New("negative value in unsigned operation"),
	NotEnoughBytes:         errors.New("not enough bytes"),
	InvalidFormat:          errors.New("invalid format"),
	PrecisionOutOfRange:    errors.New("precision out of range"),
	RescaleToLessPrecision: errors.New("rescale to less precision"),
	SqrtNegative:           errors.New("square root of negative number"),
}

// Value returns this arithmetic error as a Go error.
func (e Error) Value() error {
	return code2err[e]
}

// Error returns the textual description of this arithmetic error.
// Do not call Error on None, because None has no underlying error value.
func (e Error) Error() string {
	return e.Value().Error()
}
