# Clean Code

> El código se lee mucho más a menudo de lo
> que se escribe. Optimiza para la lectura.

## 1. El código se lee más de lo que se escribe

La fatiga al descifrar lógica críptica se
evita priorizando la legibilidad. El
refactoring es un acto de comunicación.

**Prueba del código legible:**

1. ¿Un junior lo entendería?
2. ¿Lo entenderé en 6 meses?
3. ¿Suena coherente al leerlo en voz alta?

## 2. Programas para personas, no para máquinas

El código "puzzle" se evita tratando el
código como literatura, donde cada función
cuenta una historia clara.

```go
// Mal
function proc(d: number[], t: number)

// Bien
function findAffordableProducts(prices, budget)
```

**Hábitos:**

1. Nombres con sentido
2. Funciones que responden "¿qué hace?"
3. Flujo predecible
4. Cero comentarios "muleta"

## 3. El código limpio no es el objetivo final

La sobreingeniería vuelve el sistema frágil.
Sé pragmático: acepta duplicación si evita
acoplamientos innecesarios.

**Pregunta clave:** "¿Qué pasa si los
requisitos divergen?". Si la respuesta es
"romper la abstracción", no la crees.

## 4. La Regla del Boy Scout

Deja el código mejor de como lo encontraste.

**Acciones "un poco mejor":**

1. Renombra una variable
2. Extrae una función
3. Elimina código muerto
4. Simplifica una expresión

## 5. El mejor error se explica solo

Responde qué pasó, por qué y cómo arreglarlo
en el mensaje de error.

```go
// Mal
Error: invalid input

// Bien
ValidationError: "email" must be a valid
email address. Received: "not-an-email".
Fix: Use "user@example.com".
```

**Tres preguntas del error:**

1. ¿Qué pasó? (concreto)
2. ¿Por qué pasó? (contexto)
3. ¿Cómo lo arreglo? (paso a seguir)

## 6. Solo hay dos cosas difíciles: nombres y caché

Los malos nombres generan coste cognitivo.
Usá patrones que revelen intención.

**Patrones:**

- Funciones: Verbo + Sustantivo
  (`getUserById`, `cancelSubscription`)
- Booleanos: Pregunta sí/no
  (`isActive`, `hasPermission`)
- Colecciones: Plural (`activeUsers`)

## 7. El código nunca miente, los comentarios sí

Refactorizá para que el código sea auto-
explicativo. Comentá solo el "por qué".

**Test:** ¿Puedo mejorar el nombre o extraer
una función para que este comentario sobre?

```go
// Mal: repite el código
counter += 1

// Bien: explica el por qué
time.Sleep(100 * time.Millisecond)
```

## 8. Si no podés explicarlo simple, no lo entendés

Técnica Feynman: explicá como a un niño.

**Síntomas de no entender:**

1. Muchos condicionales
2. Copy-paste
3. Usar "básicamente" al explicar

## 9. Programá como si el mantenedor fuera violento

Código claro, sin sorpresas, nombres que no
requieran decodificador.

**Antes de commit:** Imaginate que sos un
recién llegado con un bug crítico un viernes
a las 3 AM.

## 10. El código que no existe es el que mejor funciona (YAGNI)

Aplicá YAGNI y borrá código muerto
activamente.

**Filtro de Atwood:**

1. ¿Realmente lo necesito?
2. ¿Ya existe una solución?
3. ¿Puedo borrar código en su lugar?

## 11. Hacé que funcione, hacelo bien, hacelo rápido

Red-Green-Refactor. No optimices antes de
que funcione.

1. Funciona (tests pasan)
2. Bien (limpio y legible)
3. Rápido (solo si es necesario)

## 12. Una función hace una sola cosa (SRP)

Cada función en un único nivel de
abstracción (metáfora del periódico).

**Regla del AND:** Si describís tu función
con "Y", hace demasiado.

```go
// Mal: hace 4 cosas
func processOrder(order Order) {
    validate(order)
    saveToDB(order)
    sendEmail(order)
    generateInvoice(order)
}

// Bien: una cosa a la vez
func processOrder(order Order) {
    err := validate(order)
    if err != nil { return err }
    err = saveToDB(order)
    if err != nil { return err }
    return notifyCustomer(order)
}
```

## 13. Código corto no siempre es mejor

Optimizá para comprensión humana, no para
ahorrar líneas.

**Test:** ¿Lo entendería alguien recién
llegado al equipo?

## 14. Explícito es mejor que implícito

Pasá parámetros claros. Evitá estado global
e importaciones oscuras.

## 15. Los nombres revelan intención

Responden al por qué existe, qué hace y
cómo se usa.

**Detector de diseño:** Si no encontrás un
buen nombre, la función hace demasiado.

## 16. Dividí el problema hasta que sea trivial

Descomposición top-down hasta funciones
pequeñas.

**Test de atasco:** ¿Qué es lo más chico
que puedo resolver ahora?

## 17. Animaciones con propósito, no decoración

Feedback, guiar atención, mostrar relaciones.

**Preguntas:** ¿Qué problema resuelve?
¿El usuario necesita esta info?

## 18. La DX importa tanto como la UX

Errores accionables, documentación útil,
feedback instantáneo.

## 19. Consultá la referencia de markdownlint

Para consistencia en documentación, revisá
`docs/references/markdownlint.md`. Ahí están
las ~60 reglas con descripción y ejemplos.

## 20. La terminal es tu superpoder

Contenedores, automatización, alias.
La terminal es tu sistema de seguridad y
entorno operativo completo.

## 21. Escribí código fácil de borrar, no de extender

El 70% del código se reescribe en 2 años.
Optimizá para que desaparezca.

Diseñá módulos independientes. Preferí props
sobre contextos globales.
