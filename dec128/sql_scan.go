package dec128

import (
	"database/sql"
	"fmt"
	"strconv"
)

// ScanInto decodes a database/sql driver value into dst. Unlike the legacy
// Dec128.Scan value-receiver method, it writes to the destination on success.
// Invalid inputs do not change dst. It accepts the SQL driver types handled
// by Scan, plus []byte (commonly returned by database drivers).
func ScanInto(dst *Dec128, src any) error {
	if dst == nil {
		return fmt.Errorf("dec128: nil scan destination")
	}

	var parsed Dec128
	switch v := src.(type) {
	case nil:
		parsed = Zero
	case string:
		parsed = FromString(v)
	case []byte:
		parsed = FromString(v)
	case int:
		parsed = FromInt(v)
	case int64:
		parsed = FromInt64(v)
	case float32:
		// Use the shortest round-trip decimal for this float32. The original
		// Scan uses fmt %%f, silently dropping digits beyond six places.
		parsed = FromString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	case float64:
		parsed = FromFloat64(v)
	default:
		return fmt.Errorf("can't scan %T to Dec128: %T is not supported", src, src)
	}
	if parsed.IsNaN() {
		return parsed.ErrorDetails()
	}
	*dst = parsed
	return nil
}

type sqlScanDestination struct {
	dst *Dec128
}

// Scan implements the historical Dec128 scanning behavior. Refer to SQL_SCANNING.md for the pointer receiver limitation.
func (s sqlScanDestination) Scan(src any) error {
	return ScanInto(s.dst, src)
}

// ScanDestination wraps dst as a functional database/sql.Scanner. Pass it to
// rows.Scan instead of &dst, because the historical Dec128.Scan method has a
// value receiver and cannot modify the variable it was invoked on.
//
// Example:
//
//	err := rows.Scan(dec128.ScanDestination(&amount))
func ScanDestination(dst *Dec128) sql.Scanner {
	return sqlScanDestination{dst: dst}
}
