# Stitch Prompt: Rago File Dashboard

**Product**: Rago — a RAG (Retrieval-Augmented Generation) platform. Users upload documents, the backend parses/chunks/embeds/indexes them into a vector DB (Qdrant), then users can chat/ask questions against their documents.

**Goal**: Design the file/document dashboard page.

---

## User Flow

1. User drags/drops or clicks to select a file
2. File uploads via multipart form (field: `file`) to `POST /api/v1/documents/`
3. Immediate response: `{ id: 123, filename: "report.pdf", status: "pending" }`
4. Document appears in the list with status badge
5. Backend worker picks it up asynchronously — status transitions streamed in real-time via SSE
6. User sees live progress: pending → processing → completed (or failed)

---

## Supported File Types

| Extension | Icon/Color Suggestion |
|-----------|----------------------|
| `.txt` | Document icon, gray |
| `.pdf` | PDF icon, red |
| `.csv` | Table/spreadsheet icon, green |
| `.json` | Code/brackets icon, orange |
| `.docx` | Word document icon, blue |
| `.xlsx` | Excel/spreadsheet icon, green (darker) |

**Max file size**: 50 MB

---

## Document Statuses & Visual Treatment

| Status | Badge Color | Behavior |
|--------|-------------|----------|
| `uploading` | Yellow/amber | Brief state while uploading to cloud storage |
| `pending` | Gray/neutral | Queued, waiting for worker to pick up |
| `processing` | Blue, with animated pulse/spinner | Shows pipeline step progress (5 steps) |
| `completed` | Green with checkmark | Ready for RAG queries |
| `failed` | Red with X icon | Shows error message on hover/expand |

---

## Processing Pipeline Steps (shown when status = `processing`)

The document goes through 5 sequential steps. Each step can be in one of three sub-states: `started`, `completed`, `failed`.

| Step | Label | Description |
|------|-------|-------------|
| 1. Download | "Downloading" | Pulling file from cloud storage |
| 2. Parse | "Parsing" | Extracting text based on file type |
| 3. Chunk | "Chunking" | Splitting text into embeddable fragments |
| 4. Embed | "Embedding" | Generating AI vector embeddings (OpenAI) |
| 5. Index | "Indexing" | Storing vectors in Qdrant for search |

---

## Dashboard Layout Suggestions

### Top Section
- **Header**: "Documents" with document count
- **Upload zone**: Prominent drop area with dashed border, file icon, "Drag & drop or click to browse" text, supported formats listed below, 50 MB limit note

### Main Content: Document List / Table
| Column | Content |
|--------|---------|
| File icon | Icon based on extension/type |
| Filename | Truncated with tooltip showing full name |
| Type | e.g. "PDF", "CSV", "JSON" |
| Size | Human-readable (KB/MB) |
| Status | Colored badge (see above) |
| Progress | When processing: step tracker showing 5 dots/checks with current step highlighted |
| Date | Time since upload (relative: "2 min ago", "1 hour ago") or absolute date |
| Actions | Delete button (trash icon), maybe retry for failed |

### Empty State
When no documents uploaded yet: centered illustration, "No documents yet" message, CTA "Upload your first document" button linked to the drop zone.

### Error / Failed State
Failed documents show a red badge. Expanding the row (or hovering) reveals the error message. A "Retry" action button could trigger re-processing.

### Real-time Updates
All status changes arrive via SSE (`GET /api/v1/stream`) as:
```json
{
  "type": "document_status",
  "data": {
    "id": 123,
    "filename": "report.pdf",
    "status": "processing",
    "error": ""
  }
}
```
The dashboard should listen to this stream and update document rows live without polling.

---

## API Endpoints for Dashboard

| Method | Path | Use |
|--------|------|-----|
| `GET` | `/api/v1/documents/` | Fetch all user documents (sorted by created_at desc) |
| `POST` | `/api/v1/documents/` | Upload new document (multipart, field: `file`) |
| `DELETE` | `/api/v1/documents/:id` | Delete document (removes from storage + DB + vector index) |

**Document response object**:
```json
{
  "id": 123,
  "filename": "report.pdf",
  "content_type": "application/pdf",
  "status": "completed",
  "size": 2457600,
  "created_at": "2026-05-06T10:30:00Z",
  "processing_started_at": "2026-05-06T10:30:05Z",
  "error_message": "",
  "retry_count": 0
}
```

---

## UX Requirements — Non-Negotiable

- **Zero confusion upload**: The drop zone must communicate exactly what's supported before any interaction. Show file type icons + "up to 50 MB" inline. No modals, no separate info pages.
- **Immediate feedback**: After dropping a file, show a file card instantly (optimistic UI) — don't wait for server response. The card transitions from "uploading" to the real status when the server responds.
- **Progress must be glanceable**: When a file is `processing`, show the 5 pipeline steps as a compact inline stepper. Current step is highlighted; completed steps show a checkmark. Don't force the user to expand or click.
- **Errors must self-diagnose**: A failed document shows the error message inline — not just a red badge. The message should be human-readable (e.g. "File too large — maximum is 50 MB" instead of "413 Payload Too Large").
- **Undoable deletes**: Deleting a document should show a toast with "Undo" for 5 seconds. After that, deletion is permanent (no confirmation dialog needed — the undo pattern is sufficient).
- **Keyboard-first for power users**: Tab through files, Enter to expand details, Delete key with selected row triggers deletion, Ctrl+U or Cmd+U to focus the upload zone.
- **Empty state guides, doesn't blame**: "Upload your first document to start asking questions" with a clear CTA. Show the supported formats as small badges below the CTA so users don't have to guess.
- **Progressive disclosure**: By default, show filename, type icon, status badge, and relative time. On row click/expand, reveal file size, pipeline step timing, content type, exact date, and full error detail.
- **Responsive without compromise**: On mobile, the drop zone becomes a full-width "Upload file" button. The file list becomes stacked cards. Pipeline stepper collapses to a single animated progress bar with step label.
- **Live without refresh**: The SSE stream updates statuses in real time. No manual refresh button needed. Show a subtle "Live" indicator (green dot + "Live" text) in the header when the SSE connection is active.
- **Batch upload**: Allow selecting multiple files at once. Show parallel upload progress for each file. Don't force users to upload one at a time.
- **Search & filter**: Search by filename (debounced, client-side). Filter by status (chips/tabs: All, Processing, Completed, Failed). Sort by name, date, size, or status.
- **Accessibility**: Status badges include `aria-label` with text equivalent (not just color). Pipeline stepper is keyboard-navigable. Color is never the sole indicator of state — pair color with icon + text. Contrast ratios meet WCAG AA.

## Visual Style Notes
- Clean, modern SaaS aesthetic — think Linear, Vercel, Notion
- Card-based list (not a dense table) with generous whitespace
- Subtle animations for status transitions (especially processing pulse)
- High contrast status badges for quick scanning
- Dark mode consideration (if applicable)
