package optimus

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"math/big"
	"strings"
	"testing"
)

type decimalTestConnector struct {
	source  driver.Value
	written driver.Value
}

func (c *decimalTestConnector) Connect(context.Context) (driver.Conn, error) {
	return &decimalTestConn{c}, nil
}
func (c *decimalTestConnector) Driver() driver.Driver { return decimalTestDriver{} }

type decimalTestDriver struct{}

func (decimalTestDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type decimalTestConn struct{ c *decimalTestConnector }

func (*decimalTestConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not needed") }
func (*decimalTestConn) Close() error                        { return nil }
func (*decimalTestConn) Begin() (driver.Tx, error)           { return nil, errors.New("not needed") }
func (c *decimalTestConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &decimalTestRows{value: c.c.source}, nil
}
func (c *decimalTestConn) ExecContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Result, error) {
	if len(args) != 1 {
		return nil, errors.New("expected one parameter")
	}
	c.c.written = args[0].Value
	return driver.RowsAffected(1), nil
}

type decimalTestRows struct {
	value   driver.Value
	emitted bool
}

func (*decimalTestRows) Columns() []string { return []string{"amount"} }
func (*decimalTestRows) Close() error      { return nil }
func (r *decimalTestRows) Next(dest []driver.Value) error {
	if r.emitted {
		return io.EOF
	}
	r.emitted = true
	dest[0] = r.value
	return nil
}
func TestPhaseCSQLDatabaseRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		src   driver.Value
		valid bool
	}{
		{"text", "12345678901234567890.12345678901234567890", true},
		{"bytes", []byte("-0.000000000000000000000000001"), true},
		{"int64", int64(-9223372036854775808), true},

		{"overflow", strings.Repeat("9", 200), false},
		{"too_precise", "0." + strings.Repeat("0", 154) + "1", false},
		{"float", float64(0.1), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &decimalTestConnector{source: tc.src}
			db := sql.OpenDB(c)
			defer db.Close()
			original := SQLDecimal{Decimal: FromInt(7), Valid: true}
			err := db.QueryRowContext(context.Background(), "SELECT amount").Scan(&original)
			if (err == nil) != tc.valid {
				t.Fatalf("Scan error %v; want success %v", err, tc.valid)
			}
			if !tc.valid {
				if !original.Valid || !original.Decimal.Equal(FromInt(7)) {
					t.Fatal("failure changed destination")
				}
				return
			}
			if _, err := db.ExecContext(context.Background(), "INSERT INTO amounts VALUES (?)", original); err != nil {
				t.Fatal(err)
			}
			text, ok := c.written.(string)
			if !ok {
				t.Fatalf("written value is %T", c.written)
			}
			if expected, ok := new(big.Rat).SetString(asDecimalString(tc.src)); !ok {
				t.Fatalf("bad fixture %v", tc.src)
			} else {
				actual, ok := new(big.Rat).SetString(text)
				if !ok || actual.Cmp(expected) != 0 {
					t.Fatalf("roundtrip changed value %s => %s", asDecimalString(tc.src), text)
				}
			}
		})
	}
}
func asDecimalString(v driver.Value) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case int64:
		return big.NewInt(x).String()
	}
	return ""
}
func TestPhaseCSQLNullAndNaN(t *testing.T) {
	c := &decimalTestConnector{source: nil}
	db := sql.OpenDB(c)
	defer db.Close()
	d := SQLDecimal{Decimal: FromInt(17), Valid: true}
	if err := db.QueryRow("SELECT amount").Scan(&d); err != nil || d.Valid || !d.Decimal.IsZero() {
		t.Fatalf("null scan: %+v, %v", d, err)
	}
	if _, err := db.Exec("INSERT INTO amounts VALUES (?)", d); err != nil || c.written != nil {
		t.Fatalf("null value: %v, %v", c.written, err)
	}
	bad := SQLDecimal{Decimal: NaN(0), Valid: true}
	if _, err := db.Exec("INSERT INTO amounts VALUES (?)", bad); err == nil {
		t.Fatal("NaN accepted for insert")
	}
}
