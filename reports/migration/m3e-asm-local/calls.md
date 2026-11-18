# M3-E call observations (diagnostic, not causal proof)

## adapter-QuoRem256By128.asm.txt
```text
  m3_adapter.go:150	0x53b6ee		e84df8ffff		CALL github.com/jokruger/dec128/uint128.QuoRem256By128(SB)			
  m3_adapter.go:149	0x53b777		e8846bf3ff		CALL runtime.morestack_noctxt.abi0(SB)						
```
## adapter-Uint128-Mul.asm.txt
```text
NOT OBSERVED: symbol absent (potentially inlined or eliminated)
```
## adapter-Uint128-QuoRem.asm.txt
```text
  m3_adapter.go:120	0x53b5ee		e80df1ffff		CALL github.com/jokruger/dec128/uint128.Uint128.QuoRem(SB)			
  m3_adapter.go:119	0x53b635		e8c66cf3ff		CALL runtime.morestack_noctxt.abi0(SB)						
```
## historical-QuoRem256By128.asm.txt
```text
  tools.go:14		0x53ae62		e8d9f7ffff		CALL github.com/profe-ajedrez/badassitron/dec128/uint128.Uint128.QuoRem(SB)	
  tools.go:74		0x53b2cb		e85091f3ff		CALL runtime.panicIndex(SB)							
  tools.go:57		0x53b2d0		e8eb98efff		CALL runtime.panicoverflow(SB)							
  tools.go:57		0x53b2d5		e88698efff		CALL runtime.panicdivide(SB)							
  tools.go:53		0x53b2e0		e83b91f3ff		CALL runtime.panicIndex(SB)							
  tools.go:53		0x53b2ea		e83191f3ff		CALL runtime.panicIndex(SB)							
  tools.go:39		0x53b2ef		e80c98efff		CALL runtime.panicshift(SB)							
  tools.go:121		0x53b2f4		e8c798efff		CALL runtime.panicoverflow(SB)							
  tools.go:118		0x53b2f9		e86298efff		CALL runtime.panicdivide(SB)							
  tools.go:12		0x53b320		e8db6ff3ff		CALL runtime.morestack_noctxt.abi0(SB)						
```
## historical-Uint128-Mul.asm.txt
```text
```
## historical-Uint128-QuoRem.asm.txt
```text
  arithmetic.go:158	0x53a94c		e86fa2efff		CALL runtime.panicoverflow(SB)							
  arithmetic.go:158	0x53a951		e80aa2efff		CALL runtime.panicdivide(SB)							
  arithmetic.go:138	0x53a96b		e89079f3ff		CALL runtime.morestack_noctxt.abi0(SB)						
```
