# SQL scanning without breaking the `Dec128` method set

The historical `func (decimal Dec128) Scan(src any) error` has a value
receiver. It satisfies `database/sql.Scanner` but cannot update its caller:
`rows.Scan(&value)` calls that legacy method on a copy. This is a known API
limitation; switching to a pointer receiver would remove `Scan` from the
value method set and is intentionally not done in this compatible release.

To receive a column from `database/sql`, wrap the destination explicitly:

```go
var amount dec128.Dec128
err := rows.Scan(dec128.ScanDestination(&amount))
```

For direct conversion of database driver values:

```go
var amount dec128.Dec128
if err := dec128.ScanInto(&amount, []byte("123.456")); err != nil {
    // Handle invalid decimal or unsupported SQL driver type.
}
```

`ScanInto` accepts `nil` (sets zero), `string`, `[]byte`, `int`, `int64`,
`float32`, and `float64`. Inputs are parsed as decimals; binary floating-point
inputs cannot recover digits already lost in the source float. When parsing
fails, the previous destination value is preserved. A nil destination returns
an error. A SQL NULL is mapped to zero for compatibility with the legacy
implementation; if an application must distinguish NULL, use `sql.Null*` or
its own nullable wrapper.

**Compatibility warning:** do not assume `rows.Scan(&value)` updates `value`.
The old `Scan` signature remains callable and unchanged. Applications that
require the standard pointer-receiver `sql.Scanner` behavior directly on
`*Dec128` need a separately planned API-breaking release.
