package internal

import (
	e "errors"
	"runtime"
	"strconv"
	str "strings"
)

var (
	ErrNegativeValue          error = e.New("value is negative")
	ErrZeroDivision           error = e.New("dividing by zero")
	ErrCantUndiscountFromZero       = e.New("cant calculate the value without discount from zero when there are discounts presents")
)

var ()

type WrappingError struct {
	Inner      error
	Message    string
	Additional string
}

func (e *WrappingError) Error() string {
	sb := GetSB()
	defer PutSB(sb)

	sb.WriteString(e.Inner.Error())
	sb.WriteString(" --- ")
	sb.WriteString("error caused due to ")
	sb.WriteString(e.Message)
	sb.WriteString(" --- ")
	sb.WriteString("; additional context: ")
	sb.WriteString(e.Additional)

	return sb.String()
}

func (e *WrappingError) Unwrap() error {
	// Return the inner error.
	return e.Inner
}

func WrapWithWrappingError(err error, message string) *WrappingError {
	additional := trace(2, 3)
	add := ""

	if len(additional) > 1 {
		add = str.Join(additional, " | ")
	} else if len(additional) == 1 {
		add = additional[0]
	}

	return &WrappingError{
		Inner:      err,
		Message:    message,
		Additional: add,
	}
}

// trace
// Construye el slice de frames asociados al [TracerError], aplicando el formato
// según el [FrameFormatter] asociado.
func trace(startLevel, stopLevel uint) []string {

	if stopLevel < startLevel {
		stopLevel = startLevel + 8
	}

	if stopLevel == startLevel {
		stopLevel += 1
	}

	frames := make([]string, stopLevel-startLevel)
	k := 0
	t := int(startLevel)
	b := int(stopLevel)
	sl := int(startLevel)

	for i := t; i < b; i++ {
		_, filename, line, ok := runtime.Caller(i)

		if ok && filename != "" {
			frames[k] = filename + " " + strconv.Itoa(line) + " [callstack level] " + strconv.Itoa(i-sl)
			//formatter.FormatLine(fileName, line, i-sl)
			k += 1
		}
	}

	frames = frames[0:k]

	return frames
}
