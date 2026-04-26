# Proceso de Release - RAGo

Guía para crear una nueva versión y publicarla en GitHub.

## Requisitos Previos

- Go installed
- goreleaser instalado: `go install github.com/goreleaser/goreleaser@latest`
- Acceso a GitHub con permisos de write en el repositorio

## Pasos para Releases

### 1. Verificar que todo funcione localmente

```bash
# Build local
go build ./...

# Probar CLI
./rago debug
./rago ask "test"
```

### 2. Verificar cambios desde último release

```bash
git log --oneline v0.1.0..HEAD
```

### 3. Determinar tipo de release

| Tipo | Commits | Tag |
|------|--------|-----|
| **Patch** | `fix:`, `chore:`, `refactor:` | `v0.1.1` |
| **Minor** | `feat:` (sin BREAKING CHANGE) | `v0.2.0` |
| **Major** | `BREAKING CHANGE:` en commit o `feat!:` | `v1.0.0` |

### 4. Crear tag

```bash
# Eliminar tag local si existe
git tag -d v0.1.1

# Crear tag (semver)
git tag v0.1.1

# Push tag a GitHub
git push origin v0.1.1
```

**Importante:** El tag debe empezar con `v` seguido de semver.

### 5. Esperar a que Actions/dispatch se ejecute

1. Ir a GitHub > Actions > Release workflow
2. Esperar a que termine
3. Verificar release creado en GitHub > Releases

---

## Solución de Problemas

### Error: "couldn't find main file"

**Causa:** `main.go` no está commiteado.

**Solución:**
```bash
git status
git add .
git commit -m "feat: descripción del cambio"
git push
```

### Error: "GITHUB_TOKEN"

**Causa:** Permisos insuficientes.

**Solución:** En workflow, asegurar:
```yaml
permissions:
  contents: write
```

### Error: "shallow clone"

**Causa:** GoReleaser necesita historial completo.

**Solución:** En workflow agregar:
```yaml
- name: Checkout
  uses: actions/checkout@v4
  with:
    fetch-depth: 0
```

---

## Estructura del Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## Configuración GoReleaser

```yaml
# .goreleaser.yml
version: 2

project_name: rago

release:
  github:
    owner: whoAngeel
    name: rago

builds:
  - main: ./cmd/rag
    binary: rago
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - format: targz

checksum:
  algorithm: sha256

changelog:
  filters:
    exclude:
      - '^docs:'
      - '^chore:'
```

---

## Conventional Commits

Formato de commitspara semver automático:

```
<tipo>[opcional ámbito]: descripción

[opcional cuerpo]

[opcional pie(s)]
```

### Tipos:

| Tipo | Descripción | Release |
|------|------------|---------|
| `feat:` | Nueva funcionalidad | Minor |
| `fix:` | Bug fix | Patch |
| `docs:` | Documentación | No release |
| `chore:` | Tareas menores | No release |
| `refactor:` | Refactorización | No release |
| `test:` | Tests | No release |
| `BREAKING CHANGE:` | Cambio incompatible | Major |

### Ejemplos:

```bash
git commit -m "fix: corregir bug en search"
git commit -m "feat: agregar comando delete"
git commit -m "feat!: cambiar API - incompatible"
git commit -m "docs: actualizar README"
```

---

## Checklist Pre-Release

- [ ] `go build ./...` funciona
- [ ] `go test ./...` pasa (si hay tests)
- [ ] Cambios commiteados y push
- [ ] Tag creado con versión correcta
- [ ] Tag pusheado a origin
- [ ] Release aparece en GitHub
- [ ] Binarios disponibles para download
- [ ] CHANGELOG actualizado (opcional)

---

## Links Útiles

- [GoReleaser Docs](https://goreleaser.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [SemVer](https://semver.org/)