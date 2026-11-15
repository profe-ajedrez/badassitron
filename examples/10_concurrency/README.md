# Calculate concurrently with explicit scales

Run this from the repository root:

```sh
go run ./examples/10_concurrency
```

Read [main.go](main.go) for the complete program.

Three workers divide one by seven. Each worker chooses its own scale and
uses half even rounding. They read the same input values and write to
separate elements of a result slice. A WaitGroup ensures that all writes
finish before the main goroutine reads the results.

Optimus arithmetic returns new values and treats its inputs as immutable.
This allows the workers to share the decimal inputs. Sharing an input does
not make concurrent writes to the same result variable safe. Keep separate
destinations or protect shared writes with synchronization.

SetDefaultPrecision changes a process wide setting. Its atomic access avoids
a data race, but another goroutine can still change it between calculations.
Use DivRound and the other explicit scale methods when independent work
needs predictable rounding. Do not modify exported numeric globals or
lookup tables while other goroutines use them.

To execute this program with Go's race detector:

```sh
go run -race ./examples/10_concurrency
```

Results are printed after all workers finish, so their order is stable.

Expected output:

```text
Scale 2: 0.14
Scale 4: 0.1429
Scale 6: 0.142857
```
