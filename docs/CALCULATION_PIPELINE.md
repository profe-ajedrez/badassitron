# Existing calculation workflow

The repository has an older calculation pipeline in `withdec128` and its
`handler` subpackage. This is separate from the Optimus numeric API.

It uses `Enterable` for input, `Outputable` for output, and
`CalculationConfiger` for options. A handler receives those three interfaces
and the remaining handlers. `handler.Next` calls the next handler in order.

## Read the current handler flow

The existing test uses these handlers in this order.

```go
err := handler.Next(
    opts, input, output,
    handler.EntryValidation,
    handler.Bootstrap,
    handler.Netter,
    handler.Taxer,
    handler.Grosser,
)
```

The handlers have distinct responsibilities.

1. `EntryValidation` checks required values, quantity, and discount bounds.
2. `Bootstrap` copies the starting values to output and can normalize the unit value.
3. `Netter` calculates net values and the applied discount.
4. `Taxer` computes taxes from configured tax stages.
5. `Grosser` adds the tax to net values and calculates gross values.

The order matters. `Taxer` needs net values produced by `Netter`.

## Limits in the current implementation

This pipeline is a historical implementation. It should not be treated as a
complete general rule engine. Some methods and interfaces are not fully
implemented. In particular, `withdec128.Output.WithTaxes` currently panics.
The package tests use a custom output implementation for some workflows.

The current source also contains a parallel `withfloat64` implementation,
but that package has existing compilation errors. The Optimus v8 audit does
not validate or repair either workflow architecture.

For new exact arithmetic code, start with `optimus` or a fixed width decimal
type. Keep discount, tax, and document rules in the calling application.

See [the README](../README.md) for the numeric package entry points and
[the v8 audit](../OPTIMUS_V8_PHASE_C_PRECISION_AUDIT.md) for verified scope.
