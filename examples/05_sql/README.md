# Convert SQL decimal values

Run this from the repository root:

```sh
go run ./examples/05_sql
```

Read [main.go](main.go) for the complete program. It calls the database
conversion methods directly, so it needs no server or database driver.

## Exact nullable values with Optimus

SQLDecimal implements sql.Scanner and driver.Valuer. Its Valid field is
false for SQL NULL and true for a numeric value, including zero. Read the
Decimal field only when Valid is true.

Scan accepts decimal text, bytes, and supported integers. It rejects binary
floating point input and text that cannot be represented exactly. A failed
scan leaves the previous value unchanged. Value returns decimal text for a
present number and nil for SQL NULL. It returns an error for a present NaN.
Value preserves the numeric value but removes trailing fractional zeros.

With an existing database connection, the same type can be used like this:

```go
var amount optimus.SQLDecimal
err := db.QueryRow("SELECT amount FROM invoices WHERE id = ?", id).Scan(&amount)
if err != nil {
    return err
}
if amount.Valid {
    fmt.Println(amount.Decimal.StringFixed())
}
_, err = db.Exec("UPDATE invoices SET amount = ? WHERE id = ?", amount, id)
return err
```

This fragment belongs inside a function with a database connection, an id,
and an error return. Adjust placeholders for your driver. The driver must
return exact decimal text or an integer, and the database column must have
enough capacity to store the value. This local example exercises conversion;
it does not test a particular database or driver.

## Receive values with Dec128

The older Dec128.Scan method has a value receiver and cannot update the
caller's variable. Use the adapter with database/sql:

```go
var amount dec128.Dec128
err := rows.Scan(dec128.ScanDestination(&amount))
```

For a driver value you already have, use dec128.ScanInto(&amount, source).
These Dec128 helpers map SQL NULL to zero and also accept floats. Use text
when exact decimal input matters. They do not have SQLDecimal's strict
nullable contract.

Expected output:

```text
Present: true
Scanned: 123.4500
Value for SQL: 123.45
Float rejected: true
Previous value: 123.4500
Null is present: false
Null writes nil: true
Zero is present: true
Dec128 adapter: 42.500
Dec128 direct scan: 17.25
```
