---
name: add-platform-variant
description: Scaffold matched per-OS Go files (windows/ios/android/wasm + shared + test) for a new go-glyph feature, with correct `//go:build` tags. Invoke when adding any API that needs a platform-specific implementation. User-only.
disable-model-invocation: true
---

# add-platform-variant

You scaffold a set of parallel per-OS files for a new feature in **go-glyph**. The repo convention is: one shared file (no build tag) declaring the public API or cross-platform helpers, plus one per-OS file for each supported GOOS, plus one or more tests.

## Inputs

Ask the user (if not given as arguments):

1. **Feature base name** in `snake_case` — e.g. `shadow`, `emoji_flag`, `color_font`.
   This becomes the file stem: `shadow.go`, `shadow_windows.go`, etc.
2. **Package directory** (relative to repo root) — defaults to the repo root package `glyph`. Other valid targets: `backend/ebitengine`, `backend/sdl2`, `backend/gpu`, `backend/android`, `backend/ios`, `backend/web`, `accessibility`, `ime`.
3. **Which platforms** to scaffold. Default: all five — `windows`, `ios`, `android`, `js && wasm`, plus the shared file. The user can opt out of any.
4. **One-line purpose** — a human-readable description of the feature, used in the top-of-file doc comment.

## Files to create

For feature `<name>` in package directory `<pkg>`:

| File | Build tag | Content |
|------|-----------|---------|
| `<pkg>/<name>.go` | *(none)* | Shared types, public API declarations, platform-neutral helpers. |
| `<pkg>/<name>_windows.go` | `//go:build windows` | Windows impl stub (GDI/DirectWrite/purego). |
| `<pkg>/<name>_ios.go` | `//go:build ios` | iOS impl stub (CoreText, via CGo). |
| `<pkg>/<name>_android.go` | `//go:build android` | Android impl stub (FreeType/HarfBuzz, via CGo). |
| `<pkg>/<name>_wasm.go` | `//go:build js && wasm` | WASM impl stub. `CGO_ENABLED=0` — no C imports. |
| `<pkg>/<name>_test.go` | *(none)* | Platform-neutral tests; per-OS tests live next to their impl files as `<name>_<os>_test.go` if needed. |

## Template shape

Every generated file must begin with a build tag (if any), then `package <pkg>`, then a brief doc comment referencing the feature name and a TODO marker so it's obvious the impl is pending. Example `shadow_ios.go`:

```go
//go:build ios

package glyph

// drawShadowImpl renders the shadow on iOS using CoreText.
// TODO(shadow): implement — see shadow.go for the API contract.
func drawShadowImpl( /* TODO: fill in params */ ) {
	// TODO(shadow): iOS implementation
}
```

The shared file declares the public surface (types, exported funcs) that each platform impl must satisfy, plus a dispatch point (e.g., `drawShadowImpl(...)`) that platform files implement.

## Procedure

1. Verify the target package directory exists. If it doesn't, ask before creating.
2. Check for pre-existing files with any of the target names. **Never overwrite**. If a collision exists, report it and stop.
3. Read one existing multi-platform feature as a reference for style (good anchors: `draw_*.go`, `layout_*.go`, `context_*.go`). Match indentation (tabs), import grouping, doc comment style.
4. Generate the files listed above.
5. Run `gofmt -w` on every new file.
6. Run `go vet ./<pkg>/...` on the host to confirm the host-platform file at least builds.
7. Report which files were created with a 1-line purpose each, and remind the user that **each `TODO` must be filled in** before the feature is real.

## Guardrails

- Never add a CGo import (`import "C"`) on the WASM variant — WASM is `CGO_ENABLED=0`.
- Never generate a file without a package declaration.
- Never scaffold into `examples/` — examples have their own module and layout conventions.
- If the user says "also add a backend file", ask which backend (`backend/ebitengine`, `backend/sdl2`, ...) — don't guess.
- If the feature already has a shared file but is missing one platform variant, only create the missing one; don't clobber.
