# Guía: Release 1.5 - Ecosistema de Parsers y Tipos de Archivos

## Objetivo
Permitir que el sistema ingiera el 90% de los formatos ofimáticos estándar: PDF, DOCX, XLSX, CSV, JSON, TXT.

## Decisiones Aplicadas

| Fuente | Decisión |
|--------|----------|
| Architecture 7.1 | Strategy Pattern con ParserRegistry |
| Architecture 7.2 | CSV/JSON devuelven []schema.Document directo (sin chunker) |
| Architecture 7.3 | PDF texto nativo + OCRmyPDF fallback |
| Architecture 7.4 | XLSX fila por fila con metadata de headers |
| Architecture 7.5 | DOCX extracción directa con librería Go |
| Architecture 7.6 | PDFs página por página con page_number metadata |
| Architecture 7.7 | Imágenes embebidas: placeholder `[IMAGEN: name]` |
| Architecture 7.8 | OCRmyPDF como Docker service en homelab |

---

## Estado Actual (Phase 1.4)

El worker procesa documentos con:
```
download → parse → chunk → embed → upsert
```

Solo existe `PlainTextParser`. El chunker semántico ya funciona para texto plano.

---

## Paso 1: Puerto ParserRegistry

Crea `internal/core/ports/parser.go`:

```go
type Parser interface {
    Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error)
}

type ParserRegistry interface {
    Register(contentType string, parser Parser)
    Get(contentType string) (Parser, error)
}
```

Nota: `Parse` retorna `[]schema.Document` directamente. Los parsers estructurados (CSV, JSON) devuelven documentos ya listos. Los parsers de texto (PDF, DOCX, TXT) devuelven documentos por página/párrafo que pasan por el chunker.

---

## Paso 2: Implementación ParserRegistry

Crea `internal/infrastructure/parser/registry.go`:

```go
type ParserRegistryImpl struct {
    parsers map[string]ports.Parser
}

func (r *ParserRegistryImpl) Register(contentType string, parser ports.Parser) {
    r.parsers[contentType] = parser
}

func (r *ParserRegistryImpl) Get(contentType string) (ports.Parser, error) {
    p, ok := r.parsers[contentType]
    if !ok {
        return nil, fmt.Errorf("unsupported content type: %s", contentType)
    }
    return p, nil
}
```

Registration en `main.go`:
```go
registry := parser.NewParserRegistry()
registry.Register("text/plain", parser.NewPlainTextParser())
registry.Register("text/csv", parser.NewCSVParser())
registry.Register("application/json", parser.NewJSONParser())
registry.Register("application/pdf", parser.NewPDFParser(ocrEndpoint))
registry.Register("application/vnd.openxmlformats-officedocument.wordprocessingml.document", parser.NewDOCXParser())
registry.Register("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", parser.NewXLSXParser())
```

---

## Paso 3: PlainText Parser (ya existe, ajustar firma)

Actualizar `internal/infrastructure/parser/plaintext.go`:

```go
func (p *PlainTextParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
    content, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }
    // Devuelve un solo documento que pasará por el chunker
    return []schema.Document{{
        PageContent: string(content),
        Metadata:    map[string]any{"content_type": contentType},
    }}, nil
}
```

---

## Paso 4: CSV Parser

Crea `internal/infrastructure/parser/csv.go`:

```go
func (p *CSVParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
    csvReader := csv.NewReader(reader)
    headers, err := csvReader.Read()
    if err != nil {
        return nil, err
    }

    var docs []schema.Document
    rowNum := 1
    for {
        record, err := csvReader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            rowNum++
            continue
        }

        // Construir texto: "header1: value1, header2: value2, ..."
        var parts []string
        for i, h := range headers {
            if i < len(record) {
                parts = append(parts, fmt.Sprintf("%s: %s", h, record[i]))
            }
        }

        docs = append(docs, schema.Document{
            PageContent: strings.Join(parts, ", "),
            Metadata: map[string]any{
                "row":        rowNum,
                "headers":    strings.Join(headers, "|"),
                "chunk_type": "structured", // flag para saltar chunker semántico
            },
        })
        rowNum++
    }
    return docs, nil
}
```

Metadata `chunk_type: "structured"` indica que este documento ya está listo y no necesita chunking.

---

## Paso 5: JSON Parser

Crea `internal/infrastructure/parser/json.go`:

```go
func (p *JSONParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
    var data interface{}
    if err := json.NewDecoder(reader).Decode(&data); err != nil {
        return nil, err
    }

    var docs []schema.Document

    // Si es array, cada elemento es un documento
    if arr, ok := data.([]interface{}); ok {
        for i, item := range arr {
            docs = append(docs, schema.Document{
                PageContent: flattenJSON(item),
                Metadata: map[string]any{
                    "index":      i,
                    "chunk_type": "structured",
                },
            })
        }
    } else {
        // Si es objeto, un solo documento
        docs = append(docs, schema.Document{
            PageContent: flattenJSON(data),
            Metadata: map[string]any{
                "chunk_type": "structured",
            },
        })
    }
    return docs, nil
}

func flattenJSON(v interface{}) string {
    // Recursivamente convierte JSON a "key: value, key: value" formato
    // ...
}
```

---

## Paso 6: PDF Parser (con OCR fallback)

Crea `internal/infrastructure/parser/pdf.go`:

```go
type PDFParser struct {
    ocrEndpoint string // URL del servicio OCRmyPDF
}

func (p *PDFParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
    // 1. Guardar temporalmente
    tmpDir, err := os.MkdirTemp("", "pdf-parse-*")
    if err != nil {
        return nil, err
    }
    defer os.RemoveAll(tmpDir)

    inputPath := filepath.Join(tmpDir, "input.pdf")
    f, err := os.Create(inputPath)
    if err != nil {
        return nil, err
    }
    io.Copy(f, reader)
    f.Close()

    // 2. Intentar extraer texto nativo
    nativeText, err := extractPDFText(inputPath) // usando pdfcpu o similar
    if err == nil && len(strings.TrimSpace(nativeText)) > 0 {
        return p.splitByPages(nativeText), nil
    }

    // 3. Fallback: OCR
    if p.ocrEndpoint == "" {
        return nil, fmt.Errorf("PDF requires OCR but no OCR endpoint configured")
    }

    ocrOutputPath := filepath.Join(tmpDir, "ocr_output.pdf")
    if err := runOCR(ctx, inputPath, ocrOutputPath, p.ocrEndpoint); err != nil {
        return nil, fmt.Errorf("OCR failed: %w", err)
    }

    ocrText, err := extractPDFText(ocrOutputPath)
    if err != nil {
        return nil, err
    }
    return p.splitByPages(ocrText), nil
}

func (p *PDFParser) splitByPages(text string) []schema.Document {
    pages := strings.Split(text, "\x0c") // Form feed character = page separator
    var docs []schema.Document
    for i, page := range pages {
        if strings.TrimSpace(page) == "" {
            continue
        }
        // Reemplazar imágenes con placeholder
        page = replaceImagePlaceholders(page)
        docs = append(docs, schema.Document{
            PageContent: page,
            Metadata: map[string]any{
                "page_number": i + 1,
                "page_count":  len(pages),
            },
        })
    }
    return docs
}
```

OCR service call:
```go
func runOCR(ctx context.Context, inputPath, outputPath string, endpoint string) error {
    cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
        "-v", fmt.Sprintf("%s:/tmp", filepath.Dir(inputPath)),
        "jbarlow83/ocrmypdf",
        "--skip-text",
        filepath.Base(inputPath),
        filepath.Base(outputPath),
    )
    return cmd.Run()
}
```

---

## Paso 7: DOCX Parser

Crea `internal/infrastructure/parser/docx.go`:

Usar `github.com/unidoc/unioffice/document` o similar.

```go
func (p *DOCXParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
    tmpDir, err := os.MkdirTemp("", "docx-parse-*")
    if err != nil {
        return nil, err
    }
    defer os.RemoveAll(tmpDir)

    filePath := filepath.Join(tmpDir, "input.docx")
    f, err := os.Create(filePath)
    if err != nil {
        return nil, err
    }
    io.Copy(f, reader)
    f.Close()

    doc, err := document.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer doc.Close()

    var paragraphs []string
    for _, p := range doc.Paragraphs() {
        for _, r := range r.Runs() {
            paragraphs = append(paragraphs, r.Text())
        }
    }

    // Si hay imágenes embebidas, agregar placeholders
    // ...

    content := strings.Join(paragraphs, "\n\n")
    return []schema.Document{{
        PageContent: content,
        Metadata:    map[string]any{"content_type": contentType},
    }}, nil
}
```

---

## Paso 8: XLSX Parser

Crea `internal/infrastructure/parser/xlsx.go`:

Similar al CSV parser pero iterando por hojas.

```go
func (p *XLSXParser) Parse(ctx context.Context, reader io.Reader, contentType string) ([]schema.Document, error) {
    tmpDir, err := os.MkdirTemp("", "xlsx-parse-*")
    if err != nil {
        return nil, err
    }
    defer os.RemoveAll(tmpDir)

    filePath := filepath.Join(tmpDir, "input.xlsx")
    f, err := os.Create(filePath)
    if err != nil {
        return nil, err
    }
    io.Copy(f, reader)
    f.Close()

    xlFile, err := xlsx.OpenFile(filePath)
    if err != nil {
        return nil, err
    }

    var docs []schema.Document
    for _, sheet := range xlFile.Sheets {
        for i, row := range sheet.Rows {
            if i == 0 {
                continue // skip header row
            }
            headers := sheet.Rows[0]
            var parts []string
            for j, cell := range row.Cells {
                if j < len(headers) {
                    parts = append(parts, fmt.Sprintf("%s: %s", headers[j].Value, cell.Value))
                }
            }
            docs = append(docs, schema.Document{
                PageContent: strings.Join(parts, ", "),
                Metadata: map[string]any{
                    "sheet":      sheet.Name,
                    "row":        i + 1,
                    "chunk_type": "structured",
                },
            })
        }
    }
    return docs, nil
}
```

---

## Paso 9: Adaptar Worker para manejar structured docs

El worker necesita saber si un documento es "structured" (CSV, JSON) y saltarse el chunker:

En `ingest_worker.go`, método `processDocument`:

```go
// Parse
parsedDocs, err := w.Parser.Parse(ctx, reader, doc.ContentType)
if err != nil {
    return w.failDocument(doc, err)
}

for _, parsedDoc := range parsedDocs {
    // Check if structured (skip chunker)
    if parsedDoc.Metadata["chunk_type"] == "structured" {
        // Direct to embedding
        if err := w.processChunk(ctx, doc, parsedDoc.PageContent, parsedDoc.Metadata); err != nil {
            return err
        }
    } else {
        // Normal: pass through semantic chunker
        chunks, err := w.Chunker.Chunk(parsedDoc.PageContent)
        if err != nil {
            return err
        }
        for _, chunk := range chunks {
            if err := w.processChunk(ctx, doc, chunk, parsedDoc.Metadata); err != nil {
                return err
            }
        }
    }
}
```

---

## Paso 10: OCR Docker Service

Agregar a `docker-compose.yml`:

```yaml
ocrmypdf:
  image: jbarlow83/ocrmypdf
  volumes:
    - ./tmp/ocr:/tmp
  deploy:
    resources:
      limits:
        memory: 2G
```

Agregar env var en `.env`:
```
OCR_ENABLED=true
OCR_ENDPOINT=docker://jbarlow83/ocrmypdf
```

Config en `config.go`:
```go
OCREnabled   bool
OCREndpoint  string
```

---

## Paso 11: Extender validación de extensiones

Actualizar el handler de upload para aceptar nuevas extensiones:

```go
var allowedExtensions = map[string]bool{
    ".txt":  true,
    ".csv":  true,
    ".json": true,
    ".pdf":  true,
    ".docx": true,
    ".xlsx": true,
}
```

---

## Estructura de archivos nuevos

```
internal/
├── core/
│   └── ports/
│       └── parser.go          ← actualizar con nueva firma
├── infrastructure/
│   └── parser/
│       ├── registry.go        ← nuevo
│       ├── plaintext.go       ← actualizar firma
│       ├── csv.go             ← nuevo
│       ├── json.go            ← nuevo
│       ├── pdf.go             ← nuevo
│       ├── docx.go            ← nuevo
│       └── xlsx.go            ← nuevo
└── worker/
    └── ingest_worker.go       ← actualizar para structured docs
```

---

## Dependencias nuevas

```
go get github.com/unidoc/unioffice/document
go get github.com/tealeg/xlsx
go get github.com/pdfcpu/pdfcpu/pkg/api
```

---

## Orden sugerido

```
1. Actualizar puerto Parser (nueva firma: retorna []schema.Document)
2. ParserRegistry
3. PlainTextParser (actualizar firma)
4. CSV Parser
5. JSON Parser
6. DOCX Parser
7. XLSX Parser
8. PDF Parser (texto nativo)
9. OCR integration (PDF fallback)
10. Adaptar worker para structured docs
11. Extender validación de extensiones
12. Docker compose: agregar ocrmypdf service
13. Config: agregar vars de OCR
14. main.go: registrar todos los parsers
15. Probar: subir cada tipo de archivo y verificar ingestión
```
