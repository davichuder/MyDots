# Guía del Sistema de Documentación — Spec-Driven Development

> Este archivo es el mapa de la documentación del proyecto.
> **Si sos humano:** leelo una vez y volvé cuando necesites orientarte.
> **Si sos una IA:** empezá por `AI_CONTEXT.md` si existe, o por la herramienta de memoria del proyecto. Este archivo es la referencia estructural, no el inicio de sesión.

---

## 0. Jerarquía de autoridad

Cuando dos documentos se contradicen, gana el de mayor jerarquía. No se resuelve con intuición: se actualiza el documento de mayor nivel y se propaga hacia abajo.

```text
specs.md
  └─ Fuente de verdad primaria. Qué hace el sistema.
     Prevalece sobre design.md, tasks.md y cualquier decisión de implementación.

design.md
  └─ Cómo lo hace. Surge de specs, nunca al revés.
     Prevalece sobre tasks.md y adr/.

proposal.md
  └─ Por qué existe. Visión, alcance, métricas.
     Prevalece sobre roadmap, métricas y decisiones de alcance.

adr/
  └─ Por qué se decidió así. Inmutable una vez aceptado.
     Explica decisiones irreversibles. No reemplaza specs ni design.
```

**Caso especial — conflicto entre `proposal.md` y `specs.md`:**  
Si el alcance definido en `proposal.md` contradice lo especificado en `specs.md`, `specs.md` es la fuente de verdad ejecutable. La discrepancia debe resolverse en uno de los dos documentos y, si la decisión fue difícil de revertir, registrarse en `adr/`. Mientras no se resuelva, manda `specs.md`.

---

## 1. Filosofía

Esta estructura sigue **Spec-Driven Development (SDD)** con ciclos TDD/BDD:

1. Primero se define **qué** hace el sistema (`specs.md`)
2. Luego **cómo** lo hace (`design.md`)
3. Luego **cuándo y quién** lo construye (`tasks.md`)
4. El código emerge de las specs. Nunca al revés.

**Principios no negociables:**

- Un dato vive en un solo lugar. Si el mismo concepto aparece en dos archivos, uno va a mentir.
- La documentación sirve al proyecto. Si cuesta más mantenerla que leerla, sobra.
- Nada manual que pueda ser automático. Trazabilidad y evidencia son artefactos del pipeline, no markdown actualizado a mano.
- Los archivos deben ser autoexplicativos. Ninguna sección requiere contexto externo para ser entendida.
- La complejidad documental sigue a la complejidad del proyecto, no la anticipa.

---

## 2. Tiers de archivos

### Tier 1 — Obligatorios (todo proyecto, desde el día 1)

| Archivo | Qué contiene |
| - | - |
| `proposal.md` | Visión, problema, alcance, roadmap y métricas de éxito |
| `specs.md` | Requisitos funcionales + criterios de aceptación por los 4 caminos |
| `design.md` | Arquitectura técnica, APIs, trade-offs |
| `glossary.md` | Lenguaje ubicuo del dominio |
| `adr/` | Carpeta con Architecture Decision Records |

### Tier 2 — Recomendados (con UI, equipo o camino a producción)

| Archivo | Qué contiene | Cuándo agregarlo |
| - | - | - |
| `AI_CONTEXT.md` | Contexto estable para IAs (solo si no hay herramienta de memoria) | Ver Sección 7 |
| `tasks.md` | Backlog priorizado y estado de tareas | Si no usás issue tracker externo |
| `wireframes/` | Bocetos de layout por pantalla | Antes de `ui-system.md` |
| `ui-system.md` | Tokens, componentes y patrones visuales | Después de wireframes validados |
| `schema.md` | Modelos de datos y contratos entre servicios | Si hay API o DB |
| `tests.md` | Estrategia de testing organizada por los 4 caminos | Antes del primer CI |
| `styleguide.md` | Convenciones de código, linting, commits | Con el equipo |
| `ONBOARDING.md` | Cómo levantar el proyecto en menos de 15 minutos | Con el segundo dev |
| `workflow.md` | CI/CD, pipelines, triggers de deploy | Con automatización |
| `principles/` | Principios de ingeniería separados por tema | Con el equipo |
| `CHANGELOG.md` | Qué cambió entre versiones | Desde v1.0.0 |

### Tier 3 — Opcionales (necesidades específicas)

| Archivo | Cuándo usarlo |
| - | - |
| `runbook.md` | Antes del primer deploy a producción |
| `security.md` | Si el sistema maneja datos sensibles o tiene superficie de ataque real |
| `experiments.md` | Solo en proyectos ML/IA con ciclos de entrenamiento |
| `agents/` | Solo si hay agentes IA con contratos formales (ver Sección 8) |

---

## 3. Orden de creación (precedencia SDD)

```text
1.  proposal.md        ← Visión y problema primero.
2.  glossary.md        ← Términos que surgieron al escribir la propuesta.
3.  specs.md           ← Qué hace el sistema, organizado por los 4 caminos.
4.  adr/ (primeros)    ← Decisiones que surgen al diseñar.
5.  design.md          ← Cómo lo hace.
6.  wireframes/        ← Solo si hay UI. Bocetos antes del sistema de diseño.
7.  ui-system.md       ← Solo si hay UI. Surge de wireframes validados.
8.  schema.md          ← Contratos de datos y APIs.
9.  AI_CONTEXT.md      ← Solo si no hay herramienta de memoria.
10. styleguide.md      ← Antes del primer commit del equipo.
11. principles/        ← Cultura técnica del equipo.
12. tests.md           ← Estrategia organizada por los 4 caminos.
13. ONBOARDING.md      ← Antes del onboarding del segundo dev.
14. workflow.md        ← Con el pipeline de CI/CD.
15. tasks.md           ← Derivado de todo lo anterior.
16. CHANGELOG.md       ← Desde la primera release versionada.
17. runbook.md         ← Antes de producción.
```

---

## 4. Catálogo de archivos

### `proposal.md`

Contiene: visión, problema, alcance, fuera de alcance, roadmap por versiones y métricas de éxito.  
No contiene: detalles de implementación, tareas concretas, contratos técnicos.  
Se actualiza cuando: cambia la visión, el alcance o los stakeholders.

---

### `specs.md`

Contiene: qué hace el sistema, criterios de aceptación y edge cases.  
No contiene: cómo se implementa.  
Se actualiza cuando: cambia el comportamiento esperado del sistema.

Cada feature organiza sus criterios en cuatro caminos: **happy path** (usuario bien, sistema bien), **sad path** (usuario mal, sistema bien), **failure path** (usuario bien, sistema mal) y **chaos path** (ambos mal). Dentro de cada uno pueden existir edge cases. El detalle del framework y cuándo cada camino es obligatorio vive en el propio `specs.md` del proyecto.

Cada criterio lleva un ID (`SPEC-001`) que los tests referencian con un tag (`@spec:SPEC-001`).

**Cuándo partir `specs.md` en `/features/`:** cuando una feature tiene más de cinco criterios propios, requiere más de dos decisiones arquitectónicas independientes, o la desarrolla un sub-equipo. El spec global mantiene los criterios transversales; cada carpeta de feature tiene solo el delta.

---

### `design.md`

Contiene: arquitectura, componentes, flujo de datos, decisiones técnicas, trade-offs, integraciones, seguridad y estrategia de despliegue.  
No contiene: diseño visual ni tokens UI.  
Referencia a otros archivos en lugar de duplicar: si existe `schema.md`, `design.md` apunta a él.

---

### `ui-system.md`

Contiene: tokens de diseño, catálogo de componentes, patrones de navegación y estados visuales.  
No contiene: lógica de backend ni arquitectura.  
Precondición: debe existir al menos un wireframe validado antes de definir el sistema de diseño.

---

### `adr/` — Architecture Decision Records

Un archivo por decisión. Nombre: `ADR-001-nombre-descriptivo.md`.  
No se edita una vez aceptado. Si la decisión cambia, se crea un nuevo ADR con estado `Reemplazado por ADR-XXX`.

```markdown
# ADR-001: [Título]
**Fecha:** YYYY-MM-DD
**Estado:** Propuesto | Aceptado | Reemplazado por ADR-XXX

## Contexto
Qué situación requería tomar esta decisión.

## Opciones consideradas
1. Opción A — qué gana, qué pierde
2. Opción B — qué gana, qué pierde

## Decisión
Qué se eligió y por qué.

## Consecuencias
Qué se gana. Qué se pierde. Qué deuda técnica se acepta.
```

---

### `glossary.md`

Contiene: definiciones de los términos clave del dominio.  
Emerge de la propuesta y se refina al escribir specs.

Formato: `**Término:** definición en una oración. Ejemplo de uso si ayuda.`

---

### `schema.md`

Contiene: modelos de datos y contratos entre servicios.  
Si existe un formato ejecutable (`openapi.yaml`, schema de DB), `schema.md` es un índice legible que apunta a él. El formato ejecutable manda. Nunca duplicar contenido entre ambos.

---

### `tasks.md`

Tier 2. Si usás issue tracker externo (GitHub Issues, Linear, Jira), `tasks.md` no existe o es solo un resumen de milestones.  
Formato: `T-XXX` referenciando `SPEC-XXX` cuando implementa un criterio concreto.

---

### `tests.md`

Contiene: estrategia de testing (qué tipos, cobertura mínima, convenciones, golden files).  
No contiene: reportes de CI, logs ni listas de tests individuales.

La suite se organiza reflejando los cuatro caminos definidos en `specs.md` (happy, sad, failure, chaos) más edge cases. La trazabilidad vive en los tests mismos (`@spec:SPEC-001`) y en los artefactos del pipeline, no en markdown manual.

---

### `principles/`

```text
principles/
├── INDEX.md
├── clean-code.md
├── architecture.md
├── testing.md        ← Incluye el framework de escenarios completo y su razonamiento
├── security.md
├── collaboration.md
└── ai-tools.md
```

**Cap: máximo 7 archivos.** Antes de agregar uno nuevo, evaluá si puede incorporarse a uno existente.

`INDEX.md` describe cada archivo en dos líneas y dice cuándo es relevante leerlo.

`principles/testing.md` es donde vive el razonamiento completo del framework de escenarios: por qué existe la matriz de dos dimensiones, cuándo el chaos path es obligatorio, cómo pensar en edge cases, y cómo la suite de tests debe reflejar los cuatro caminos.

---

### `styleguide.md`

Contiene: convenciones verificables de código, linters, formato de commits, branching.  
Relación con `principles/`: el styleguide son reglas automatizables; principles son el razonamiento. Si una convención puede verificarse con un linter, va en styleguide.

---

### `ONBOARDING.md`

Contiene: todo para que un dev nuevo levante el proyecto en menos de 15 minutos.  
No existe `context.md` separado — es el mismo objetivo.

---

### `workflow.md`

Contiene: CI/CD, etapas del pipeline, triggers de deploy, estrategia de entornos.  
No contiene: procedimientos de emergencia en producción.

---

### `runbook.md`

Contiene: qué hacer cuando algo falla en producción.  
Mínimo: rollback, diagnóstico de errores comunes, dashboards, cadena de escalación.  
Se crea antes del primer deploy, no después del primer incidente.

---

### `CHANGELOG.md`

Formato: [Keep a Changelog](https://keepachangelog.com/). Secciones: `Added`, `Changed`, `Fixed`, `Removed`, `Security`.

---

## 5. Wireframes — Bocetos de layout

### Propósito

Los wireframes documentan únicamente **dónde va cada cosa en pantalla**. Sin color, sin bordes decorativos, sin estilo. La respuesta a "¿cómo se ubica esto?" antes de decidir cómo se ve.

Existen para validar layout antes de invertir tiempo en `ui-system.md`. Una vez validados, el sistema de diseño se construye sobre ellos.

### Cuándo usar markdown y cuándo HTML

**Markdown:** pantalla simple, layout lineal, pocas secciones.  
**HTML + CDN:** columnas, grids, sidebars, o cualquier estructura que en markdown quede ambigua.

### Convenciones para wireframes HTML

- Un archivo `.html` por pantalla. Nombre descriptivo: `dashboard.html`, `checkout.html`.
- CDN único por proyecto: Tailwind CSS o DaisyUI. Elegir uno y ser consistente en todos los archivos.
- **Prohibido:** frameworks de JS (React, Vue, Alpine, htmx, cualquiera). Solo HTML y el CDN de CSS elegido.
- **Prohibido:** colores de contenido, bordes decorativos, sombras, gradientes, imágenes reales.
- **Permitido:** clases de layout y espaciado únicamente (`flex`, `grid`, `gap`, `p-`, `m-`, `w-`, `h-`, `col-span-`).
- Todo texto es placeholder descriptivo: `[Título de la página]`, `[Lista de productos]`, `[Botón primario]`.
- Cajas grises para imágenes o componentes complejos.

### Wireframe markdown (pantalla simple)

```markdown
## Login

[ Logo ]

[ Campo: Email ]
[ Campo: Contraseña ]
[ Botón: Ingresar ]

[ Link: ¿Olvidaste tu contraseña? ]
```

### Wireframe HTML (pantalla con layout)

```html
<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <title>Wireframe — Dashboard</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-white p-4">

  <header class="flex justify-between items-center mb-6 border-b pb-4">
    <div>[Logo]</div>
    <nav class="flex gap-4">
      <span>[Nav item]</span>
      <span>[Nav item]</span>
      <span>[Avatar]</span>
    </nav>
  </header>

  <div class="flex gap-6">

    <aside class="w-48 shrink-0">
      <ul class="flex flex-col gap-2">
        <li>[Sección 1]</li>
        <li>[Sección 2]</li>
        <li>[Sección 3]</li>
      </ul>
    </aside>

    <main class="flex-1">
      <h1 class="text-xl mb-4">[Título de pantalla]</h1>

      <div class="grid grid-cols-3 gap-4 mb-6">
        <div class="bg-gray-100 p-4">[Métrica 1]</div>
        <div class="bg-gray-100 p-4">[Métrica 2]</div>
        <div class="bg-gray-100 p-4">[Métrica 3]</div>
      </div>

      <div class="bg-gray-100 p-4 h-48">[Tabla o lista principal]</div>
    </main>

  </div>

</body>
</html>
```

### `wireframes/README.md`

```markdown
# Wireframes

| Archivo | Pantalla | Estado |
| - | - | - |
| `home.md` | Landing page | Validado |
| `dashboard.html` | Panel principal | En revisión |
| `checkout.html` | Flujo de pago — paso 1 | Borrador |
```

---

## 6. AI_CONTEXT.md — Contexto para IAs

### Cuándo existe este archivo

`AI_CONTEXT.md` solo se crea si el proyecto **no usa una herramienta de memoria para IAs** (como engram de gentle-ai, claude-memory u otra equivalente).

Si el proyecto usa una herramienta de memoria, `AI_CONTEXT.md` no existe. La herramienta de memoria es la fuente de contexto del proyecto. Lo único necesario es documentar dónde vive ese contexto:

```markdown
## Contexto para IAs
Este proyecto usa [nombre de la herramienta] para memoria y contexto.
Documentos de referencia arquitectónica: docs/specs.md, docs/design.md, docs/adr/
```

Esta nota puede vivir en `README.md` o al inicio de `DOCUMENTATION.md`.

### Contenido cuando existe

**Qué va aquí:** stack, documentos clave en orden de lectura, reglas de negocio críticas, restricciones de implementación, archivos que requieren revisión humana, convenciones.  
**Qué no va aquí:** la tarea de la sesión actual, el estado de un ticket, notas efímeras. Eso se le pasa a la IA directamente en el prompt, no se versiona.

**Reglas:** máximo una página. Es un resumen que referencia, nunca duplica. Se actualiza cuando cambia el stack o las restricciones permanentes, no a diario.

**Protocolo de lectura para IAs según tarea:**

- Implementar una feature: `specs.md` → `design.md` → `schema.md`
- Escribir tests: `specs.md` → `tests.md`
- Decisión arquitectónica: `design.md` → `adr/`
- Cambios de infra: `workflow.md` → `ONBOARDING.md`
- Entender el dominio: `proposal.md` → `glossary.md` → `specs.md`

```markdown
# AI_CONTEXT.md — [Nombre del Proyecto]
**Última actualización:** YYYY-MM-DD

## Stack
- Lenguaje/Runtime:
- Framework:
- Base de datos:
- Infra:

## Documentos clave (leer en este orden)
1. `docs/proposal.md` — visión y alcance
2. `docs/specs.md` — qué hace el sistema (FUENTE DE VERDAD PRIMARIA)
3. `docs/design.md` — cómo lo hace
4. `docs/glossary.md` — términos del dominio
5. `docs/adr/` — decisiones arquitectónicas

## Jerarquía de autoridad
specs.md > design.md > tasks.md

## Reglas de negocio críticas
[3–5 invariantes. No copiar de specs.md: referenciar con "Ver SPEC-XXX"]

## Restricciones de implementación
[Lo que no se puede hacer aunque parezca razonable]

## Archivos que requieren revisión humana antes de merge
[Ej: adr/, schema.md, specs.md]

## Convenciones
- [Ej: Commits en imperativo, máx 50 chars]
- [Ej: Tests antes que implementación]
```

---

## 7. Agentes y Skills — Cuándo crearlos y cuándo no

### Agentes (`agents/AGENTS.md`)

**Creá un agente cuando:**

- Tiene responsabilidad con límites claros que no se solapan con otros módulos.
- Puede tomar decisiones dentro de su dominio sin consultar a otro módulo.
- Tiene ciclo de vida propio: puede iniciarse, ejecutarse y terminar independientemente.
- El mismo contrato podría implementarse de formas distintas sin cambiar a quien lo llama.

**No creés un agente cuando:**

- Es una función o clase con nombre pretencioso.
- Su lógica depende completamente del estado de otro módulo.
- Lo estás creando por anticipación.
- El proyecto tiene menos de dos semanas o es un MVP de validación.

```markdown
## [NombreAgente]
**Rol:** qué responsabilidad tiene en una oración
**Input:** qué recibe
**Output:** qué produce
**Decisiones que toma:** qué puede resolver solo
**Lo que NO decide:** qué escala o delega
**Falla cuando:** condiciones de error esperadas
```

### Skills (`agents/skills/nombre-skill/SKILL.md`)

**Creá una skill cuando:**

- La misma operación es invocada por más de un agente.
- Es atómica: recibe inputs, produce output, no guarda estado.
- Puede testearse de forma aislada sin depender del agente que la llama.

**No creés una skill cuando:**

- Solo la usa un agente. En ese caso es un método interno.
- Tiene estado propio o depende del contexto de quien la llama.
- Es una abstracción prematura.

```markdown
## [nombre-skill]
**Propósito:** qué hace en una oración
**Input:** tipo y descripción de cada parámetro
**Output:** tipo y descripción del resultado
**Errores posibles:** qué puede fallar y cómo lo comunica
**Usada por:** lista de agentes que la invocan
**Ejemplo:**
  Input: { param1: valor, param2: valor }
  Output: { resultado: valor }
```

---

## 8. Estructura de carpetas

```text
/
├── README.md
├── CHANGELOG.md
├── CONTRIBUTING.md
│
├── docs/
│   ├── DOCUMENTATION.md
│   ├── AI_CONTEXT.md           ← Solo si no hay herramienta de memoria
│   │
│   ├── proposal.md
│   ├── glossary.md
│   ├── specs.md
│   ├── design.md
│   ├── schema.md
│   ├── tests.md
│   ├── styleguide.md
│   ├── ONBOARDING.md
│   ├── workflow.md
│   ├── tasks.md                ← Solo si no hay issue tracker externo
│   ├── runbook.md              ← Antes de producción
│   │
│   ├── wireframes/             ← Solo si hay UI
│   │   ├── README.md
│   │   ├── login.md
│   │   └── dashboard.html
│   │
│   ├── ui-system.md            ← Solo si hay UI, después de wireframes
│   │
│   ├── adr/
│   │   ├── ADR-001-nombre.md
│   │   └── ADR-002-nombre.md
│   │
│   ├── principles/
│   │   ├── INDEX.md
│   │   ├── clean-code.md
│   │   ├── architecture.md
│   │   ├── testing.md
│   │   ├── security.md
│   │   ├── collaboration.md
│   │   └── ai-tools.md
│   │
│   ├── features/               ← Cuando una feature supera 5 criterios de aceptación
│   │   └── nombre-feature/
│   │       ├── spec.md         ← Solo el delta; no repetir lo que está en specs.md
│   │       └── design.md
│   │
│   └── agents/                 ← Solo si hay agentes con contratos formales
│       ├── AGENTS.md
│       └── skills/
│           └── nombre-skill/
│               └── SKILL.md
│
├── src/
├── testdata/
│   └── golden/
└── ci/
    └── pipeline.yml
```

---

## 9. Anti-patrones

**Documentar por anticipación**  
Los archivos se crean cuando el dolor es real y medible.

**Metadata manual por archivo**  
Git registra autoría, fecha y versión. Si necesitás estado: `**Estado:** draft | stable | deprecated`.

**Trazabilidad manual en markdown**  
La trazabilidad real vive en los tests (`@spec:SPEC-001`) y en el pipeline, no en archivos mantenidos a mano.

**Slugs numéricos para features**  
`/001-sms` implica orden global que no existe en desarrollo paralelo. Usar nombres semánticos.

**`tasks.md` antes de specs y design**  
Las tareas son derivadas. Crearlas antes garantiza retrabajo.

**Nombres de archivo que difieren solo en mayúsculas**  
Colisiones en macOS (APFS, case-insensitive por defecto) y Windows.

**`context.md` y `ONBOARDING.md` coexistiendo**  
Son el mismo objetivo. Solo existe `ONBOARDING.md`.

**`AI_CONTEXT.md` cuando ya hay herramienta de memoria**  
Si engram, claude-memory u otra herramienta gestiona el contexto, `AI_CONTEXT.md` es ruido duplicado.

**Especificar solo el happy path**  
Un spec que solo documenta el flujo exitoso es incompleto. Sad, failure y chaos paths son parte del contrato del sistema.

**Chaos path en todas las features**  
Solo es obligatorio en features críticas (dinero, auth, datos personales, servicios externos). En el resto es overhead.

**Wireframes con estilos**  
Un wireframe con colores o tipografía real ya no es un wireframe: es un mockup que cuesta más de hacer y más de cambiar.

**JS en wireframes**  
Los wireframes son estáticos. Cualquier interactividad en esta etapa es sobreingeniería de boceto.

**`ui-system.md` sin wireframes previos**  
Diseñar un sistema de componentes sin haber validado el layout es trabajar en la dirección equivocada.

**`principles/` con más de 7 archivos**  
Sin límite se convierte en una wiki que nadie lee.

---

## 10. Versión lite — Proyectos de menos de 2 semanas

```text
docs/
├── proposal.md
├── specs.md
├── design.md
└── tasks.md
```

Si hay UI, agregar `wireframes/` antes de cualquier cosa visual.  
`AI_CONTEXT.md` solo si no hay herramienta de memoria.  
Todo lo demás se agrega cuando el dolor sea real y medible.

---

## 11. Checklist de merge

- [ ] ¿Los criterios de aceptación en `specs.md` cubren los 4 caminos para esta feature?
- [ ] ¿Si cambió la arquitectura, está reflejado en `design.md`?
- [ ] ¿Si se tomó una decisión difícil de revertir, existe el ADR en `adr/`?
- [ ] ¿El glosario tiene los nuevos términos del dominio?
- [ ] ¿El contexto del proyecto está actualizado (herramienta de memoria o `AI_CONTEXT.md`)?
- [ ] ¿Los tests cubren los 4 caminos y referencian los IDs de specs (`@spec:SPEC-XXX`)?
- [ ] ¿`CHANGELOG.md` tiene la entrada de esta versión? (solo si es release)
- [ ] ¿Si se modificó una pantalla, el wireframe correspondiente refleja el cambio?

---

## 12. Glosario de este sistema

| Término | Significado en este contexto |
| - | - |
| **Fuente de verdad** | El archivo que manda cuando hay contradicción. Solo uno por concepto. |
| **ADR** | Architecture Decision Record. Registro inmutable de una decisión arquitectónica. |
| **Slug** | Nombre descriptivo sin espacios ni números. Ej: `sms-reminders`. |
| **Tier** | Nivel de obligatoriedad de un archivo según el tipo y tamaño del proyecto. |
| **SDD** | Spec-Driven Development. El código emerge de las specs, no al revés. |
| **Trazabilidad** | Capacidad de seguir el hilo desde un requisito hasta su test. Debe ser automática. |
| **Happy path** | Flujo donde el usuario actúa correctamente y el sistema funciona bien. |
| **Sad path** | Flujo donde el usuario comete errores y el sistema responde correctamente. |
| **Failure path** | Flujo donde el usuario actúa bien pero el sistema o entorno falla. |
| **Chaos path** | Flujo donde tanto el usuario como el sistema/entorno fallan simultáneamente. |
| **Edge case** | Situación rara, límite o poco común dentro de cualquiera de los 4 caminos. |
| **Agente** | Módulo autónomo con responsabilidad propia, contrato definido y lógica de decisión independiente. |
| **Skill** | Capacidad reutilizable, atómica y sin estado propio que uno o más agentes pueden invocar. |
| **Delta** | En `features/`: solo los criterios específicos no cubiertos por los archivos globales. |
| **Wireframe** | Boceto de layout. Solo posicionamiento, sin estilo ni color. |
| **Herramienta de memoria** | Sistema externo (engram, claude-memory, etc.) que gestiona el contexto del proyecto para IAs. Reemplaza a `AI_CONTEXT.md`. |
