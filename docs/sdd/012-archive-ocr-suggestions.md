# SDD 012 — OCR local (Tesseract) con sugerencia de campos

**Estado:** Implementación MVP  
**Módulo:** `internal/archive/`  
**Fecha:** 2026-08-08

## 1. Objetivo

Permitir que el usuario, al crear o editar un documento del archivo personal, **analice un adjunto con OCR** y reciba **sugerencias de metadatos** (título, emisor, fechas, monto, referencia, notas) que debe **confirmar o corregir** antes de guardar.

## 2. Alcance

### Incluye

- Motor **Tesseract CLI** (sin CGO), invocado desde `internal/archive/infrastructure/ocr/`.
- PDF: capa de texto con `pdftotext`; si no hay texto útil, OCR de la primera página vía `pdftoppm` + Tesseract.
- Imágenes: JPG/JPEG, PNG, TIFF, WebP, GIF.
- Endpoint web `POST /archivo/documentos/ocr-sugerir` (`archive.write`).
- UI: botón “Analizar con OCR” en el formulario (create y edit); rellena solo campos vacíos en edit.
- Config: `OCR_ENABLED`, `OCR_TESSERACT_BIN`, `OCR_LANG`, `OCR_TIMEOUT_SECS`, bins opcionales Poppler.
- Auditoría best-effort: `archive.ocr_suggested`.

### Excluye

- Cola async / workers.
- OCR de todas las páginas PDF.
- Autocompletado sin confirmación del usuario.
- DIAN / CUFE automático.
- Persistencia del texto OCR crudo en BD.
- Extracción a `pkg/ocr` (queda en el bounded context archive).

### Datos adicionales (extras)

Además de los campos fijos del formulario, el OCR puede sugerir pares `{key,label,value}` (contrato, factura electrónica, cupo Brilla, etc.). Se muestran como **badges eliminables** y se persisten en `archive_documents.extra_fields` (JSONB), sin mezclarlos con `notes`.

## 3. Arquitectura

```text
HTTP → ArchivePageHandler.SuggestOCR
     → SuggestDocumentFieldsUseCase
          → OCRTextExtractor (puerto)
               → TesseractExtractor (adapter)
          → parse heurístico ES/COP → OCRSuggestionResponse
```

**No** se coloca en `pkg/`: es un adaptador del dominio archive (mismo criterio que `DocumentStorage`).

## 4. Dependencias del host

```bash
sudo apt install tesseract-ocr tesseract-ocr-spa poppler-utils
```

Si `OCR_ENABLED=false` o falta el binario, la app no falla: el endpoint responde error claro y el create/upload siguen operativos.

## 5. Criterios de aceptación

1. En “Nuevo documento”, el usuario puede analizar el archivo y ver sugerencias editables.
2. Crear documento sigue exigiendo submit explícito + archivo.
3. Sin Tesseract la app no rompe.
4. Tests del parser y del use case (extractor mockeado) verdes.
