# Architecture

> Diseño de software, abstracciones,
> trade-offs y decisiones técnicas.

## 1. El tipo correcto elimina bugs antes de existir

Diseñá estructuras con tipos que hagan
imposibles los estados ilegales.

```go
type OrderStatus string

const (
    StatusPending  OrderStatus = "pending"
    StatusShipped  OrderStatus = "shipped"
    StatusDelivered OrderStatus = "delivered"
)

type Order struct {
    Status OrderStatus
}
```

**Regla de oro:** Estados inválidos
inexpresables. Uniones discriminadas para
forzar el manejo de cada caso.

## 2. Elegí tecnología aburrida

Tecnologías conocidas. Innovación solo para
el valor core del producto.

**Ley de los Tokens:** Máximo 3 tokens de
innovación por proyecto. De a uno, solo
cuando el dolor sea real.

## 3. APIs fáciles de usar y difíciles de usar mal

Interfaces que guíen al éxito (Pit of
Success) por defecto.

**Regla de Bloch:** Si el usuario puede
hacer algo mal y compila, la API falló.

## 4. Hacé cosas que no escalen

Empezá simple y manual. Escalá solo cuando
el dolor sea real y medible.

**Señales para escalar:**

1. BD lenta
2. Deploys con errores
3. Archivo inmanejable
4. Conflictos entre devs

## 5. Fácil de cambiar siempre (ETC)

Inyección de Dependencias. 4 pilares del
diseño flexible.

**Pilares TRUE:**

- **T**ransparente (cambio obvio)
- **R**azonable (coste proporcional)
- **U**sable (reutilizable)
- **E**jemplar (invita a seguir el diseño)

## 6. Optimización prematura = raíz de todos los males

Identificá el 3% crítico con datos reales.

**Regla 97/3:** El 97% del código no
necesita optimización. Identificá el 3%
con datos, no con intuición.

## 7. La abstracción es clave para la supervivencia

Contratos honestos (interfaces) que oculten
detalles volátiles.

```go
// Mal: filtra implementación
func (e EmailService) SendSMTP(addr, msg string)

// Bien: abstracción
type MessageService interface {
    Send(to, msg string) error
}
```

**Liskov:** Cualquier implementación debe
poder reemplazar a la interfaz sin romper
nada.

## 8. Simplicidad es prerrequisito de fiabilidad

Eliminá complejidad accidental. Separá
reglas en piezas predecibles.

**Hábitos:**

1. Decí "no" a features innecesarias
2. Eliminá antes de añadir
3. Buscá invariantes
4. Desconfiá de la "flexibilidad"

## 9. El mejor request es el que no se hace

Auditá dependencias. Usá capacidades
nativas (lazy loading).

**Regla del 20%:** Sé 20% más rápido que
tu competidor. Medí con datos reales.

## 10. La deuda técnica es como la financiera

Hacela visible. Pagá intereses con
refactorización continua.

**Estrategia:** Endeudate si el coste de no
entregar supera al de refactorizar. Tené
un plan de pago.

## 11. Todas las abstracciones tienen fugas

Aprendé al menos una capa por debajo de tus
herramientas.

**Ejemplos:** El event loop si usás JS.
Los contenedores si usás Kubernetes.

## 12. Convención sobre configuración

Frameworks con caminos por defecto que
funcionen para el 90% de los casos.

## 13. No diseñes para el peor caso

Resolvé el problema actual. Planificá solo
cambios probables.

**Diseño pragmático:**

1. Escala apropiada
2. BD apropiada
3. Arquitectura para el equipo actual

## 14. Entendé la red: cada milisegundo importa

Resource hints (preconnect, preload).
Menos viajes al servidor.

**CDNs:** Acercá datos al usuario.
La física es innegociable.

## 15. Si tu sistema es bueno, desplegá en viernes

**Requisitos:**

1. Deploys atómicos
2. Rollback inmediato
3. Feature flags
4. Observabilidad
5. Tests en CI

## 16. Menos JavaScript es mejor JavaScript

Capacidades nativas (HTML/CSS). Solo lo
interactivo.

**Coste:** Cada byte de JS cuesta descarga,
parsing y memoria en el móvil.

## 17. Estructuras de datos dominan sobre algoritmos

Modelá los datos para que la lógica sea
trivial.

**Pregunta clave:** ¿Hay una forma de
organizar mis datos que haga el código
innecesario?

## 18. Composición sobre herencia

Combiná unidades independientes. Evitá
jerarquías rígidas.

```go
type Penguin struct {
    Eater
    Swimmer
    Sleeper
}
```

## 19. CSS nativo es más potente de lo que creés

Nesting, variables, `:has()`, scroll-driven
animations antes que una librería.

**Progresivo:** HTML → CSS nativo → JS
(solo si es vital).

## 20. Duplicá código antes de abstraer mal

Dejá que el patrón emerja orgánicamente.

**Regla del Tres:** Abstraé solo la tercera
vez que se repita exactamente.

## 21. Antes del framework, aprendé el problema

No preguntes cómo se usa. Preguntá qué
problema resuelve.

**Modelo Mental:** ¿Por qué existe el
Virtual DOM? ¿Por qué es difícil
sincronizar estado y UI?

## 22. Concurrencia no es paralelismo

I/O concurrente. Workers para CPU.

**Regla de Go:** No comuniques compartiendo
memoria. Compartí memoria comunicando.

## 23. Los errores son datos, no excepciones

Resultados tipados que el llamador debe
gestionar.

```go
type Result struct {
    Ok   bool
    Data interface{}
    Err  error
}
```

## 24. Estado mutable = raíz de todos los bugs

Contené la mutación. Preferí datos
derivados.

**Estrategia:**

1. Minimizá superficie
2. Encapsulá mutación
3. Cambios predecibles
4. Datos derivados

## 25. El usuario no espera, cada ms cuenta

Skeleton screens. Optimistic UI.

**Umbrales:**

- 0-100ms: Instantáneo
- 100-300ms: Natural
- +1s: Pesado
- >10s: Abandono

## 26. No reinventes la rueda (a menos que estés aprendiendo)

Librerías probadas para tareas comunes.
Evaluá antes de construir.

**Decisión:** ¿Es core? (Hacelo).
¿Existe solución? (Usala).

## 27. Buen diseño = fácil de cambiar

Ante dos enfoques, elegí el más fácil de
revertir en el futuro.

## 28. TypeScript te hace ir más rápido

Autocompletado y refactoring con garantía
del compilador.

**Espejo de diseño:** Si tipar duele, la
API está mal diseñada.

## 29. Empezá por el final: escribí el uso primero

Diseñá desde el consumidor antes de
implementar.

**Hábito:** Antes de crear una API, escribí
3 líneas que la usen.

## 30. Cada dependencia es deuda técnica

Evaluá si podés resolverlo con APIs nativas.

**Preguntas:** ¿Cuánto usaré? ¿Está
mantenido? ¿Cuántas dependencias trae?

## 31. Separalo que cambia de lo que permanece (SRP)

Cada motivo de cambio en su propio módulo.

**Pregunta:** Si cambian reglas de precios,
¿cuántos archivos tocás?

## 32. Contexto lo es todo

Evaluá equipo, fase del producto y coste
de error antes de decidir.

**Regla AHA:** Duplicá código antes de
crear la abstracción incorrecta prematura.

## 33. El software es un proceso, no un destino

Diseñá para la maleabilidad. El deploy es
rutina diaria.

**Verdad:** El software que no evoluciona
muere. Tu trabajo es hacerlo barato de
modificar.

## 34. Automatizá si lo hacés más de tres veces

Calculá el ROI. Recuperá horas de
pensamiento real.

**Cuándo NO:** Tarea que cambia, solo dos
veces más, error catastrófico.

## 35. Camino correcto = camino fácil (Pit of Success)

Defaults seguros. Herramientas en CI.

**Mecanismos:**

1. Defaults seguros
2. Errores por defecto
3. Linters
4. Pre-commit hooks

## 36. No optimices lo que no mediste

Profilers para identificar el cuello de
botella real antes de actuar.

**Ciclo:**

1. Medí el estado actual
2. Identificá el cuello
3. Actuá
4. Verificá la mejora
