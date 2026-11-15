package withfloat64

import "errors"

var (
	// ErrNilArgument reports a missing required argument.
	ErrNilArgument = errors.New("se esperaban los argumentos de entrada config, inputy output, pero uno o más son nil ¿Esta seguro de estar invocando la cadena de responsabilidad en el orden correcto?")
	// ErrNegativeUnitary reports a negative unit value.
	ErrNegativeUnitary = errors.New("el unitario es negativo")
	// ErrNegativeQty reports a negative quantity.
	ErrNegativeQty = errors.New("la cantidad es negativa")
	// ErrTaxOver100 reports a tax percentage over one hundred.
	ErrTaxOver100 = NewTaxError(errors.New("el impuesto porcentual es mayor a 100"), "")
	// ErrNegativeTax reports a negative tax value.
	ErrNegativeTax = NewTaxError(errors.New("se detecto un impuesto negativo. El valor del impuesto no puede ser negativo, ya sea porcentual o de monto"), "")
	// ErrInvalidTaxType reports an unknown tax type.
	ErrInvalidTaxType = NewTaxError(errors.New("el impuesto se indica de un tipo invalido, debe ser: percentual, amount o amount_line"), "")
	// ErrTaxStageOutOfBounds reports a tax stage outside the supported range.
	ErrTaxStageOutOfBounds = NewTaxError(errors.New("tax stage of detail tax is out of defined limits"), "")
	// ErrZeroQty reports zero quantity.
	ErrZeroQty = errors.New("si quieres vender cantidad cero 0 ¿Que maiz paloma quieres vender?")
	// ErrInvalidTaxStage reports an invalid tax stage.
	ErrInvalidTaxStage = errors.New("tax stage of detail tax is invalid")
)

type baseError struct {
	err error
	msg string
}

// Error returns the description of this error.
func (be *baseError) Error() string {
	if be.msg == "" {
		return be.err.Error()
	}
	return be.msg + " " + be.err.Error()
}

// Unwrap returns the error wrapped by this error value.
func (be *baseError) Unwrap() error {
	return be.err
}

// TaxError wraps an error raised while processing taxes.
type TaxError struct {
	baseError
}

// NewTaxError creates a typed error with the given cause and message.
func NewTaxError(err error, msg string) *TaxError {
	return &TaxError{
		baseError: baseError{
			err: err,
			msg: msg,
		},
	}
}

// Error returns the description of this error.
func (te *TaxError) Error() string {
	if te.msg == "" {
		return "tax error: " + te.err.Error()
	}
	return "tax error: " + te.msg + " " + te.err.Error()
}

// Unwrap returns the error wrapped by this error value.
func (te *TaxError) Unwrap() error {
	return te.err
}

// NaturalTaxError wraps an error from the natural tax stage.
type NaturalTaxError struct {
	TaxError
}

// NewNaturalTaxError creates a typed error with the given cause and message.
func NewNaturalTaxError(err error, msg string) *TaxError {
	return NewTaxError(err, "natural tax error: "+msg)
}

// Error returns the description of this error.
func (nte *NaturalTaxError) Error() string {
	if nte.msg == "" {
		return "natural tax error: " + nte.err.Error()
	}
	return "natural tax error: " + nte.msg + " " + nte.err.Error()
}

// Unwrap returns the error wrapped by this error value.
func (nte *NaturalTaxError) Unwrap() error {
	return nte.err
}

// OverTaxError wraps an error from the overtax stage.
type OverTaxError struct {
	TaxError
}

// NewOverTaxError creates a typed error with the given cause and message.
func NewOverTaxError(err error, msg string) *TaxError {
	return NewTaxError(err, "natural tax error: "+msg)
}

// Error returns the description of this error.
func (nte *OverTaxError) Error() string {
	if nte.msg == "" {
		return "natural tax error: " + nte.err.Error()
	}
	return "natural tax error: " + nte.msg + " " + nte.err.Error()
}

// Unwrap returns the error wrapped by this error value.
func (nte *OverTaxError) Unwrap() error {
	return nte.err
}

// BypassTaxError wraps an error from the bypass tax stage.
type BypassTaxError struct {
	TaxError
}

// NewBypassTaxError creates a typed error with the given cause and message.
func NewBypassTaxError(err error, msg string) *TaxError {
	return NewTaxError(err, "natural tax error: "+msg)
}

// Error returns the description of this error.
func (nte *BypassTaxError) Error() string {
	if nte.msg == "" {
		return "natural tax error: " + nte.err.Error()
	}
	return "natural tax error: " + nte.msg + " " + nte.err.Error()
}

// Unwrap returns the error wrapped by this error value.
func (nte *BypassTaxError) Unwrap() error {
	return nte.err
}

// DiscountError wraps an error raised while processing a discount.
type DiscountError struct {
	baseError
}

// NewDiscountError creates a typed error with the given cause and message.
func NewDiscountError(err error, msg string) *DiscountError {
	return &DiscountError{
		baseError: baseError{
			err: err,
			msg: msg,
		},
	}
}

// Error returns the description of this error.
func (de *DiscountError) Error() string {
	if de.msg == "" {
		return "discount error: " + de.err.Error()
	}
	return "discount error: " + de.msg + " " + de.err.Error()
}

// Unwrap returns the error wrapped by this error value.
func (de *DiscountError) Unwrap() error {
	return de.err
}
