# BADASSITRON-OPT-01 + IACON-ARCH-02A — Informe técnico

**Estado:** cambios implementados y validados en el entorno local disponible; validación con Go 1.27 pendiente.

## Qué se modificó

1. `badassitron/optimus/optimus.go`: `RescaleRoundWithStatus` intenta primero la cuantización existente de `dec512` (ancho fijo); si falla, preserva exactamente `decimalwide.Quantize` como respaldo. El retorno utiliza `Optimus.New`, conserva el ancho adaptativo, precisión, error y `inexact`. No modifica firmas públicas ni otros operadores.
2. `iacon/pipeline/plan.go`: reserva adaptativa de mapas y slices en `Compile` para documentos a partir de 64 líneas. Se conservan los constructores originales para 1–63 líneas. La capacidad estimada no limita el crecimiento y no cambia la validación o las reglas.
3. Se incorporaron tests diferenciales deterministas en Badassitron, tanto independiente como incluido en Iacon. Los documentos de arquitectura y los logs reproducibles se incluyen en los ZIP.

## Invariantes comprobados

- Se mantienen los siete modos de redondeo; precisión, signo, coeficiente, error e indicador de inexactitud se comparan con la implementación original.
- La batería diferencial cubre **8000 casos aleatorios** y **462.672 combinaciones** de límites de 512 bits, signos, escalas y modos; todas pasan.
- La suite completa de **Badassitron** pasa (`go test ./...`) y `go vet ./...` termina sin problemas.
- La suite completa de **Iacon** pasa (`go test ./...`), incluyendo los **16 goldens SUNAT** originales, `go vet ./...` y `go test -race ./...`.
- Las pruebas diferenciales de Optimus también pasan con `go test -race ./optimus -run ...`.
- No se cambiaron fórmulas, reglas, snapshots, trazas, APIs ni dependencias.

## Medición antes/después

**Condiciones:** Intel Xeon Platinum 8573C compartido, Linux/amd64, Go 1.23.2; `GOMAXPROCS=1`, `-cpu=1`. Medianas de 3 ejecuciones de 10 iteraciones por escenario del pipeline, y de 30 iteraciones por microbenchmark. Memoria y objetos asignados por operación, no memoria residente. Se midieron 64 escenarios: 1, 10, 100, 1000 líneas × cuatro backends × compilación, ejecución sin/con traza, y ciclo completo.

### Pipeline de 1000 líneas, sin traza

| Backend | Tiempo antes → después (ms) | B/op antes → después | allocs/op antes → después |
|---|---:|---:|---:|
| Optimus | 13.77 → 12.33 | 5.48 → 4.09 MB | 61,619 → 474 |
| Dec128 | 7.76 → 7.98 | 3.48 → 3.48 MB | 478 → 487 |
| Dec256 | 11.02 → 10.37 | 4.70 → 4.70 MB | 486 → 490 |
| Dec512 | 15.22 → 14.53 | 7.16 → 7.16 MB | 497 → 496 |

En Optimus, las asignaciones de `Execute(false)` de 1000 líneas cayeron **aproximadamente un 99,2 %**. En los otros backends, el mismo camino de ejecución no cambia de forma significativa; no se debe presentar la optimización de Optimus como una mejora general del motor.

### Compilación de 1000 líneas

| Backend | Tiempo antes → después (ms) | Memoria antes → después (MB/op) |
|---|---:|---:|
| Optimus | 9.04 → 6.44 | 6.35 → 3.70 |
| Dec128 | 9.04 → 5.84 | 6.12 → 3.55 |
| Dec256 | 8.29 → 6.12 | 6.57 → 3.85 |
| Dec512 | 9.26 → 5.68 | 7.50 → 4.45 |

La reserva de capacidad reduce ~41–42 % los bytes/op de compilación para 1000 líneas. La cantidad de asignaciones baja menos del 1 %, porque evita realojamientos grandes más que objetos pequeños.

### Cuantización Optimus aislada

| Caso | Antes → después (ns/op) | Asignaciones antes → después |
|---|---:|---:|
| `same_scale` | 491 → 263 | 9 → 0 |
| `round_half_even` | 651 → 595 | 14 → 0 |
| `round_negative_floor` | 584 → 518 | 12 → 0 |
| `wide_76_digits` | 807 → 530 | 14 → 1 |
| `widen_scale` | 581 → 501 | 11 → 0 |

Las rutas comunes de Optimus alcanzan 0 allocs/op en el microbenchmark. La ruta de coeficiente ancho todavía puede asignar para representar el valor en `Optimus` (1 alloc/op en la prueba). Para algunos casos medidos `B/op` redondea a 1 con 0 allocs/op; no se afirma cero bytes para todas las entradas.

### Precaución sobre CPU

El servidor compartido presentó fluctuaciones grandes durante una repetición más larga, que no completó. Por ello, los cambios de CPU descritos arriba son **exploratorios** y no constituyen una garantía estadística de aceleración. El ahorro de memoria y asignaciones es más consistente en las tres réplicas completas. Los archivos de la repetición incompleta están separados y no se utilizaron en las medianas.

## Toolchain y compatibilidad

Los tres `go.mod` originales declaran `go 1.27`. Para validar se adaptaron **solo copias temporales** a `go 1.23`, que es la versión instalada. Los ZIP publicados contienen `go 1.27` sin cambios. **La compatibilidad con Go 1.27 no está comprobada.** Los benchmarks deben repetirse con ese compilador antes de atribuirles valor definitivo.

## Riesgos y pendientes

- Revalidar con Go 1.27 y CPU dedicada; también con documentos heterogéneos y valores reales de producción.
- La preparación de planes no se reutiliza universalmente; `Bind` conserva el comportamiento anterior.
- Una transformación/resultado independiente y una traza publicada seguirán requiriendo asignaciones por contrato.
- El fallback `decimalwide.Quantize` persiste para entradas inválidas y resultados no representables; por tanto **no se promete Zero Allocation absoluto**.
- No se justificó introducir `sync.Pool`, microlibrerías, slots directos o publicación fusionada.

## Artefactos

- `badassitron_opt_01_complete.zip`: Badassitron completo con prueba diferencial y ruta nativa.
- `iacon_arch_02a_complete.zip`: Iacon completo con Badassitron incluido, reserva de capacidad y registros.
- `BADASSITRON-OPT-01-comparativa.csv`, `IACON-ARCH-02A-comparativa.csv`: métricas comparadas; números negativos de delta representan mejora.
- Diffs y logs en `docs/perf_opt_01_evidence` y `docs/benchmarks_arch_02a` dentro de los ZIP.
