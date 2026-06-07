---
description: Lee una historia de usuario (HU) y extrae el requerimiento estructurado.
handoffs:
  - label: Plan y SDD
    agent: go-sdd
    prompt: Genera plan de arquitectura y SDD para este requerimiento
    send: true
---

## Entrada del usuario

```text
$ARGUMENTS
```

Debes considerar la entrada antes de continuar (si no está vacía). Debe incluir al menos uno de:

- Identificador de HU (ej. `HU-1234`) **y/o**
- Ruta a un archivo en el repo (ej. `docs/...`) **y/o**
- Texto pegado de la HU

Si falta el **bounded context** (`internal/<contexto>/` o módulo), pregúntalo antes de asumir rutas.

## Contexto del repo

- Stack: Go, Echo v4, PostgreSQL/sqlx, Clean Architecture (`internal/<contexto>/`, `pkg/`, OpenAPI).
- Alineación obligatoria con `docs/ARCHITECTURE.md` y `docs/SDD.md`; ante conflicto, prevalecen esos documentos.
- Convenciones e índice de skills: `AGENTS.md` y `.cursor/rules/`.

## Objetivo

1. Localizar el contenido de la HU (archivo en el repo, texto en argumentos, o indicaciones para buscar en `docs/`).
2. Extraer y devolver un **requerimiento estructurado** con:
   - Resumen en una frase
   - Alcance funcional
   - Criterios de aceptación (lista verificable)
   - Fuera de alcance (explícito)
   - Supuestos y dependencias
   - Riesgos o ambigüedades → listar como **preguntas abiertas**
3. No implementar código en este paso.
