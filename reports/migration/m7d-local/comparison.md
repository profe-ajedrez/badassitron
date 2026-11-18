# Local Add experiment (Go 1.23.2; NOT official)

Three samples of 120ms from separately compiled M7-C and M7-D source trees. No local figures should be extrapolated to Go 1.27.1.

| Build | Operation | Facade ns/op | Direct Jokruger ns/op | Difference |
|---|---|---:|---:|---:|
| M7-C | `Add` | 1.760 | 1.643 | +7.1% |
| M7-C | `AddOverflow` | 1.751 | 1.503 | +16.5% |
| M7-D | `Add` | 1.251 | 1.679 | -25.5% |
| M7-D | `AddOverflow` | 1.308 | 1.611 | -18.8% |

The host is noisy; medians do not establish statistical equivalence. Both approaches have 0 B/op, 0 allocs/op.
