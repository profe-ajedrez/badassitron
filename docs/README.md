# Documentation

Start with a working example, then use the reference when you need exact
rounding, error, or capacity rules.

## Start here

| Goal | Guide |
| --- | --- |
| Make your first calculation | [Getting started](GETTING_STARTED.md) |
| Find code for a specific task | [Recipe book](RECIPES.md) |
| Run complete programs | [Example index](../examples/README.md) |
| Choose between Optimus and fixed widths | [Decimal types](DECIMAL_TYPES.md) |

## Everyday tasks

| Topic | Guide | Runnable example |
| --- | --- | --- |
| Parsing and formatting | [Optimus guide](OPTIMUS.md) | [Basics](../examples/01_optimus_basics/README.md) |
| Rounding and inexact results | [Rounding and errors](ROUNDING_AND_ERRORS.md) | [Rounding](../examples/02_rounding/README.md), [line vs. total](../examples/13_rounding_boundaries/README.md) |
| Decimal behavior and edge cases | [Behavior reference](DECIMAL_BEHAVIOR.md) | [Fixed-width Dec128](../examples/06_dec128/README.md) |
| JSON and SQL | [Data input/output](DATA_IO.md) | [JSON](../examples/04_json/README.md), [SQL](../examples/05_sql/README.md) |
| Exact width conversion | [Decimal types](DECIMAL_TYPES.md) | [Conversions](../examples/09_conversions/README.md) |
| Unsigned 128-bit arithmetic | [Integer representation](UNSIGNED_INTEGERS.md) | [Uint128](../examples/12_uint128/README.md) |
| Concurrent calculations | [Concurrency notes](../optimus/CONCURRENCY.md) | [Concurrency](../examples/10_concurrency/README.md) |
| Existing handler pipeline | [Pipeline and known limits](CALCULATION_PIPELINE.md) | [Pipeline](../examples/11_decimal_pipeline/README.md) |

## API reference

- [GoDoc and executable examples](GODOC_AND_EXAMPLES.md)
- [Dec128 precision and method details](../dec128/PRECISION.md)
- [Dec128 SQL scanning](../dec128/SQL_SCANNING.md)
- [Architecture](ARCHITECTURE.md)

From a local checkout, `go doc -all ./optimus` and `go doc -all ./dec128`
show exported API declarations. Hosted GoDoc is available only for module
versions indexed by pkg.go.dev.

## Maintainers

- [Release verification and outstanding requirements](RELEASE_STATUS.md)
- [Source provenance review](PROVENANCE_REVIEW.md) and [third-party notices](../THIRD_PARTY_NOTICES.md)
- [Verification commands](../tools/release/README.md)
- [Engineering history](MIGRATION_ARCHIVE.md) — dated experiments and superseded plans,
  not instructions for the current code

The source snapshot is **not yet approved for publication**. The numeric API
can be studied and tested separately from the still-open historical
provenance review.
