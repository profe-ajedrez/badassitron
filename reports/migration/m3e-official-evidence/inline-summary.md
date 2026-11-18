# M3 compiler inlining diagnostics

Extracted from compiler output; `not observed` is not a PASS or failure.

## Historical facade

- `Uint128.Add`: can inline Uint128.Add with cost 40
- `Uint128.Mul`: cannot inline Uint128.Mul: function too complex: cost 91 exceeds budget 80
- `Uint128.Mul64`: can inline Uint128.Mul64 with cost 52
- `Uint128.MulCarry`: cannot inline Uint128.MulCarry: function too complex: cost 156 exceeds budget 80

## M3 facade

- `Uint128.Add`: can inline Uint128.Add with cost 60
- `Uint128.Mul`: can inline Uint128.Mul with cost 77
- `Uint128.Mul64`: can inline Uint128.Mul64 with cost 72
- `Uint128.MulCarry`: cannot inline Uint128.MulCarry: function too complex: cost 110 exceeds budget 80

## Jokruger direct

- `Uint128.Add`: can inline Uint128.Add with cost 40
- `Uint128.Mul`: cannot inline Uint128.Mul: function too complex: cost 95 exceeds budget 80
- `Uint128.Mul64`: can inline Uint128.Mul64 with cost 52
- `Uint128.MulCarry`: cannot inline Uint128.MulCarry: function too complex: cost 156 exceeds budget 80

