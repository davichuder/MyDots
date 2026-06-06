# markdownlint — Reference

> [markdownlint](https://github.com/DavidAnson/markdownlint)
> es un linter para archivos
> Markdown/CommonMark. Esta referencia
> lista las ~60 reglas integradas agrupadas
> por tag.

## Alcance

El proyecto MyDots aplica markdownlint
sobre:

- Toda la documentación en `docs/`
- README y `.md` en la raíz del repo
- Cheatsheets en `assets/cheatsheets/`

## Cómo usar

```bash
npx markdownlint-cli2 "docs/**/*.md" "*.md"
```

Config (`./markdownlint.json`):

```json
{
  "default": true,
  "MD013": { "line_length": 80 },
  "MD033": { "allowed_elements": ["br"] }
}
```

## Inline control

```markdown
<!-- markdownlint-disable MD013 -->
...
<!-- markdownlint-enable MD013 -->

<!-- markdownlint-disable-next-line MD033 -->
<br/>
```

---

## Headings (14 rules)

| Regla | Alias | Qué comprueba |
| ----- | ----- | ------------- |
| MD001 | `heading-increment` | Saltar niveles de heading |
| MD003 | `heading-style` | Consistencia de estilo |
| MD018 | `no-missing-space-atx` | Sin espacio tras `#` |
| MD019 | `no-multiple-space-atx` | Múltiples espacios tras `#` |
| MD020 | `no-missing-space-closed-atx` | Sin espacio en `#...#` |
| MD021 | `no-multiple-space-closed-atx` | Múltiples espacios en `#...#` |
| MD022 | `blanks-around-headings` | Blank lines faltantes |
| MD023 | `heading-start-left` | Heading no empieza al inicio |
| MD024 | `no-duplicate-heading` | Heading duplicado |
| MD025 | `single-title/h1` | Múltiples h1 |
| MD026 | `no-trailing-punctuation` | Puntuación al final |
| MD036 | `no-emphasis-as-heading` | Énfasis como heading |
| MD041 | `first-line-h1` | Sin h1 al inicio |
| MD043 | `required-headings` | Estructura obligatoria |

---

## Code (6 rules)

| Regla | Alias | Qué comprueba |
| ----- | ----- | ------------- |
| MD014 | `commands-show-output` | `$` sin mostrar output |
| MD031 | `blanks-around-fences` | Fences sin blank lines |
| MD038 | `no-space-in-code` | Espacios en code spans |
| MD040 | `fenced-code-language` | Lenguaje no especificado |
| MD046 | `code-block-style` | Estilo inconsistente |
| MD048 | `code-fence-style` | Símbolo de fence inconsistente |

---

## Lists (6 rules)

| Regla | Alias | Qué comprueba |
| ----- | ----- | ------------- |
| MD004 | `ul-style` | Símbolo inconsistente |
| MD005 | `list-indent` | Indentación inconsistente |
| MD007 | `ul-indent` | Indentación incorrecta |
| MD029 | `ol-prefix` | Prefijo de lista ordenada |
| MD030 | `list-marker-space` | Espacios tras marker |
| MD032 | `blanks-around-lists` | Lists sin blank lines |

---

## Links (9 rules)

| Regla | Alias | Qué comprueba |
| ----- | ----- | ------------- |
| MD011 | `no-reversed-links` | Link invertido |
| MD034 | `no-bare-urls` | URL sin `<>` |
| MD039 | `no-space-in-links` | Espacios en link text |
| MD042 | `no-empty-links` | Link vacío |
| MD051 | `link-fragments` | Fragmento inválido |
| MD052 | `reference-links-images` | Label sin definir |
| MD053 | `link-image-ref-defs` | Definition sin usar |
| MD054 | `link-image-style` | Estilo inconsistente |
| MD059 | `descriptive-link-text` | Texto genérico ("click here") |

---

## Emphasis (3 rules)

| Regla | Alias | Qué comprueba |
| ----- | ----- | ------------- |
| MD037 | `no-space-in-emphasis` | Espacios en `** **` |
| MD049 | `emphasis-style` | Estilo inconsistente |
| MD050 | `strong-style` | Estilo inconsistente |

---

## Whitespace (5 rules)

| Regla | Alias | Qué comprueba |
| ----- | ----- | ------------- |
| MD009 | `no-trailing-spaces` | Espacios al final |
| MD010 | `no-hard-tabs` | Tabs en vez de espacios |
| MD012 | `no-multiple-blanks` | Blank lines de más |
| MD027 | `no-multiple-space-blockquote` | Espacios tras `>` |
| MD028 | `no-blanks-blockquote` | Blank line en blockquote |

---

## Tables (4 rules)

| Regla | Alias | Qué comprueba |
| ----- | ----- | ------------- |
| MD055 | `table-pipe-style` | Pipes leading/trailing |
| MD056 | `table-column-count` | Columnas desiguales |
| MD058 | `blanks-around-tables` | Tablas sin blank lines |
| MD060 | `table-column-style` | Padding inconsistente |

---

## Other (6 rules)

| Regla | Alias | Qué comprueba |
| ----- | ----- | ------------- |
| MD013 | `line-length` | Línea > N chars (def 80) |
| MD033 | `no-inline-html` | HTML inline |
| MD035 | `hr-style` | HR inconsistente |
| MD044 | `proper-names` | Capitalización incorrecta |
| MD045 | `no-alt-text` | Imagen sin alt text |
| MD047 | `single-trailing-newline` | Sin newline al final |

---

## Tags por área

```text
accessibility   → MD045, MD059
blank_lines     → MD012, MD022, MD031, MD032, MD047
blockquote      → MD027, MD028
bullet          → MD004, MD005, MD007, MD032
code            → MD014, MD031, MD038, MD040, MD046, MD048
emphasis        → MD036, MD037, MD049, MD050
headings        → MD001, MD003, MD018-MD026, MD036, MD041, MD043
hr              → MD035
html            → MD033
images          → MD045, MD052-MD054
indentation     → MD005, MD007, MD027
line_length     → MD013
links           → MD011, MD034, MD039, MD042, MD051-MD054, MD059
ol              → MD029, MD030, MD032
spaces          → MD018-MD021, MD023
spelling        → MD044
table           → MD055, MD056, MD058, MD060
ul              → MD004, MD005, MD007, MD030, MD032
url             → MD034
whitespace      → MD009, MD010, MD012, MD027, MD028, MD030,
                  MD037-MD039
```

---

## Links oficiales

- [GitHub](https://github.com/DavidAnson/markdownlint)
- [Demo](https://dlaa.me/markdownlint/)
- [Rules.md](https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md)
