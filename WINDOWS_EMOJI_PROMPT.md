# Windows session: fix hollow Segoe UI Emoji in go-glyph

## Task

Segoe UI Emoji renders as hollow black outlines on Windows in go-glyph.
Fix it by adding a DirectWrite-based color glyph rasterization path to
the existing GDI Windows backend. Other platforms (macOS, Linux,
Android) are unaffected and should not be touched.

## Root cause

`gdi_windows.go` rasterizes all glyphs via classic GDI `TextOutW`.
Classic GDI **does not render the COLR table**. Segoe UI Emoji is COLR
v1, so GDI returns only the base outline glyph → hollow rendering.
DirectWrite **does** render COLR v0/v1 and is the standard Windows
solution. DirectWrite is a system DLL — no redistributables needed.

## Verified codebase state (do not re-investigate from scratch)

- **Windows uses GDI exclusively.** FreeType is NOT compiled into the
  Windows build. There is no FT fallback path on Windows.
  - `renderer_load_windows.go:24-49` — `LoadGlyph` calls
    `gdi.renderGlyphBitmap()` directly.
  - `gdi_windows.go:331-474` — `renderGlyphBitmap` uses `TextOutW`
    into a 32bpp DIB, then BGRA→RGBA swizzle on lines 430-451.
  - `gdi_windows.go:422` — comment says "GDI font linking handles
    emoji fallback" — this is the broken assumption; GDI font linking
    finds Segoe UI Emoji but still can't render its COLR table.

- **`isEmojiRune` already exists but is dead code.**
  - `gdi_windows.go:520-549` — checks emoji unicode ranges
    (0x1F600-0x1F64F etc). Not called from anywhere. Wire it up.

- **Atlas is already RGBA** (4 bpp). `atlas.go:246`,
  `backend/gpu/backend.go:45-50`. No atlas changes needed.

- **`CachedGlyph` has no color flag.** `atlas.go:52-60` — only
  `{X, Y, Width, Height, Left, Top, Page}`. Add a `Color bool` (or
  similar) so the renderer knows to skip text-color tinting for
  colored glyphs.

- **Vertex format carries per-vertex RGBA.** `backend/gpu/backend.go:77`
  — `Vertex{x, y, R, G, B, A, u, v}`. For color glyphs, emit white
  (1,1,1,1) so the texture passes through untinted. Simplest "ignore
  tint" path.

- **Other platforms already work** (or should — verification deferred):
  - macOS/Linux: FreeType ≥ 2.13 with `FT_LOAD_COLOR`, BGRA copy path
    live at `bitmap.go:121-141` (B/R swizzle to RGBA, premultiplied).
    `renderer_load.go:112-117` handles BGRA bearings.
  - Android: FreeType **2.13.3** statically linked
    (`scripts/build_android_deps.sh:12`). Supports COLR v1.
  - **Do not modify these paths.** If a regression appears, it's a
    Windows-only change leaking.

## Implementation plan

### 1. Add DirectWrite glyph rasterizer (new file)

Create `dwrite_windows.go` (cgo + Windows COM). Surface:

```go
type dwriteRasterizer struct { /* IDWriteFactory, etc. */ }

func newDWriteRasterizer() (*dwriteRasterizer, error)
func (d *dwriteRasterizer) Close()

// Rasterizes a single glyph as premultiplied BGRA. Returns pixels,
// width, height, bearing. Caller swizzles to RGBA.
func (d *dwriteRasterizer) RenderColorGlyph(
    fontFamily string,
    sizePx float32,
    dpi float32,
    codepoint rune,
) (pixels []byte, w, h, left, top int, err error)
```

Implementation outline:
- `DWriteCreateFactory(DWRITE_FACTORY_TYPE_SHARED, ...)` → `IDWriteFactory`.
- `IDWriteGdiInterop::CreateBitmapRenderTarget` → `IDWriteBitmapRenderTarget`
  bound to a memory HBITMAP. This is the easiest path: the render
  target gives you a GDI-compatible 32bpp DIB you can read pixels from.
- `IDWriteFactory::CreateTextFormat` for the font/size, OR build an
  `IDWriteFontFace` directly via the system font collection so you can
  call `DrawGlyphRun` for a single glyph index.
- Get glyph index from codepoint via `IDWriteFontFace::GetGlyphIndices`.
- Build a `DWRITE_GLYPH_RUN` (one glyph), call
  `IDWriteBitmapRenderTarget::DrawGlyphRun` with
  `DWRITE_MEASURING_MODE_NATURAL` and a default rendering params from
  `IDWriteFactory::CreateRenderingParams`. **Critical:** for COLR
  rendering, call the `IDWriteBitmapRenderTarget1::DrawGlyphRunWithColorSupport`
  variant if available, OR enumerate color glyph runs via
  `IDWriteFactory2::TranslateColorGlyphRun` and draw each layer with
  its palette color. The `TranslateColorGlyphRun` path is the most
  portable and supports COLR v0; for COLR v1 use
  `IDWriteFactory4::TranslateColorGlyphRun` (Windows 10 1709+).
- Read back pixels from the HBITMAP via `GetDIBits`. Format is
  premultiplied BGRA already.
- Crop to glyph bounds; return.

cgo: keep COM interaction in C (`dwrite_windows.h` / inline in the
.go file via `// #include`). Go side just calls thin wrappers.
Reference counting: every `IUnknown*` returned must be `Release()`d.
Wrap each acquired interface in a `defer` immediately after creation.

Link: `-ldwrite -ld2d1 -lgdi32` in cgo `LDFLAGS`.

### 2. Wire `isEmojiRune` into the load path

In `renderer_load_windows.go:24-49` (`LoadGlyph`), before calling
`gdi.renderGlyphBitmap`:

```go
if isEmojiRune(r) {
    pixels, w, h, left, top, err := dwrite.RenderColorGlyph(...)
    if err == nil {
        // copy into atlas, mark glyph as Color=true
        return ...
    }
    // fall through to GDI on failure
}
```

`isEmojiRune` lives in `gdi_windows.go:520`. May want to broaden the
ranges or add a "is this glyph in a COLR font" check via DWrite — but
the existing range check is good enough for v1.

### 3. Add `Color` flag to `CachedGlyph`

`atlas.go:52-60`. Add `Color bool`. Set true when the glyph came from
the DWrite path. Renderer (draw path) checks the flag and emits white
vertex color instead of the requested text color.

Search for where `CachedGlyph` is consumed by the vertex emit code
(likely `draw.go` or similar) and add the branch.

### 4. Initialize / teardown DWrite rasterizer

Owned by the Windows GDI context (`gdi_windows.go` — find the
`gdiContext` or equivalent struct). Construct in the same place GDI
DCs are created; close in the same teardown.

### 5. Verify

- Run `examples/showcase` (or whichever example uses text). Type/show
  "Hello 😀🎉🚀". Confirm filled color emoji.
- Screenshot before/after, save to `examples/windows_emoji_before.png`
  and `..._after.png`.
- Test at multiple sizes (11, 13, 16, 20, 32 px) and DPI (100%, 150%,
  200%). DPI awareness matters — DWrite rendering params should use
  the actual DPI from `GetDpiForWindow` or context DPI.
- Test fallback: temporarily force `RenderColorGlyph` to return error
  and confirm GDI fallback still produces (hollow but present) glyphs.
- Run `go build ./...` on Windows.
- Run `go test ./...` and `golangci-lint run ./...`.
- **Do not run tests for other backends** (gpu/metal, ebitengine on
  mac, etc) from Windows — they won't build.

## Files to touch

- **New**: `dwrite_windows.go` (+ optional `dwrite_windows.h`)
- **Edit**: `renderer_load_windows.go` — wire DWrite into LoadGlyph
- **Edit**: `gdi_windows.go` — own the dwriteRasterizer lifecycle;
  `isEmojiRune` may need minor expansion
- **Edit**: `atlas.go` — add `Color bool` to `CachedGlyph`
- **Edit**: `draw.go` (or wherever vertices are emitted) — branch on
  `CachedGlyph.Color` to skip text-color tint
- **Do not touch**: `bitmap.go`, `renderer_load.go`, anything under
  `*_darwin.go`, `*_linux.go`, `*_android.go`, `coretext_*.go`,
  `pango_*.go`, `freetype_*.go`. These are non-Windows.

## Hazards

- **COM ref leaks** crash hours later. Wrap every `IUnknown*` in
  `defer release()`.
- **TranslateColorGlyphRun versions**: `IDWriteFactory2` does COLR v0;
  `IDWriteFactory4` does v1. Cast factory with `QueryInterface` and
  fall back gracefully if v4 unavailable (Windows < 10 1709).
- **Premultiplied vs straight alpha**: DWrite output is premultiplied
  BGRA. Atlas/shader path on Windows must treat color glyphs as
  premultiplied. Verify the existing GDI BGRA path's expectation
  (`gdi_windows.go:430-451`) — match it.
- **DPI**: DWrite rendering params take pixels-per-DIP. Get it right
  or glyphs come out tiny/huge.
- **Glyph metrics origin**: DWrite uses baseline origin; atlas/draw
  uses top-left. Convert with the bitmap render target's pixels-per-DIP
  and font metrics from `IDWriteFontFace::GetMetrics`.
- **Caching**: cache the `IDWriteFontFace` per (family, weight, style).
  Constructing it per glyph is slow.
- **Threading**: DWrite factory is thread-safe (DWRITE_FACTORY_TYPE_SHARED)
  but bitmap render targets are not. Confine to the render thread or
  guard with a mutex.

## Project rules (from CLAUDE.md)

- Must pass `gofmt` and `golangci-lint run ./...` zero issues.
- All backends must expose identical API; this work adds no new public
  API — purely internal.
- Favor reducing heap allocations: pool the DIB / pixel staging buffer
  rather than allocating per glyph.
- Comments wrap at 90 columns.
- Do NOT commit without explicit permission.

## Reference docs to fetch on Windows

- DirectWrite color font rendering:
  https://learn.microsoft.com/en-us/windows/win32/directwrite/color-fonts
- `IDWriteFactory4::TranslateColorGlyphRun`:
  https://learn.microsoft.com/en-us/windows/win32/api/dwrite_3/nf-dwrite_3-idwritefactory4-translatecolorglyphrun
- `IDWriteBitmapRenderTarget`:
  https://learn.microsoft.com/en-us/windows/win32/api/dwrite/nn-dwrite-idwritebitmaprendertarget

## Out of scope (explicitly defer)

- Gamma-correct text blend (cross-platform shader work)
- Subpixel AA changes
- macOS/Linux/Android verification (separate task)
- Bundled fallback emoji font
- SVG-in-OT table support
- Refactoring `gdi_windows.go` beyond what's needed to wire DWrite in

## Open questions to resolve early on Windows

1. Min Windows version target? (decides whether COLR v1 via Factory4
   is acceptable or v0-only via Factory2 is the floor)
2. Is `IDWriteBitmapRenderTarget1::DrawGlyphRunWithColorSupport`
   acceptable, or stick with manual `TranslateColorGlyphRun` + per-layer
   draws? (former is less code, latter is more portable)
3. Pool size for the DIB staging buffer — one per context, or per glyph
   size bucket?
