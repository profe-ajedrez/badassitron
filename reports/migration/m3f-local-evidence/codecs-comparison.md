# Imported unsigned codec comparison

Median ns/op; negative percentages mean faster. Each build also measures direct Jokruger as a same-build control.

| Operation | Legacy facade | Direct Jokruger (legacy build) | Imported facade | Direct Jokruger (imported build) | Imported vs legacy | Imported vs direct |
|---|---:|---:|---:|---:|---:|---:|
| `ParseLong` | 105.900 | 70.620 | 70.515 | 80.590 | -33.4% | -12.5% |
| `ParseShort` | 11.855 | 11.515 | 20.535 | 21.235 | +73.2% | -3.3% |
| `StringToBufWide` | 89.550 | 112.755 | 79.590 | 109.960 | -11.1% | -27.6% |

Do not extrapolate a single representative input to every length or CPU. Verify error codes separately.
Allocation figures (bytes/op, allocs/op):

- `ParseLong`: historical (0, 0); imported (0, 0); upstream (0, 0)
- `ParseShort`: historical (0, 0); imported (0, 0); upstream (0, 0)
- `StringToBufWide`: historical (0, 0); imported (0, 0); upstream (0, 0)
