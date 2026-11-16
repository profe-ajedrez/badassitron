# JSON, SQL, and decimal conversions

## Parse exact decimal text

Prefer `NewFromString` for values coming from a user, API, or database.
Passing a `float64` value may already have lost decimal information
before the decimal constructor sees it.

`String()` prints a normalized decimal and `StringFixed()` retains the
stored number of fractional digits. These are display choices; they do
not define an economic rounding rule.

## JSON

The decimal packages provide JSON marshaling and unmarshaling.
Exact decimal text in quotes is a reliable way to preserve a source
value. Do not assume that a JSON number has preserved every digit if
it was first decoded into `float64` by another system.

Use the [JSON example](../examples/04_json/README.md) for the concrete
transport behavior and accepted inputs. A consuming API may impose
stricter JSON rules than the numeric packages themselves.

## SQL and NULL

`optimus.SQLDecimal` implements nullable SQL decimal handling. It
separates a SQL NULL from numeric zero. Its scanner rejects `float64`
input to avoid silently accepting a rounded value:

```go
var amount optimus.SQLDecimal
if err := amount.Scan("123.4500"); err != nil {
    return err
}
if amount.Valid {
    fmt.Println(amount.Decimal.StringFixed())
}
```

The fragment above assumes a surrounding function returning `error`.
See the executable [`ExampleSQLDecimal`](../optimus/example_doc01_test.go).

The older `Optimus.Scan` keeps inherited compatibility semantics, such as
mapping NULL to zero and accepting floating point. When the distinction
between NULL and zero matters, prefer the nullable type.

Driver support and column types still matter. If the database provides
rounded floating point text or numbers, the decimal library cannot
recover the original digits. See [SQL example](../examples/05_sql/README.md)
and [Dec128 SQL scanning](../dec128/SQL_SCANNING.md).

## Convert between decimal types

Widening to Optimus uses `optimus.FromDec128`, `FromDec256`, or
`FromDec512`. Narrowing uses `Dec128Exact`, `Dec256Exact`, or
`Dec512Exact` and returns an error when a precise representation
cannot be found. Handle that error explicitly.

See [`ExampleOptimus_Dec128Exact`](../optimus/example_doc01_test.go)
and [conversions](../examples/09_conversions/README.md).
