# Decimal architecture and safe use

## Representation

Each decimal stores a signed base 10 value through a nonnegative integer
coefficient and a fractional scale. The 128-, 256-, and 512-bit decimal
packages use fixed coefficient capacities. Optimus stores small values
in smaller coefficient tiers and can promote as needed.

Scale is bounded even in the 512-bit type. A value may be representable
with a large coefficient and low scale, or a small coefficient and a
high scale. Neither automatically implies economic exactness for future
operations.

## Arithmetic and error boundaries

Arithmetic aims to retain exact results when they fit. Different methods
have distinct fallback and rounding contracts. Invalid input, a division
by zero, or an unrepresentable result may produce NaN. Checked parsers
and exact narrowing conversions return errors instead.

When output precision matters, choose explicit `DivRound`, `MulRound`,
`RescaleRound`, or `SqrtRound` and verify the inexact flag where offered.
Do not assume `Round` and `RoundBank` are interchangeable.

## Value ownership and concurrency

Optimus values are immutable by convention. Wide coefficients may be
indirect but are not mutated by ordinary arithmetic; independent
operations may read the same source concurrently. Other destination
values, such as those written by `Scan` and `UnmarshalJSON`, must not
be modified concurrently without synchronization.

Default precision is a process-wide atomic setting. This prevents
unsynchronized storage access, but it does not ensure that a group of
arithmetic operations observes the same precision if another worker
changes it. Use explicit precision to avoid this ambiguity.

See [Optimus concurrency](../optimus/CONCURRENCY.md) and
[rounding and errors](ROUNDING_AND_ERRORS.md).

## Capacity and cost

Operations with wide coefficients can require more memory and CPU work.
Do not infer a zero-allocation guarantee from a microbenchmark. A useful
performance comparison records the exact type, input values, scale,
operation, Go version, CPU, and allocation measurement method.

The adaptive type is a convenience and a memory optimization for small
values, not an unlimited precision type. Overflow and precision limits
still apply.

## Legacy calculation workflows

The `withdec128` and `withfloat64` packages are separate, older handler
pipelines. Their validation, tax stages, and output ownership need
application-level scrutiny. They do not supply an automatically correct
jurisdictional calculation engine. The decimal packages are usable
without those pipelines.

See [calculation pipeline](CALCULATION_PIPELINE.md) and the
[complete example](../examples/11_decimal_pipeline/README.md).
