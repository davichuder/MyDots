# Collaboration

> El software se construye en equipo.
> La comunicación, la empatía y los
> procesos definen la velocidad real.

## 1. Talk is cheap. Show me the code

Prototipos funcionales que generen
información real.

**Principios:**

1. Discusiones de 15 min máximo
2. Los prototipos matan debates
3. El código es la verdad absoluta

## 2. Ley de Brooks: más gente en proyecto retrasado = más retraso

El onboarding tiene coste. Reducí alcance
en vez de añadir personas.

**Coste de comunicación:** Crece
cuadráticamente con cada persona.

## 3. Tu comunidad es tu superpoder

Compartí, pedí ayuda, construí en público.

**Beneficios:**

1. Red de seguridad
2. Oportunidades
3. Aprendizaje
4. Salud mental

## 4. Cuestioná todo lo que das por hecho

Preguntá "¿por qué?". Experimentá con
alternativas simples.

**Preguntas incómodas:** ¿Seguimos
necesitando microservicios? ¿Esta
abstracción simplifica o solo mueve
complejidad?

## 5. Los mensajes de commit son cartas al futuro

Modo imperativo. Explicá el por qué en el
cuerpo.

```bash
# Mal
git commit -m "fix"

# Bien
git commit -m "fix(auth): revoke all
sessions on expired token access"
```

**Reglas:**

1. Subject ≤ 50 chars
2. Modo imperativo
3. Línea en blanco subject / cuerpo
4. El cuerpo explica el POR QUÉ

## 6. La DX importa tanto como la UX

Errores accionables, documentación útil,
feedback instantáneo.

**Pilares DX:**

1. Tiempo hasta "Hola Mundo"
2. Feedback rápido
3. Errores que ayudan
4. Documentación útil

## 7. Enseñá valentía, no perfección

Trabajo en progreso. Preguntas "tontas".

**Hábitos valientes:**

1. Mostrá trabajo en curso
2. Hacé preguntas obvias
3. Publicá código imperfecto

## 8. La estimación es predicción, no promesa

Rangos probabilísticos. 3 puntos:
Optimista, Nominal, Pesimista.

## 9. Preguntá lo que no sabés (y documentá mientras aprendés)

Admití ignorancia. Documentá públicamente.

**Pasos:**

1. Admití que no sabés
2. Hacé preguntas concretas
3. Documentá mientras aprendés

## 10. Los datos importan, no las opiniones

Métricas reales de calidad, diversidad,
rendimiento y éxito.

**Mecanismo:** Sin medición no hay mejora.
Usá datos para forzar honestidad.

## 11. Aprendé en público: tu inversión profesional

Compartí notas, errores y aprendizajes
mientras ocurren.

**Técnica Feynman:** Explicá como a un
niño para detectar tus huecos.

## 12. Aprendé haciendo, no solo leyendo (20/80)

Proyecto real en los primeros 15 minutos.

**Regla 20/80:** 20% leyendo, 80%
construyendo, rompiendo y arreglando.

## 13. Git es tu máquina del tiempo

Commits atómicos, `git bisect` para bugs,
rebase para historial limpio.

**Flujo:**

1. Rama feature
2. Commits pequeños
3. Rebase para limpiar
4. Merge con historia nítida

## 14. El README es tu primera impresión

Landing page: hook, instalación, ejemplos.

**Checklist:**

1. Hook inmediato
2. Instalación en un paso
3. Ejemplo real
4. Screenshots / Demos

## 15. El código es comunicación entre humanos

Lenguaje del dominio (Ubiquitous Language)
en funciones y variables.

**Code Reviews:** Conversaciones para
alinear vocabulario e intención.

## 16. Code reviews son para aprender, no para juzgar

Preguntas abiertas. Contexto educativo.
Reconocé lo positivo.

**Checklist:** ¿Entendería esto alguien
nuevo? ¿Hay tests? ¿Aprendí algo?

## 17. La documentación es un acto de empatía

Diátaxis: 4 tipos de contenido.

1. Tutoriales (aprender haciendo)
2. How-to (resolver un problema)
3. Referencia (hechos, describir)
4. Explicación (por qué, entender)

## 18. No asumas que el otro sabe, documentá (ADRs)

Decisiones con contexto y consecuencias.

**Qué documentar:** Decisiones difíciles
de revertir, alternativas descartadas,
consecuencias.

## 19. Aprendé a decir "no" o "no todavía"

Cada feature = mantenimiento + tests + doc.

**Costo del Sí:** Todo eso para siempre.

## 20. Cuidá tu salud mental: el burnout es real

Límites de horario. Desconectá de verdad.

**Estrategias:**

1. Límites estrictos
2. Vacaciones reales
3. Hobbies fuera de tech
4. Apoyo profesional

## 21. Antes de cómo, preguntá para qué (5 Porqués)

Llegá a la raíz del problema, no al
síntoma.

## 22. El feedback temprano vale oro

Reducí a segundos el tiempo entre escribir
código y ver resultado (HMR, Previews).

**Mantra:** Ese tiempo define la calidad
de tu trabajo.

## 23. La humildad te hace mejor desarrollador

Admití lo que no sabés. Pedí revisiones con
apertura.

**Impacto:** Código más legible, mejores
tests, decisiones documentadas.

## 24. Pair programming no es perder el tiempo

Driver (detalles, sintaxis) y Navigator
(estrategia, diseño, casos borde).

## 25. Nunca dejés de aprender: la curiosidad es tu motor

Mente de principiante (shoshin). Explorá
fuera de tu zona.

**Regla del 20%:** Tiempo para explorar
sin relación con tu trabajo actual.

## 26. Aprendé haciendo, no leyendo

Salí del "Tutorial Hell". Proyectos reales
desde el día uno. La práctica genera
memoria muscular.
