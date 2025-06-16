package withfloat64

import "errors"

var (
	ErrNilArgument         = errors.New("se esperaban los argumentos de entrada config, inputy output, pero uno o más son nil ¿Esta seguro de estar invocando la cadena de responsabilidad en el orden correcto?")
	ErrNegativeUnitary     = errors.New("el unitario es negativo")
	ErrNegativeQty         = errors.New("la cantidad es negativa")
	ErrTaxOver100          = NewTaxError(errors.New("el impuesto porcentual es mayor a 100"), "")
	ErrNegativeTax         = NewTaxError(errors.New("se detecto un impuesto negativo. El valor del impuesto no puede ser negativo, ya sea porcentual o de monto"), "")
	ErrInvalidTaxType      = NewTaxError(errors.New("el impuesto se indica de un tipo invalido, debe ser: percentual, amount o amount_line"), "")
	ErrTaxStageOutOfBounds = NewTaxError(errors.New("tax stage of detail tax is out of defined limits"), "")
	ErrZeroQty             = errors.New("si quieres vender cantidad cero 0 ¿Que maiz paloma quieres vender?")
	ErrInvalidTaxStage     = errors.New("tax stage of detail tax is invalid")
)

type baseError struct {
	err error
	msg string
}

func (be *baseError) Error() string {
	if be.msg == "" {
		return be.err.Error()
	}
	return be.msg + " " + be.err.Error()
}

func (be *baseError) Unwrap() error {
	return be.err
}

type TaxError struct {
	baseError
}

func NewTaxError(err error, msg string) *TaxError {
	return &TaxError{
		baseError: baseError{
			err: err,
			msg: msg,
		},
	}
}

func (te *TaxError) Error() string {
	if te.msg == "" {
		return "tax error: " + te.err.Error()
	}
	return "tax error: " + te.msg + " " + te.err.Error()
}

func (te *TaxError) Unwrap() error {
	return te.err
}

type NaturalTaxError struct {
	TaxError
}

func NewNaturalTaxError(err error, msg string) *TaxError {
	return NewTaxError(err, "natural tax error: "+msg)
}

func (nte *NaturalTaxError) Error() string {
	if nte.msg == "" {
		return "natural tax error: " + nte.err.Error()
	}
	return "natural tax error: " + nte.msg + " " + nte.err.Error()
}

func (nte *NaturalTaxError) Unwrap() error {
	return nte.err
}

type OverTaxError struct {
	TaxError
}

func NewOverTaxError(err error, msg string) *TaxError {
	return NewTaxError(err, "natural tax error: "+msg)
}

func (nte *OverTaxError) Error() string {
	if nte.msg == "" {
		return "natural tax error: " + nte.err.Error()
	}
	return "natural tax error: " + nte.msg + " " + nte.err.Error()
}

func (nte *OverTaxError) Unwrap() error {
	return nte.err
}

type BypassTaxError struct {
	TaxError
}

func NewBypassTaxError(err error, msg string) *TaxError {
	return NewTaxError(err, "natural tax error: "+msg)
}

func (nte *BypassTaxError) Error() string {
	if nte.msg == "" {
		return "natural tax error: " + nte.err.Error()
	}
	return "natural tax error: " + nte.msg + " " + nte.err.Error()
}

func (nte *BypassTaxError) Unwrap() error {
	return nte.err
}

type DiscountError struct {
	baseError
}

func NewDiscountError(err error, msg string) *DiscountError {
	return &DiscountError{
		baseError: baseError{
			err: err,
			msg: msg,
		},
	}
}

func (de *DiscountError) Error() string {
	if de.msg == "" {
		return "discount error: " + de.err.Error()
	}
	return "discount error: " + de.msg + " " + de.err.Error()
}

func (de *DiscountError) Unwrap() error {
	return de.err
}
