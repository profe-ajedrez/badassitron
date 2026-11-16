package optimus

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

// SQLDecimal is an opt-in, exact, nullable database/sql representation of
// Optimus. Valid distinguishes SQL NULL from numeric zero.
//
// Scan accepts decimal text ([]byte or string) and integral values. It rejects
// binary floating-point inputs, since a driver returning a float may have
// already discarded digits from the original SQL NUMERIC/DECIMAL column.
// Values that would require Optimus to round are rejected rather than stored
// approximately. The destination is unchanged if Scan fails.
//
// The historical (*Optimus).Scan and (Optimus).Value contracts are unchanged.
// Drivers must return decimal data as text to guarantee exact conversion.
// The SQL column must also be able to store the complete decimal value.
//
// Example:
//
//	var price optimus.SQLDecimal
//	err := db.QueryRow("SELECT price FROM products WHERE id = ?", id).Scan(&price)
//	if err == nil && price.Valid {
//		_, err = db.Exec("UPDATE products SET price = ? WHERE id = ?", price, id)
//	}
type SQLDecimal struct {
	// Decimal stores the value when Valid is true.
	Decimal Optimus
	// Valid reports whether a value was present instead of SQL NULL.
	Valid bool
}

// Scan implements sql.Scanner without silent rounding or NULL-to-zero
// coercion. A failed scan leaves both Decimal and Valid unchanged.
func (d *SQLDecimal) Scan(src any) error {
	if d == nil {
		return fmt.Errorf("optimus: nil SQLDecimal destination")
	}
	if src == nil {
		*d = SQLDecimal{}
		return nil
	}

	var result Optimus
	switch v := src.(type) {
	case string:
		parsed, err := parseSQLDecimalExact(v)
		if err != nil {
			return err
		}
		result = parsed
	case []byte:
		parsed, err := parseSQLDecimalExact(string(v))
		if err != nil {
			return err
		}
		result = parsed
	case int64:
		result = FromInt64(v)
	case int:
		result = FromInt(v)
	default:
		return fmt.Errorf("optimus: exact SQL scan rejects %T; use a driver returning decimal text or an integer", src)
	}
	*d = SQLDecimal{Decimal: result, Valid: true}
	return nil
}

// Value implements driver.Valuer. SQL NULL maps to nil, while an invalid
// Optimus produces an error rather than the textual value "NaN".
func (d SQLDecimal) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	if d.Decimal.IsNaN() {
		return nil, d.Decimal.ErrorDetails()
	}
	return d.Decimal.String(), nil
}

func parseSQLDecimalExact(text string) (Optimus, error) {
	if text == "" {
		return Zero, fmt.Errorf("optimus: empty SQL decimal")
	}
	parsed := FromString(text)
	if parsed.IsNaN() {
		return Zero, fmt.Errorf("optimus: invalid SQL decimal %q: %w", text, parsed.ErrorDetails())
	}
	original, ok := new(big.Rat).SetString(text)
	if !ok {
		return Zero, fmt.Errorf("optimus: invalid SQL decimal %q", text)
	}
	coefficient := parsed.coefBig()
	if parsed.IsNegative() {
		coefficient.Neg(coefficient)
	}
	represented := new(big.Rat).SetFrac(coefficient, pow10(int(parsed.exp)))
	if represented.Cmp(original) != 0 {
		return Zero, fmt.Errorf("optimus: SQL decimal %q cannot be represented exactly", text)
	}
	return parsed, nil
}
