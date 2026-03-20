package glyph

import (
	"math"
	"time"
)

// cachedLayout holds a Layout and its last access time for eviction.
type cachedLayout struct {
	layout     Layout
	lastAccess int64 // Unix milliseconds.
}

// TextSystem is the main entry point for text rendering. It owns
// the Context, Renderer, and a layout cache.
//
// Not safe for concurrent use. Callers must serialize access
// externally if shared across goroutines.
type TextSystem struct {
	ctx             *Context
	renderer        *Renderer
	cache           map[uint64]*cachedLayout
	evictionAge     int64 // Milliseconds. Default 5000.
	maxCacheEntries int   // Max layout cache entries. Default 1024.
}

// NewTextSystem creates a TextSystem with default atlas size (1024x1024).
func NewTextSystem(backend DrawBackend) (*TextSystem, error) {
	scale := backend.DPIScale()
	ctx, err := NewContext(scale)
	if err != nil {
		return nil, err
	}
	renderer, err := NewRenderer(backend, scale)
	if err != nil {
		ctx.Free()
		return nil, err
	}
	return &TextSystem{
		ctx:             ctx,
		renderer:        renderer,
		cache:           make(map[uint64]*cachedLayout),
		evictionAge:     5000,
		maxCacheEntries: 1024,
	}, nil
}

// NewTextSystemAtlasSize creates a TextSystem with custom atlas dimensions.
func NewTextSystemAtlasSize(backend DrawBackend, atlasW, atlasH int) (*TextSystem, error) {
	if err := ValidateDimension(atlasW, "atlas_width", "NewTextSystemAtlasSize"); err != nil {
		return nil, err
	}
	if err := ValidateDimension(atlasH, "atlas_height", "NewTextSystemAtlasSize"); err != nil {
		return nil, err
	}
	scale := backend.DPIScale()
	ctx, err := NewContext(scale)
	if err != nil {
		return nil, err
	}
	renderer, err := NewRendererWithConfig(backend, scale, atlasW, atlasH, RendererConfig{})
	if err != nil {
		ctx.Free()
		return nil, err
	}
	return &TextSystem{
		ctx:             ctx,
		renderer:        renderer,
		cache:           make(map[uint64]*cachedLayout),
		evictionAge:     5000,
		maxCacheEntries: 1024,
	}, nil
}

// Free releases all TextSystem resources.
func (ts *TextSystem) Free() {
	if ts.renderer != nil {
		ts.renderer.Free()
		ts.renderer = nil
	}
	if ts.ctx != nil {
		ts.ctx.Free()
		ts.ctx = nil
	}
	ts.cache = nil
}

// DrawText renders text at (x, y) using configuration.
// Uses layout cache for repeated calls.
func (ts *TextSystem) DrawText(x, y float32, text string, cfg TextConfig) error {
	item, err := ts.getOrCreateLayout(text, cfg)
	if err != nil {
		return err
	}
	if cfg.Gradient != nil {
		ts.renderer.DrawLayoutWithGradient(item.layout, x, y, cfg.Gradient)
	} else {
		ts.renderer.DrawLayout(item.layout, x, y)
	}
	return nil
}

// TextWidth returns the width (pixels) of text if rendered with cfg.
func (ts *TextSystem) TextWidth(text string, cfg TextConfig) (float32, error) {
	item, err := ts.getOrCreateLayout(text, cfg)
	if err != nil {
		return 0, err
	}
	return item.layout.Width, nil
}

// TextHeight returns the visual height (pixels) of text.
func (ts *TextSystem) TextHeight(text string, cfg TextConfig) (float32, error) {
	item, err := ts.getOrCreateLayout(text, cfg)
	if err != nil {
		return 0, err
	}
	return item.layout.VisualHeight, nil
}

// FontHeight returns the font height (ascent + descent) in pixels.
func (ts *TextSystem) FontHeight(cfg TextConfig) (float32, error) {
	return ts.ctx.FontHeight(cfg)
}

// FontMetrics returns detailed font metrics.
func (ts *TextSystem) FontMetrics(cfg TextConfig) (TextMetrics, error) {
	return ts.ctx.FontMetrics(cfg)
}

// Commit uploads atlas textures and prunes the layout cache.
// Call once per frame after all draw calls.
func (ts *TextSystem) Commit() {
	if ts.renderer == nil {
		return
	}
	ts.renderer.Commit()
	ts.pruneCache()
}

// AddFontFile registers a font file (TTF/OTF).
// Clears the layout cache to prevent stale FT_Face pointers.
func (ts *TextSystem) AddFontFile(path string) error {
	if err := ValidateFontPath(path, "AddFontFile"); err != nil {
		return err
	}
	if err := ts.ctx.AddFontFile(path); err != nil {
		return err
	}
	clear(ts.cache)
	return nil
}

// ResolveFontName returns the actual font family name that Pango
// resolves for the given description string.
func (ts *TextSystem) ResolveFontName(name string) (string, error) {
	return ts.ctx.ResolveFontName(name)
}

// LayoutText computes a new Layout (bypasses cache).
func (ts *TextSystem) LayoutText(text string, cfg TextConfig) (Layout, error) {
	if err := ValidateTextInput(text, MaxTextLength, "LayoutText"); err != nil {
		return Layout{}, err
	}
	return ts.ctx.LayoutText(text, cfg)
}

// LayoutTextCached retrieves a cached layout or creates a new one.
func (ts *TextSystem) LayoutTextCached(text string, cfg TextConfig) (Layout, error) {
	item, err := ts.getOrCreateLayout(text, cfg)
	if err != nil {
		return Layout{}, err
	}
	return item.layout, nil
}

// LayoutRichText computes a Layout for multi-styled text.
func (ts *TextSystem) LayoutRichText(rt RichText, cfg TextConfig) (Layout, error) {
	return ts.ctx.LayoutRichText(rt, cfg)
}

// DrawLayout renders a pre-computed Layout at (x, y).
func (ts *TextSystem) DrawLayout(l Layout, x, y float32) {
	if ts.renderer == nil {
		return
	}
	ts.renderer.DrawLayout(l, x, y)
}

// DrawLayoutTransformed renders with an affine transform.
func (ts *TextSystem) DrawLayoutTransformed(l Layout, x, y float32,
	transform AffineTransform) {
	if ts.renderer == nil {
		return
	}
	ts.renderer.DrawLayoutTransformed(l, x, y, transform)
}

// DrawLayoutRotated renders rotated by angle (radians).
func (ts *TextSystem) DrawLayoutRotated(l Layout, x, y, angle float32) {
	ts.DrawLayoutTransformed(l, x, y, AffineRotation(angle))
}

// DrawLayoutWithGradient renders with gradient colors.
func (ts *TextSystem) DrawLayoutWithGradient(l Layout, x, y float32,
	gradient *GradientConfig) {
	if ts.renderer == nil {
		return
	}
	ts.renderer.DrawLayoutWithGradient(l, x, y, gradient)
}

// DrawLayoutTransformedWithGradient renders with both an affine transform
// and gradient colors.
func (ts *TextSystem) DrawLayoutTransformedWithGradient(
	l Layout,
	x, y float32,
	transform AffineTransform,
	gradient *GradientConfig,
) {
	if ts.renderer == nil {
		return
	}
	ts.renderer.DrawLayoutTransformedWithGradient(
		l, x, y, transform, gradient,
	)
}

// DrawLayoutPlaced renders glyphs at individual placements.
func (ts *TextSystem) DrawLayoutPlaced(l Layout, placements []GlyphPlacement) {
	if ts.renderer == nil {
		return
	}
	ts.renderer.DrawLayoutPlaced(l, placements)
}

// Renderer returns the underlying Renderer for advanced usage.
func (ts *TextSystem) Renderer() *Renderer { return ts.renderer }

// Context returns the underlying Context for advanced usage.
func (ts *TextSystem) Context() *Context { return ts.ctx }

// --- internal helpers ---

func (ts *TextSystem) getOrCreateLayout(text string, cfg TextConfig) (*cachedLayout, error) {
	if err := ValidateTextInput(text, MaxTextLength, "getOrCreateLayout"); err != nil {
		return nil, err
	}

	key := ts.getCacheKey(text, cfg)
	if item, ok := ts.cache[key]; ok {
		item.lastAccess = time.Now().UnixMilli()
		return item, nil
	}

	layout, err := ts.ctx.LayoutText(text, cfg)
	if err != nil {
		return nil, err
	}
	item := &cachedLayout{
		layout:     layout,
		lastAccess: time.Now().UnixMilli(),
	}
	ts.cache[key] = item
	if ts.maxCacheEntries > 0 && len(ts.cache) > ts.maxCacheEntries {
		ts.evictOldestLayout()
	}
	return item, nil
}

// getCacheKey hashes text + config into a cache key (FNV-1a).
// Gradient is excluded — it affects rendering color only.
func (ts *TextSystem) getCacheKey(text string, cfg TextConfig) uint64 {
	h := fnvOffsetBasis

	// Text.
	h = fnvHashString(h, text)
	h = fnvHashU64(h, 124) // separator

	// TextStyle.
	h = fnvHashString(h, cfg.Style.FontName)
	h = fnvHashF32(h, cfg.Style.Size)
	h = fnvHashColor(h, cfg.Style.Color)
	h = fnvHashColor(h, cfg.Style.BgColor)
	h = fnvHashF32(h, cfg.Style.LetterSpacing)
	h = fnvHashF32(h, cfg.Style.StrokeWidth)
	h = fnvHashColor(h, cfg.Style.StrokeColor)

	// Pack scalar fields.
	packed := uint64(cfg.Style.Typeface)
	if cfg.Style.Underline {
		packed |= 1 << 4
	}
	if cfg.Style.Strikethrough {
		packed |= 1 << 5
	}
	packed |= uint64(cfg.Block.Align) << 6
	packed |= uint64(int64(cfg.Block.Wrap)+1) << 10
	if cfg.UseMarkup {
		packed |= 1 << 14
	}
	if cfg.NoHitTesting {
		packed |= 1 << 15
	}
	packed |= uint64(cfg.Orientation) << 16
	h = fnvHashU64(h, packed)

	// Features.
	if cfg.Style.Features != nil {
		for _, f := range cfg.Style.Features.OpenTypeFeatures {
			h = fnvHashString(h, f.Tag)
			h = fnvHashU64(h, uint64(f.Value))
		}
		for _, a := range cfg.Style.Features.VariationAxes {
			h = fnvHashString(h, a.Tag)
			h = fnvHashF32(h, a.Value)
		}
	}

	// Inline object.
	if cfg.Style.Object != nil {
		h = fnvHashString(h, cfg.Style.Object.ID)
		h = fnvHashF32(h, cfg.Style.Object.Width)
		h = fnvHashF32(h, cfg.Style.Object.Height)
		h = fnvHashF32(h, cfg.Style.Object.Offset)
	}

	// BlockStyle.
	h = fnvHashF32(h, cfg.Block.Width)
	h = fnvHashF32(h, cfg.Block.Indent)
	h = fnvHashF32(h, cfg.Block.LineSpacing)
	for _, t := range cfg.Block.Tabs {
		h = fnvHashU64(h, uint64(t))
	}

	return h
}

func (ts *TextSystem) pruneCache() {
	if len(ts.cache) == 0 {
		return
	}
	now := time.Now().UnixMilli()
	for k, item := range ts.cache {
		if now-item.lastAccess > ts.evictionAge {
			delete(ts.cache, k)
		}
	}
}

func (ts *TextSystem) evictOldestLayout() {
	var oldestKey uint64
	oldestTime := int64(math.MaxInt64)
	for k, item := range ts.cache {
		if item.lastAccess < oldestTime {
			oldestTime = item.lastAccess
			oldestKey = k
		}
	}
	if oldestTime < math.MaxInt64 {
		delete(ts.cache, oldestKey)
	}
}

// --- FNV-1a hash helpers ---

const fnvOffsetBasis = uint64(14695981039346656037)
const fnvPrime = uint64(1099511628211)

func fnvHashString(h uint64, s string) uint64 {
	for i := range len(s) {
		h ^= uint64(s[i])
		h *= fnvPrime
	}
	return h
}

func fnvHashU64(h, v uint64) uint64 {
	h ^= v
	h *= fnvPrime
	return h
}

func fnvHashF32(h uint64, v float32) uint64 {
	return fnvHashU64(h, uint64(math.Float32bits(v)))
}

func fnvHashColor(h uint64, c Color) uint64 {
	u := uint32(c.R) | uint32(c.G)<<8 | uint32(c.B)<<16 | uint32(c.A)<<24
	return fnvHashU64(h, uint64(u))
}
