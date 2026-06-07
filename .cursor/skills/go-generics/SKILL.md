---
name: go-generics
description: >-
  Generics prudentes en pkg/ y helpers compartidos; no sustituyen puertos del
  dominio en docs/ARCHITECTURE.md. Preferir interfaces en bounded contexts.
  Usar al diseñar APIs genéricas o helpers reutilizables.
---

# Go: generics (docsy-main)

## Arquitectura de referencia

- **[`docs/ARCHITECTURE.md`](../../../docs/ARCHITECTURE.md):** el **dominio** se expresa con **entidades e interfaces**; los generics son herramientas de implementación en **`pkg/`** o utilidades compartidas, no una forma de saltarse **domain/ports** por módulo.

**Regla:** usar generics solo si eliminan duplicación clara. Si una interfaz pequeña basta, preferir interfaz.

## Casos típicos

- Utilidades sobre colecciones (`Contains`, `Map`, filtros) con `comparable` o constraints mínimos.
- Contenedores tipados (cache, sets) cuando la lógica es idéntica.
- **No** imponer repositorios genéricos en cada bounded context; el proyecto usa **adaptadores sqlx concretos** salvo necesidad demostrada.

## Constraints

- `comparable` para claves de map y comparaciones en slices.
- `any` con moderación.
- Interfaces de constraint mínimas.

## Limitación de métodos

- No type parameters en métodos de receptor no genérico; usar struct genérico o función libre.

## Resumen

- Generics en `pkg/` o helpers con **2+ usos** reales; sin abstracciones “por si acaso”.

## Skills relacionados

- `go-conventions`
