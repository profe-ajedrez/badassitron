> **Historical migration record (archived at M7-H).** This file contains

> **Archived engineering record.** This document describes a dated source
> checkpoint, not the currently recommended API or release state.
> For supported usage, see the [documentation index](../../../docs/README.md).
> time-specific stage status, proposed code and, in places, old rollback
> tag instructions that **do not apply to the current source tree**.
> The authoritative state is [M7-H release status](../../../docs/RELEASE_STATUS.md);
> do not follow an old "current" or "pending" claim as a live rule.

# Badassitron — Plan maestro de migración a Jokruger dec128

**M3-F update (2026-11-17, current source tree):** the public `dec128/uint128`
arithmetic **now defaults to imported Jokruger v1.4.0**, with historical
arithmetic available only through an explicit rollback tag during final
validation. Public signatures/representation remain compatible. The older
M3-E decision below is historical and is superseded by D65–D68. See the
[M3-F default migration report](../m3f-default-imported-uint128-20261117.md).
Go 1.27.1 verification on the new source remains pending.


**Estado:** original plan; M1 implementation/verification in progress; **rama objetivo:** `migration/jokruger-dec128-v1.4.0`; **fuente local auditada:** `badassitron-master(6).zip`; **motor previsto:** `github.com/jokruger/dec128 v1.4.0` (versión fijada).  
**Principio rector:** dependencia declarada, procedencia explícita, API pública de Badassitron estable, semántica verificada y migración atómica por tipo.  
**Idioma de documentación del repositorio:** inglés. Este plan de coordinación se redacta en español; los archivos de implementación indicados abajo se escribirán en inglés.

**M3-E architectural resolution (2026-11-17):** the default M2-E decimal coefficient engine uses imported Jokruger primitives directly, while the public `dec128/uint128` nominal type remains historical by default. The M3 unsigned adapter is **opt-in only** after the user-supplied Go 1.27.1 benchmarks showed persistent overhead. Added a source boundary guard, explicit internal coefficient constructor and 4,000 boundary cases. See [M3-E report](../m3e-architecture-resolution-20261117.md). Publication still needs official M3-E verification and historical provenance approval.

**M3-D experimental unsigned-division checkpoint (2026-11-17):** Narrowed ordinary division-state mapping, added 2,000 interop input sets and 11-path benchmark coverage, including error paths. M3 remains opt-in; official Go 1.27.1 verification is pending. See [M3-D division-status and 11-path report](../m3d-division-status-three-way-20261117.md).

**M3-C unsigned adapter checkpoint (2026-11-17):** experimental `Mul`
wrapper is now inlinable under the local Go 1.23.2 compiler, with new
exact-success and overflow comparisons. M3 remains opt-in because measured
regressions persist; Go 1.27.1 official-module verification is pending. See
[M3-C report](../m3c-mul-inline-20261117.md).

**M3-A source experiment (2026-11-17):** a complete `dec128/uint128`
public unsigned-arithmetic adapter is available only with
`-tags badassitron_m3`. The AST inventory matches all 66 exported
`uint128` declarations and local math/test/race gates pass, but sampled
Add/MulOverflow performance regressions preclude default
promotion. See [M3-A report](../m3-uint128-opt-in-20261117.md).

**M2-E source-tree promotion (2026-11-17):** the Jokruger-backed M2-D
implementation is now selected by default, with historical `dec128` available
through `-tags badassitron_legacy`. This is an **unpublished rollback-capable
checkpoint**, not a completed M7 release. Official Go 1.27.1 results were
supplied for the unchanged pre-promotion M2-D code; a fresh run on this tree,
historical authorship review, and acceptance of the residual 12.6% sampled
`Rescale/Truncate` slowdown remain open. See the [M2-E report](../m2e-default-promotion-20261117.md).

**M2-D historical status (2026-11-17):** [DivRound and official audit automation](../m2d-fast-division-provenance-20261117.md). Locally, both full engines, race, 5,243 frozen cases and >63k independent division/rounding assertions pass. Official Go 1.27.1 tag/source/provenance and representative M6 approval remain open; default stays legacy.

**M2-C status (2026-11-17):** [Performance and provenance checkpoint](../m6c-performance-provenance-20261117.md). The previous M2-B implementation passed an externally supplied Go 1.27.1 full verification log; M2-C improves multiple measured paths but has only a local Go 1.23.2/snapshot pass. Default remains legacy. Full official rerun, final provenance and release decision are still outstanding.

**Current M2-B extension:** [2026-11-17 opt-in whole-package public cutover and broad M6 matrix](../m2b-experimental-public-cutover-20261117.md). A `badassitron_m2` build tag now selects the complete candidate within the public `dec128` package; the default remains historical. Both local suites, oracles and scoped races pass, but official Go 1.27/Jokruger-tag, full race, provenance and M6 gates remain open. The six temporarily mirrored M2 sources are protected by a consistency check.

**Current checkpoint:** [2026-11-17 M1-final / isolated M2-CANDIDATE-A report](../m1-final-m2-candidate-a-20261117.md). The previous [M1 parser and signed-zero checkpoint](../m1-parser-signed-zero-20261117.md) remains in the historical record. M0 provenance and M1 official toolchain/upstream verification are still open; M2 has a full-declaration-surface isolated candidate and 5,243 local frozen oracle passes; the atomic public replacement has not happened. The current dependency declaration is present in `go.mod`, but the Go module checksum is unavailable offline.

## 0. Diagnóstico y decisiones de arquitectura

### 0.1 Hechos constatados en el snapshot

- El módulo es `github.com/profe-ajedrez/badassitron`, con `go 1.27` en `go.mod`.
- Paquetes afectados: `dec128`, `dec128/errors`, `dec128/uint128`, `dec256`, `dec256/uint256`, `dec512`, `dec512/uint512`, `optimus`, `internal/decimalwide`, `internal/uintwide`, `withdec128`, `withfloat64`, ejemplos y documentación.
- `dec128.Dec128` contiene coeficiente de 128 bits, escala `uint8` y estado `uint16`; su `Precision()` es **escala almacenada**, no dígitos significativos. Se documenta una representación de 24 bytes en amd64 en `docs/FASTPATH_V4.md`.
- `dec128.Mul` es exacto o NaN; el error histórico de fallo de encaje es `errors.Overflow`. Hay un fast path con `bits.Mul64` para coeficientes de 64 bits.
- `dec128.Round` es half-away-from-zero; existen siete valores públicos `RoundingMode`, numerados de 0 a 6.
- `dec128/errors.Error` usa códigos históricos 0–10; `Value()` actualmente indexa una tabla sin validar rango.
- `dec128.Dec128.Scan` tiene receptor por valor; existen `ScanInto` y `ScanDestination`.
- `dec256` y `dec512` realizan aritmética con `MulCarry` y `QuoRem`; `internal/decimalwide` utiliza los tipos públicos de errores y redondeo de `dec128`.
- `docs/FASTPATH_V4.md` registra referencia de rendimiento en Ryzen 7 7735HS con Go 1.27.1, no una medición nueva de esta migración.

### 0.2 Hechos comprobados de Jokruger v1.4.0

- Proporciona `MulRound`, `DivRound`, `RescaleRound` y `ROUND_NAN`, así como política de pérdidas `LossPolicy`.
- La política global predeterminada (`LossRound`) permite redondeo; las operaciones ordinarias no son sustituto semánticamente equivalente de las ordinarias de Badassitron.
- Las variantes explícitas pueden producir `state.Inexact` o `state.InvalidRoundingMode`, sin equivalencia numérica directa con los códigos históricos.
- `MulAddRound` y `MulDivRound` existen upstream; su adopción como nueva API pública queda fuera del alcance de compatibilidad inicial.
- Los ajustes globales de upstream no deben alterar la corrección de las operaciones públicas de Badassitron.

**Fuentes upstream:** `https://pkg.go.dev/github.com/jokruger/dec128@v1.4.0`; antes de codificar, inspeccionar también el *source* y las licencias/NOTICE reales del tag fijado.

### 0.3 Decisiones vinculantes

1. **No venderizar ni copiar** de nuevo la implementación decimal de Jokruger. Añadir `require github.com/jokruger/dec128 v1.4.0` y conservar `go.sum`.
2. **Conservar** la ruta `github.com/profe-ajedrez/badassitron/dec128` y su tipo nominal propio. `type Dec128 struct{ value jk.Dec128 }` es la hipótesis inicial, sujeta a prueba de invariantes (escala, cero negativo, errores); no se presupone compatibilidad binaria ni que todas las conversiones sean gratuitas.
3. **Mantener** el vocabulario público histórico (`RoundingMode`, `errors.Error`, `uint128`, constructores, interfaces, nombres y firmas), salvo decisiones aprobadas y documentadas.
4. **Errores con contexto:** normalización en los métodos externos; los caminos internos críticos pueden operar con primitivas upstream y traducir al salir de su frontera, sin traducción por operación elemental.
5. **Migraciones atómicas por tipo:** ninguna entrega deja `Add` delegado y `Compare` antiguo en un mismo tipo. La implementación histórica se congela en una referencia **fuera del módulo principal** y se usa solamente como oráculo de pruebas, sin redistribuirla innecesariamente.
6. **Sin cambios oportunistas:** separación de módulos, generación de fuentes anchas, mejoras Knuth D, cambios de contrato tributario y bajada de versión mínima de Go son tareas posteriores y separadas.
7. **Compatibilidad verificable:** distinguir API fuente, semántica, representación/serialización, estados de error, concurrencia y rendimiento. No se promete compatibilidad binaria/`unsafe`.
8. **No publicar nuevos tags** hasta cerrar los gates de M7.

### 0.4 Propuesta de estructura final

```text
badassitron/
  dec128/                 # fachada pública histórica, adaptada a jk.Dec128
    errors/               # códigos públicos históricos y traducciones seguras
    uint128/              # API pública histórica; adaptador si es preciso
  internal/dec128bridge/  # construcción, traducción de estados/modos, reglas invariantes
  dec256/ dec512/         # decimales amplios con API histórica
  optimus/               # selección/promoción por capacidad y exactitud
  internal/decimalwide/   # primitivas y políticas comunes, sin dependencias cíclicas
  docs/
    DEC128_MIGRATION_CONTRACT.md
    MIGRATION_DECISIONS.md
    MIGRATION_VALIDATION.md
    MIGRATION_PERFORMANCE.md
  NOTICE.md
```

La ubicación final del puente puede cambiar si aparece un ciclo de dependencias; las reglas importan más que el nombre del directorio.

## M0 — Migration Contract: contrato, evidencia y atribución inmediata

**Objetivo:** inmovilizar el contrato observable del snapshot y dejar constancia responsable de procedencia **antes** de modificar algoritmos. **No se reemplaza aritmética.**

**Actividades obligatorias:**

- M0.1 Guardar huella verificable del ZIP/revisión (`sha256`, commit si existe, Go, OS/arquitectura, CPU, `go env`) y fijar la versión exacta del oráculo histórico.
- M0.2 Inventariar automáticamente símbolos exportados (`go doc`, AST o `go list`) y sus firmas de `dec128`, `errors`, `uint128`, `dec256`, `dec512`, `optimus` y adaptadores; guardar snapshot legible por máquina y humano.
- M0.3 Extraer invariantes por método: constructores, parseo (`""`, `"5."`, signos, ceros), escalas, `Canonical`, `Precision`, `Exponent`, `Coefficient`, estados NaN, propagación de primer NaN, `Equal/Compare`, aritmética, sign-zero, round/trunc, `WithStatus`, formatos, JSON/text/binary/SQL.
- M0.4 Construir **matriz de traducción contextual**: operación; causa upstream (`Overflow`, `Underflow`, `Inexact`, `Null`, `InvalidRoundingMode`, `DomainError`, etc.); código antiguo esperado; preservación o cambio autorizado. No usar conversión ordinal de enums.
- M0.5 Congelar tablas de códigos públicos `errors.Error` (0–10) y `RoundingMode` (0–6). Los nuevos códigos se **anexan**, sin renumerar. Validar `None`, desconocidos 11–255 y manejo seguro de `Value()` / `Error()`.
- M0.6 Congelar reglas de compatibilidad no trivial: `Mul` exacto-o-NaN con error histórico; divisiones truncadas; límites de escala; `WithStatus` significa **dígitos no nulos descartados**, no error; cero negativo; configuración global de Badassitron y concurrencia.
- M0.7 Fijar las **excepciones aprobadas** antes del test diferencial: overflow de borde de `Mul`, `PowInt(math.MinInt)`, error inválido y tratamiento exacto de `1e20 / 1`. Para `"5."`, mantener por defecto mientras se pruebe. Para `Scan` por valor, conservar símbolo como obsoleto, sin atribuirle funcionamiento correcto; recomendar `ScanDestination` hasta cambiar receptor con una decisión expresa.
- M0.8 Revisar código y licencias de procedencia (`jokruger`, `udecimal`, `lukechampine` u otros) por archivo y evidencia; incorporar **hoy** `NOTICE.md` con afirmaciones verificadas, enlaces, licencia y estado (copia anterior / dependencia futura), sin afirmaciones no comprobadas.
- M0.9 Capturar línea base funcional (tests, race, vet, ejemplos, docs) y de rendimiento `FASTPATH_V4` con mismo hardware si está disponible; si no, guardar la referencia existente como histórica sin inventar números nuevos.
- M0.10 Guardar oráculo legacy en checkout/tag/archivo inmóvil externo al módulo, sujeto a revisión de licencias; su ejecución se orquesta desde tests de integración o proceso separado.

**Entregables:**

```text
docs/DEC128_MIGRATION_CONTRACT.md        # autoridad del contrato congelado
docs/MIGRATION_DECISIONS.md             # decisiones por conducta + motivo
docs/MIGRATION_VALIDATION.md            # entorno, comandos, matrices, oráculos
NOTICE.md                               # atribución verificable inmediata
testdata/migration/api-baseline.json    # snapshot legible por máquina (o ubicación equivalente)
testdata/migration/golden/*.json        # entradas y resultados esperados
reports/migration/m0-baseline.md        # estado y medidas reproducibles
```

**Pruebas/gates:** inventario revisado; ejemplos compilan; suite ordinaria, race y vet (fallas preexistentes clasificadas); ninguna firma pública perdida; estados fuera de rango no causan panic después de la corrección autorizada; NOTICE respaldado. **Cierre:** M0 `GREEN` solo con contrato y matriz aprobados; no avanzar a M1 con interrogantes semánticos sin dueño/decisión.

## M1 — Preparación del puente y harness diferencial

**Objetivo:** construir infraestructura sin cambiar la implementación pública activa.

- Fijar upstream `v1.4.0`; verificar checksums, Go mínimo real, licencia, NOTICE, API y límites.
- Crear prototipo **privado** del puente: modos explícitos, constructor seguro de coeficiente+escala+signo, lectura de estado, conversión de errores y manejo de cero negativo.
- Probar si la representación por composición permite `Coefficient`, `Exponent`, signed zero, NaN con códigos históricos y escala idénticos; si no, evaluar una fachada con metadatos mínimos sin duplicar motor aritmético.
- Diseñar rutas para operaciones sin equivalente 1:1 (`Mod`, `QuoRem`, `PowInt`, `Scan`, `New` y métodos `WithStatus`).
- Ejecutar tests diferenciales mediante dos módulos/procesos separados (legacy y candidata), con formato de casos versionado; jamás importar a la vez rutas del mismo módulo de revisiones diferentes como si fuesen paquetes independientes.
- Matriz de prueba: small / middle / max-coefficient; escalas 0,1,18,19 y fuera de rango; signos; ceros con signo; NaN por cada causa; operaciones que exceden 128 bits; underflow, división no exacta, empate de redondeo; serialización.
- Crear test de independencia frente a globales upstream (configurados **antes** de lanzar las operaciones; evitar mutar globales durante cálculos concurrentes por las carreras descritas por upstream).

**Entregables:** puente privado, generadores, comparador de fixtures, mapa de incompatibilidades. **Gate:** 100 % de casos esperados clasificados y rutas técnicas viables; ninguna dependencia circular.

## M2 — Sustitución atómica del tipo público `dec128`

**Objetivo:** migrar **todo** `dec128` en un cambio integrable único, manteniendo fachada histórica y usando Jokruger internamente.

1. Implementar constructores, constantes y representación de NaN/signos/escala.
2. Implementar aritmética ordinaria (Add/Sub/Mul/Div/Mod/QuoRem/PowInt/Sqrt) conforme al contrato M0; `Mul` **no** llama sin validación a upstream `Mul` sujeto a `LossRound`.
3. Implementar variantes explícitas (`MulRound`, `DivRound`, `RescaleRound`, `SqrtRound`) con mapeo de los siete modos históricos; mantener semántica exacta de `WithStatus`.
4. Para `Mul`, considerar fast path con `math/bits.Mul64` solo tras demostrar ajuste del coeficiente, suma de escalas y preservación de cero/signo; usar fallback upstream con `ROUND_NAN` o ruta equivalente verificada. **Cuidado:** las variantes upstream de escala fija pueden diferir del algoritmo histórico de *escala flexible*; diseñar selección exacta de escala antes de delegar.
5. Implementar Compare/Equal/Canonical y orden/propagación de NaN; no asumir que `jk.Compare` reproduce por sí solo contratos históricos.
6. Implementar parseo, texto, JSON, interfaces SQL, conversión a `float64`, enteros y binario con golden previos.
7. Ejecutar diferencial completo antiguo↔nuevo; cada diferencia debe apuntar a `MIGRATION_DECISIONS.md` y tener test explícito.
8. Asegurar que `dec128/uint128` y `dec128/errors` se mantienen compilables para consumidores externos mientras `internal` puede usar el `uint128` upstream directamente, sin forzar alias si cambia el tipo de retorno.

**Entregables:** `dec128` integrado como fachada; dependencia Go fijada; pruebas de API y semántica; decisión explícita sobre `Scan` y sobre compatibilidad de `uint128`.

**Gate:** suite de `dec128` GREEN; consumidores de todo el módulo compilan; cero diferencias no aprobadas; ausencia de acceso a configuración upstream en cálculos deterministas; `go test -race` para los caminos cubiertos. **Regla de integración:** no aceptar una PR intermedia con un `Dec128` semánticamente mezclado.

## M3 — Reorganización de primitivas y consumidor interno

**Objetivo:** quitar coste de compatibilidad del *hot path* sin romper los paquetes externos.

- Auditar `dec128/uint128` vs `jk/uint128` por layout, funciones, códigos de error, `MulCarry`, `QuoRem`, `QuoRem256By128`, mutabilidad y overflow; validar qué funciones existen realmente en v1.4.0.
- Usar upstream directamente solo en caminos internos donde el contrato coincida; adaptar en la **frontera de paquete** para `dec128/uint128` público.
- Mantener `errors.Error` como tipo histórico; traducir estados una vez por operación pública, no en cada paso del algoritmo ancho.
- Remover duplicación probada del código de 128 bits, sin borrar primitivos genuinos de 256/512 bits necesarios para los tipos anchos.
- Medir costes de conversiones, inline y escape (bench + `-gcflags=-m=2` selectivo); no afirmar «coste cero» por inspección.

**Gate:** equivalencia completa de `uint128` público, benchmarks representativos sin regresión injustificada y sin nuevas dependencias cíclicas.

## M4 — Validación de `dec256`, `dec512` y el núcleo ancho

**Objetivo:** alinear reglas compartidas conservando capacidades distintas, **no** reducir los tipos anchos a un `jk.Dec128`.

- Migrar tipos **uno por uno y de forma atómica**: primero `dec256`, después `dec512`. Se conserva la representación propia de coeficientes anchos.
- Reemplazar dependencias internas sobre la implementación antigua de `dec128` por interfaces estables / primitivos directos y adaptadores de frontera.
- Mantener límites de escala 77 y 154; no imponer el máximo 19 de Jokruger a tipos anchos.
- Replicar en lo aplicable las reglas públicas congeladas de errores, modos y `WithStatus`, incluidas divisiones y redondeos; `internal/decimalwide` no debe redondear anticipadamente.
- Oráculo dual: si entradas/resultado caben en 128 bits y **se exige idéntica semántica**, comparar contra fachada `dec128` y upstream con modo explícito; fuera de ese espacio, `math/big.Int` / `big.Rat` con cuantización independiente.
- Verificar conversiones exactas 128→256→512 y de vuelta, identidad decimal, sign-zero, NaN, serialización y SQL.

**Gate individual por tipo:** compile + unit + fuzz corpus + race + golden + benchmarks + documentación; no migrar `dec512` antes de cerrar `dec256`.

## M5 — `Optimus`: selección de tier y promoción por exactitud

**Objetivo:** `Optimus` conserva la propiedad central de calcular exactamente cuando su capacidad adaptativa lo permite.

- Definir invariante `Optimus`: el tier 128 solo delega en Jokruger cuando **demuestra** capacidad de representar el resultado exacto y su escala; de lo contrario promueve a 256/512 según necesidad.
- Diferenciar NaN por `Inexact` de overflow irrecuperable y errores de dominio: solo promover cuando una anchura mayor puede resolver la operación, **no** como respuesta indiscriminada a todo NaN.
- No usar `jk.Mul`, `jk.Div` ordinarios si pueden redondear por política global; usar variantes explícitas o aritmética exacta validada.
- Asegurar que el escalado flexible, resultados con hasta 154 decimales y operaciones `MulRound/DivRound` no pierden dígitos antes de una posible promoción.
- Verificar descenso de tier después de cancelación/canonicalización cuando corresponda, límites de potencia, consultas de exactitud y coste de conversiones.
- Probar con escenarios SUNAT/invoice como **fixtures numéricos**, sin convertir reglas tributarias en responsabilidad del motor decimal.

**Gate:** suite de tiers/cambio de capacidad, oráculos racionales, concurrencia e invariantes de no redondeo prematuro; pruebas con las tres configuraciones globales upstream; benchmarks small/wide/promote.

## M6 — Rendimiento, documentación y saneamiento

**Objetivo:** cuantificar la migración, explicar los trade-offs y eliminar código duplicado de forma segura.

- Ejecutar benchmarks repetidos en un mismo host/Go/flags, `GOMAXPROCS=1` para microbenchmarks secuenciales y pruebas aparte para concurrencia. Capturar `ns/op`, `B/op`, `allocs/op`, CPU, frecuencia/energía si se conoce, commit, `go version`, `go env`.
- Incluir: Add/Sub/Mul/Div/Compare/Parse/String/Rescale/SQL; fast path 64 bits, coeficientes grandes, escalas 0/19/77/154, errores, promoción y trabajo end-to-end. Medir `jk.Dec128` directa, Badassitron anterior y posterior; `shopspring/decimal` u otras solo si se iguala operación y política.
- Comparar contra `docs/FASTPATH_V4.md` únicamente con advertencia de entorno; nueva línea base no reemplaza números históricos.
- Usar `benchstat` si está disponible, incluir resultados crudos y dispersión; definir presupuesto orientativo de regresión (por ejemplo, **≤10 % mediana en operaciones comunes** salvo justificación aprobada), no como criterio ciego para funciones que cambiaron correctamente su semántica.
- Auditar source similarity y eliminar implementación heredada que ya no cumple función: **preservar** código original requerido para anchuras 256/512 y wrappers públicos.
- Actualizar README, `docs/README.md`, `DECIMAL_BEHAVIOR.md`, ejemplos, GoDoc, `NOTICE.md` y, tras auditoría final, `CHANGELOG.md`; enlazar informe comparativo.
- Evaluar por separado `withdec128`, `withfloat64`, APIs deprecated, `go.mod` mínimo y generación de wide types; no retirarlos en esta migración.

**Entregables:** `docs/MIGRATION_PERFORMANCE.md`, reportes `.txt` crudos, tabla de compatibilidad, historial de decisiones y referencias enlazadas desde README.

**Gate:** diferencias de rendimiento explicadas; documentación consistente con código probado; sin duplicación no justificada ni afirmaciones de procedencia sin evidencia.

## M7 — Auditoría final, release y seguimiento

**Objetivo:** publicar una versión verificable que no esconda cambios.

- Ejecutar suites en versiones de Go soportadas *realmente*; no bajar `go.mod` de 1.27 sin un paso dedicado que compile y pruebe en 1.24/1.25/1.26 y revise uso de APIs nuevas.
- Validar `go vet`, `go test -race`, ejemplos, fuzz con corpus, `go mod tidy`, `go mod verify`, `go list -m all`, documentación y licencias.
- Verificar que el código heredado congelado ya no forma parte del árbol distribuido, salvo fragmentos explícitos necesarios y licenciados.
- Revisar compatibilidad fuente por snapshot AST; enumerar excepciones deliberadas en release notes. **Un cambio de receptor de método puede afectar métodos de interfaces o method sets**, aun cuando una llamada simple siga compilando.
- Crear tag solo tras aprobar checklist completo y revisar semver frente a diferencias de comportamiento.
- Añadir CI de actualización informativa para Jokruger `@latest` sin actualizar automáticamente la dependencia fijada; si aparece v2, tratarlo como migración explícita.

**Gate de publicación:** M0–M6 GREEN, 0 diferencias no clasificadas, licencias aprobadas, reporte reproducible, Go compatible, changelog y aprobación de release.

## Matriz mínima de compatibilidad (autoridad: M0)

| Área | Preservar inicialmente | Excepción/corrección candidata | Verificación |
|---|---|---|---|
| Import paths y firmas | Todos los paquetes públicos, métodos, constructores y interfaces | Solo deprecaciones acordadas | Snapshot AST + compilación consumidores |
| Escalas | 19/77/154; `Precision()` = escala almacenada | Ninguna | Vectores de fronteras |
| `Mul` | Exacto o NaN, error contextual histórico | Corregir overflow de borde | Differential + `big.Int` |
| `Div`, `Sqrt` | Política histórica explícita, sin global upstream | `1e20/1` exacto cuando representable | Golden |
| Redondeo | Siete modos y valores numéricos | Nuevos modos solo al final, si se agregan | Tie cases ± |
| Error codes | 0–10, significado histórico | Evitar panic en 11–255; anexar nuevos | Exhaustivo 0–255 |
| Parsing | Sintaxis actual, `"5."`, empty→0 | Invalid format solo si explícitamente aprobado | Golden + fuzz |
| Signed zero y NaN | Representación observable, propagación | Bugs confirmados | Comparaciones, StringFixed, JSON |
| SQL | `ScanInto`, `ScanDestination`, Valuer actual | `Scan` por valor deprecated; receptor puntero requiere decisión API | Compile-time interfaces + base fake |
| Uint128 público | Nombre y retornos históricos | No reemplazar automáticamente por alias upstream | API test |
| Configuración | `SetDefaultPrecision` y snapshot concurrente | Independencia de globales upstream | Race + valores invariantes |
| Rendimiento | FASTPATH_V4 como referencia histórica | Regresiones justificadas por corrección | Benchstat y escenarios |

## Protocolo operativo común a todos los hitos

**Un hito = un PR integrable, salvo M2 que puede desarrollarse en una rama interna pero aterriza atómicamente.** Cada PR debe añadir: objetivo, procedencia, rutas tocadas, firmas afectadas, diferencias aprobadas, comandos ejecutados, resumen de benchmarks y rollback.

### Comandos de referencia (ajustar a la versión Go exigida)

```sh
GOWORK=off go test ./... -count=1 -timeout=20m
GOWORK=off go test -race ./... -count=1 -timeout=20m
GOWORK=off go vet ./...
GOWORK=off go test ./... -run '^$' -bench=. -benchmem
GOWORK=off GOMAXPROCS=1 go test ./dec128 -run '^$' \
  -bench '^BenchmarkV4Common/' -benchtime=250ms -count=5 -benchmem
python3 tools/check_doc_links.py
```

La suite fuzz se programa por objetivos con corpus, semillas y duración registrados, no se reemplaza por `go test ./...` convencional.

### Reglas de reversión

- M0–M1: revertir solo archivos de preparación, sin cambio semántico.
- M2: revertir **el cambio atómico completo** de `dec128`; nunca mezclar implementación parcial antigua/nueva.
- M3–M5: revertir el tipo/paquete atómico correspondiente y sus adaptadores; conservar los golden aprobados.
- M6–M7: no publicar si aparecen diferencias no explicadas; reconstruir desde la última revisión GREEN.

### Riesgos críticos y medidas

1. **Diferencia en representación de signed zero/NaN:** prototipo M1; metadatos propios si la composición simple no basta.
2. **Escala flexible frente a `*Round` de escala fija:** selector de escala explícito y oráculo M2.
3. **Mapeo erróneo `Inexact→Overflow`:** tablas por operación y pruebas de frontera M0/M2.
4. **Globals upstream y carreras:** solo variantes deterministas; nunca cambiar ajustes globales durante pruebas paralelas.
5. **Fuga de tipos upstream a firmas públicas:** snapshot AST y consumidores externos compilados.
6. **Regresión fast path:** medir antes/después en mismo host; conservar `bits.Mul64` si demuestra beneficio.
7. **Licencias y procedencia:** NOTICE inmediato, auditoría por archivo y CHANGELOG al cierre.
8. **Go 1.27 vs 1.24:** no hacer downgrade oportunista; evaluar en tarea separada.
9. **Referencia legacy no portable:** ejecutarla externamente con Go y entorno fijados; resultados serializados versionados.
10. **Cambios invisibles en `Scan` e interfaces:** tests de conjuntos de métodos y documentación deprecada.

## Dependencias y orden de integración

```text
M0 contract + notice + golden
           |
           v
M1 bridge + differential harness
           |
           v
M2 dec128 (atomic) ----> M3 internal primitives
                                |
                                v
                         M4 dec256 then dec512
                                |
                                v
                         M5 optimus adaptive
                                |
                                v
                         M6 performance/docs/cleanup
                                |
                                v
                         M7 final audit/release
```

**Condición de inicio:** primero aprobar y convertir este plan en tareas verificables; la primera tarea de código pertenece a M0, no a M2. **Condición de término:** Badassitron depende de Jokruger donde aporta valor, no contiene una segunda implementación decimal de 128 bits injustificada y conserva su API/sus garantías documentadas con excepciones explícitas y medidas reproducibles.
