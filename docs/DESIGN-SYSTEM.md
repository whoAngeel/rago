# Design System — Brainfish (`brainfishai.com`)

> Extraído del CSS computado en producción desde `https://www.brainfishai.com/product/ai-support-agents`  
> Fecha de análisis: 2026-05-03

---

## 0. Identidad visual general

Estética **"soft neo-brutalism"** / "playful flat":
- Trazos negros gruesos (1–2 px), esquinas suaves (no totalmente redondeadas excepto pills)
- Sombras *hard-offset* sin blur (desplazadas X/Y px) que dan profundidad sin realismo
- Paleta saturada pero pastel-friendly: verde lima como acento dominante, rosas/violetas/melocotón como ambientes
- Ilustraciones planas con motivos acuáticos (peces, burbujas, ondas) que refuerzan la marca "Brainfish"

---

## 1. Paleta de colores

### 1.1 Neutros

| Token | Hex | Uso |
|---|---|---|
| `--white` | `#ffffff` | Fondo cards, botones secundarios, navbar pill |
| `--black` | `#000000` | Texto principal, contornos |
| `--dark-50` | `#fafafa` | Fondos sutiles |
| `--dark-100` | `#f5f5f5` (whitesmoke) | Sección painpoints (`background-color-smokewhite`) |
| `--dark-200` | `#e5e5e5` | Líneas |
| `--dark-300` | `#d4d4d4` | Bordes suaves |
| `--dark-400` | `#a3a3a3` | Texto deshabilitado |
| `--dark-500` | `#737373` | Borde de input |
| `--dark-600` | `#525252` | Texto secundario |
| `--dark-700` | `#404040` | — |
| `--dark-800` | `#262626` | — |
| `--dark-900` / `--border` / `--dark-border` | `#171717` | Color universal de bordes (no negro puro) |
| `--box-shadow` / `--dark-boxshadow` | `#0a0a0d` | Color de sombras hard-offset |

### 1.2 Primario — Verde lima (firma de marca)

| Token | Hex | Nota |
|---|---|---|
| `--primary-50` | `#f7fee8` | |
| `--primary-200` | `#d9f99e` | |
| `--primary-300` | `#bef265` | |
| `--primary-400` / `--green` | `#a3e635` | **Botón "Book Demo"** |
| `--primary-500` | `#84cc17` | |
| `--primary-600` | `#65a30e` | |
| `--primary-700` | `#4e7c10` | |
| `--primary-800` | `#406213` | |
| `--primary-900` | `#375415` | |
| `--light-green` | `#ecfccc` | Inicio gradiente CTA |
| `--dark-green` | `#35d399` | |
| `--base-light-green` | `#d2fae5` | Gradiente CTA |
| `--green-700` | `#057857` | |

### 1.3 Acentos cromáticos

| Familia | Light / Pastel | Saturado |
|---|---|---|
| Rosa | `#fae9ff` · `#f5d1fe` (fondo hero) · `#f5d1fe` | `#e87af9` |
| Púrpura | `#edeafe` · `#f5f4ff` · `#ded7fe` | `#8c5df6` |
| Azul | `#dceafe` · `#f0f6ff` · `#94c5fd` · `#c0dbfe` | `#3c82f6` |
| Naranja | `#ffedd6` · `#fdba75` | `#fb923d` |
| Amarillo | `#fef3c8` · `#fffbec` · `#fde68b` | `#fbbf25` |
| Rojo | `#ffe5e6` | `#f4405e` |

### 1.4 Gradientes

| Sección | Valor |
|---|---|
| **CTA "Interested?"** | `linear-gradient(90deg, #ecfccc 0%, #d2fae5 32%, #edeafe 72%, #fae9ff 100%)` |
| **Footer green** | `linear-gradient(180deg, #b4edd8 0%, #6adbb1 100%)` |

---

## 2. Tipografía

### 2.1 Familias

- **Primaria: `Satoshi`** — titulares, body, todos los UI strings.  
  Stack: `Satoshi, Arial, sans-serif`  
  Pesos usados: **500** (dominante), **700** (números/énfasis), 900 (declarado, no activo en esta página)
- **`Inter`** — declarada y precargada (300/400/500/600/700) pero reservada para blog/docs; no aparece como `font-family` computada en esta página
- `webflow-icons` — iconografía interna del CMS

### 2.2 Escala tipográfica (viewport 1440 px)

| Rol | Size | Weight | Line-height | Letter-spacing |
|---|---|---|---|---|
| **H1** (hero) | 64 px | 500 | 74 px | -1.28 px |
| **H2** (sección) | 48 px | 500 | 64 px | -0.96 px |
| **H3** (card title) | 32 px | 500 | 44 px | -0.64 px |
| Body `<body>` | 20 px | 500 | 28 px | normal |
| Body párrafo `<p>` | 14 px | 500 | 22 px | normal |
| Card number "01" | 24 px | **700** | — | normal |
| Botón / pill / tag | 16 px | 500 | — | normal |
| Subscribe button | 18 px | 500 | — | normal |
| Status badge / footer link | 14 px | 500 | — | normal |

> Títulos siempre con letter-spacing negativo ≈ -2 % del tamaño.  
> Peso **500** en todos los roles — no usan bold para titulares, lo que aporta el aire "amigable".

---

## 3. Espaciado y layout

### 3.1 Escala de padding (deducida del CSS computado)

| Componente | Padding |
|---|---|
| Pill tag (default) | `6px 14px` |
| Pill tag (small / "Take a Look") | `4px 14px` |
| Botón nav | `8px 16px` |
| Botón large (subscribe) | `12px 24px` |
| Status badge | `8.5px 16px` |
| Navbar pill interior | `12px 48px` |
| Results card | `40px 30px 40px 60px` |
| Painpoint card | `18px 18px 18px 38px` |
| Footer | `80px 0 40px` |

### 3.2 Contenedor y grid

| Nivel | Ancho |
|---|---|
| Viewport de diseño | 1440 px |
| Navbar + wrappers | 1376 px (margen 32 px) |
| Cards / sección CTA | 1312 px (margen 64 px) |
| Hero (full bleed) | 1440 × 863 px |
| Gap entre columnas en results card | 64 px |
| Gap entre iconos/texto en botones y pills | 8 px |
| Separación entre painpoint cards | margin-bottom 22 px |

---

## 4. Bordes, radios y sombras

### 4.1 Border-radius

| Componente | Radius |
|---|---|
| Botón estándar | **4 px** |
| Input | 4 px |
| Status badge | 4 px |
| Painpoint card | **10 px** |
| Hero / CTA gradient | **12 px** |
| Footer green (solo esquinas inferiores) | `0 0 12px 12px` |
| Results card | **20 px** |
| Pill / tag / navbar | **100 px** (cápsula) |

### 4.2 Bordes

| Uso | Valor |
|---|---|
| Pills, botones, inputs, status badge | `1px solid #171717` |
| Results card (jerarquía alta) | `2px solid #0a0a0d` |
| Input (idle) | `1px solid #737373` |

### 4.3 Box-shadow — hard-offset, 0 blur

| Componente | Shadow |
|---|---|
| Status badge | `1px 1px 0 0 #0a0a0d` |
| Botones / navbar pill / subscribe | `2px 2px 0 0 #0a0a0d` |
| Results card / painpoint card | `4px 4px 0 0 #0a0a0d` |
| Pills inline (`.tag`) | `none` |

> **Patrón**: a mayor jerarquía visual → mayor offset (1 → 2 → 4 px). Sin blur en ningún caso.

---

## 5. Componentes UI

### 5.1 Navbar (pill flotante)

```
Clase: .nav_min-block
bg: #ffffff
border: 1px solid #171717
border-radius: 100px
box-shadow: 2px 2px 0 0 #0a0a0d
padding: 12px 48px
width: 1376px
height: 68px
font: Satoshi 20px/500
```

Estructura: logo izq · links centrales (`Why Brainfish`, `Product ▾`, `Customers`, `Resources ▾`, `Pricing`) · `Sign in` (btn white) + `Book Demo` (btn verde)

### 5.2 Botón primario (CTA verde)

```
Clase: .button.is-small-nav
bg: #a3e635
color: #000
border: 1px solid #171717
border-radius: 4px
box-shadow: 2px 2px 0 0 #0a0a0d
padding: 8px 16px
height: 42px
font: Satoshi 16px/500
gap: 8px (icono + texto)
```

Variante large (subscribe): `padding 12px 24px`, `font-size 18px`, `height 52px`

### 5.3 Botón secundario (white / Sign in)

```
Igual que el primario pero bg: #ffffff
```

### 5.4 Tag / Pill (cabeceras de sección)

```
Clase: .tag
bg: #ffffff
border: 1px solid #171717
border-radius: 100px
padding: 6px 14px (default) / 4px 14px (small)
height: 38px (default) / 34px (small)
font: Satoshi 16px/500
gap: 8px (icono + texto)
box-shadow: none
```

Cada sección abre con un pill: icono coloreado (estrella rosa, roja, sparkle…) + nombre de sección.

### 5.5 Results card (`.results_card`)

```
bg: #ffffff
border: 2px solid #0a0a0d
border-radius: 20px
box-shadow: 4px 4px 0 0 #0a0a0d
padding: 40px 30px 40px 60px
gap: 64px (dos columnas)
width: 1312px
```

Estructura columna izq: número `24px/700` → H3 `32px/500` → línea punteada → párrafo + logo cliente + Case Study link + "Take a Look" pill  
Columna der: imagen ilustrativa de producto

### 5.6 Painpoint card

```
Clase: .painpoints_points
bg: #ffffff
border: 1px solid #171717
border-radius: 10px
box-shadow: 4px 4px 0 0 #0a0a0d
padding: 18px 18px 18px 38px
margin: 0 0 22px 40px
width: 440px / height: 66px
font: Satoshi 20px/500
```

Cada card lleva un "pin" rojo de 70×70 px posicionado fuera del contenedor (izquierda).

### 5.7 Input (newsletter)

```
Clase: .email_field
bg: #ffffff
border: 1px solid #737373
border-radius: 4px
padding: 12px
width: 263px / height: 52px
font: Satoshi 16px/500
```

### 5.8 Status badge (footer)

```
Clase: .footer_bottom_tag
bg: #ffffff
border: 1px solid #000
border-radius: 4px
box-shadow: 1px 1px 0 0 #0a0a0d
padding: 8.5px 16px
width: 203px / height: 41px
font: Satoshi 14px/500
```

### 5.9 Testimonios (slider Splide)

```
Dimensión slide: 403 × 512 px
bg: transparente (sin borde ni sombra en el contenedor)
Font: Satoshi 20px/500
```

Estructura: foto cuadrada redondeada · nombre/cargo/empresa + quote en cursiva

---

## 6. Backgrounds por sección

| Sección | Background |
|---|---|
| Hero "Reduce Customer Effort" | `#f5d1fe` (`--pink-200`) plano |
| Tira de logos clientes | `#a3e635` con ondas SVG blancas top/bottom |
| "The Impact of AI Support Agents" | `#f5f5f5` (whitesmoke) |
| "Stop Losing Users" (painpoints) | `#f5f5f5` (smokewhite wrapper exterior) |
| "Why Teams Choose Brainfish" | `#f5f5f5` |
| CTA "Interested?" | `linear-gradient(90deg, #ecfccc, #d2fae5 32%, #edeafe 72%, #fae9ff)` · radius 12px |
| Footer | `linear-gradient(180deg, #b4edd8, #6adbb1)` · radius `0 0 12px 12px` |

---

## 7. Iconografía y elementos de marca

- **Logo**: pez estilizado + wordmark "Brainfish" en Satoshi bold negro
- **Peces verde lima** dispersos por el hero como decoración flotante
- **Burbujas** (círculos negros pequeños de distintos tamaños) — decenas alrededor de hero y footer
- **Onda verde** (`#a3e635`) en el borde inferior del hero y superior del footer — separador entre secciones
- **Estrellas de 4 puntas** (rosa pastel, roja) antes de los pills de sección
- **Polígonos** gradiente en sección CTA
- **Pins** rojos (70×70 px) con sombra hard-offset en la sección de painpoints
- **Logos de clientes** en blanco/negro sobre carrusel verde lima con bordes ondulados

---

## 8. Estados interactivos

El patrón neo-brutalist implica:

| Estado | Comportamiento esperado |
|---|---|
| **Hover botón/card** | `transform: translate(2px, 2px)` + `box-shadow: 0 0 0 0` → "el elemento se hunde sobre la sombra" |
| **Focus** | Outline negro accesible (borde ya existente reforzado) |
| **Input idle** | border `#737373` |
| **Input focus** | border `#171717` |
| **Disabled** | Colores `--dark-300` / `--dark-400` (gris claro/medio) |
| **Links navbar** | `color #000`, cursor pointer; el hover lo marca el cursor + underline |

---

## 9. Patrones de layout

- **Secciones-card**: cada sección es una "card" gigante con radius 12 px embebida en wrapper blanco — crea el efecto de tarjetas dentro de la página
- **Split 50/50**: títulos H2 a la izquierda, párrafo introductorio a la derecha en cabeceras de sección
- **Numeración explícita**: `01–05` como ancla visual de jerarquía en cards de impacto
- **Carruseles infinitos**: logos clientes y testimonios usan Splide.js
- **Ondas SVG**: sustituyen líneas rectas como separadores entre secciones
- **Alta densidad decorativa** (peces, burbujas, pins, estrellas) compensada con tipografía neutra (Satoshi 500) y espaciado generoso

---

## 10. Cheat sheet

```
Tipografía:  Satoshi — peso 500 dominante, 700 para énfasis numérico
Acento:      #a3e635 (verde lima) sobre fondo #ffffff
Borde:       1px–2px sólido #171717
Sombra:      hard-offset sin blur, color #0a0a0d — escala 1/2/4 px
Radios:      pill 100px · btn/input 4px · sección 12px · card 20px · painpoint 10px
Fondos:      pink hero #f5d1fe · smokewhite #f5f5f5 · gradient lima→menta→lila→rosa CTA · gradient menta footer
Marca:       peces, burbujas, ondas, estrellas, pins — flat illustrations
Layout:      secciones-card 1312–1376 px sobre lienzo 1440 px
```
