# Testing

> El código legacy es código sin tests.
> Los tests son la red de seguridad que
> permite cambiar con confianza.

## 1. El código legacy es código sin tests

Identificá puntos de cambio. Creá "costuras"
(seams) para añadir tests antes de modificar.

**Algoritmo:**

1. Identificá puntos de cambio
2. Encontrá puntos de test
3. Rompé dependencias
4. Escribí tests de caracterización

## 2. Escribí tests que fallen, no que pasen

Asegurate de que el test falla al romper
el código. Probá comportamiento, no
implementación.

**Verificación:** Rompé el código. El test
debe fallar.

## 3. Refactorizá sin piedad, pero con tests

Cambios estructurales pequeños y frecuentes
protegidos por tests.

**Regla del Tres:**

1. Primera vez: hacelo
2. Segunda: notá duplicación
3. Tercera: refactorizá

## 4. El mejor bug es el que no llega a producción

Capas de defensa: lint, tipos, tests,
preview deploys.

**Coste:** Un bug en producción cuesta
10,000x más que uno detectado por el linter.

## 5. No optimices lo que no mediste

Profilers para identificar el cuello de
botella real. La intuición sobre
rendimiento suele estar equivocada.

## 6. Los tests son la mejor documentación

Los tests que describen reglas de negocio
nunca mienten.

**TDD:** Diseñá la API desde la perspectiva
del usuario: escribí el test antes del
código.

## 7. Los logs son tu mejor amigo en producción

Logs estructurados con contexto.
Buscables por ID de transacción.

**Regla:** Diseñá logs como si fueran
una API para tu "yo" del futuro.

## 8. El primer paso para arreglar un bug es reproducirlo

**Proceso:**

1. Observá
2. Hipótesis
3. Reproducí
4. Test que falle
5. Arreglá
6. Verificá

## 9. Observá tu código en producción

Logs estructurados, métricas agregadas,
trazas distribuidas.

**Test:** Si hay un problema ahora, ¿cuánto
tardás en enterarte y entender por qué?

## 10. El rendimiento se degrada, monitorizá

Presupuestos de rendimiento (performance
budgets) automáticos en el CI.

**Cultura:**

1. Budgets en CI
2. Dashboards visibles
3. Perf reviews periódicas
4. Alertas proactivas
