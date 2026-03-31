//go:build windows

package glyph

import "math"

// Renderer renders laid-out text using a DrawBackend and glyph atlas.
//
// Not safe for concurrent use.
type Renderer struct {
	backend         DrawBackend
	atlas           *GlyphAtlas
	cache           map[uint64]CachedGlyph
	cacheAges       map[uint64]uint64
	pageKeys        map[int][]uint64
	maxCacheEntries int
	scaleFactor     float32
	scaleInv        float32
}

// RendererConfig configures the Renderer.
type RendererConfig struct {
	MaxGlyphCacheEntries int
}

func NewRenderer(backend DrawBackend, scaleFactor float32) (*Renderer, error) {
	return NewRendererWithConfig(backend, scaleFactor, 1024, 1024,
		RendererConfig{})
}

func NewRendererWithConfig(backend DrawBackend, scaleFactor float32,
	atlasW, atlasH int, cfg RendererConfig) (*Renderer, error) {

	atlas, err := NewGlyphAtlas(backend, atlasW, atlasH)
	if err != nil {
		return nil, err
	}
	safeScale := scaleFactor
	if safeScale <= 0 {
		safeScale = 1.0
	}
	maxEntries := cfg.MaxGlyphCacheEntries
	if maxEntries == 0 {
		maxEntries = 4096
	} else if maxEntries < 256 {
		maxEntries = 256
	}

	return &Renderer{
		backend:         backend,
		atlas:           atlas,
		cache:           make(map[uint64]CachedGlyph, 1024),
		cacheAges:       make(map[uint64]uint64, 1024),
		pageKeys:        make(map[int][]uint64),
		maxCacheEntries: maxEntries,
		scaleFactor:     safeScale,
		scaleInv:        1.0 / safeScale,
	}, nil
}

func (r *Renderer) Free() {
	r.atlas.Free()
	r.cache = nil
	r.cacheAges = nil
	r.pageKeys = nil
}

func (r *Renderer) Commit() {
	r.atlas.FrameCounter++
	r.atlas.SwapAndUpload()
}

func (r *Renderer) DrawLayout(layout Layout, x, y float32) {
	r.drawLayoutImpl(layout, x, y, AffineIdentity(), nil)
}

func (r *Renderer) DrawLayoutTransformed(layout Layout, x, y float32,
	transform AffineTransform) {
	r.drawLayoutImpl(layout, x, y, transform, nil)
}

func (r *Renderer) DrawLayoutRotated(layout Layout, x, y, angle float32) {
	r.drawLayoutImpl(layout, x, y, AffineRotation(angle), nil)
}

func (r *Renderer) DrawLayoutWithGradient(layout Layout, x, y float32,
	gradient *GradientConfig) {
	r.drawLayoutImpl(layout, x, y, AffineIdentity(), gradient)
}

func (r *Renderer) DrawLayoutTransformedWithGradient(layout Layout,
	x, y float32, transform AffineTransform, gradient *GradientConfig) {
	r.drawLayoutImpl(layout, x, y, transform, gradient)
}

// DrawLayoutPlaced renders each glyph at individual placements.
// Decorations are skipped. placements must match layout.Glyphs length.
func (r *Renderer) DrawLayoutPlaced(layout Layout, placements []GlyphPlacement) {
	if len(placements) != len(layout.Glyphs) || len(layout.Glyphs) == 0 {
		return
	}
	r.atlas.Cleanup(r.atlas.FrameCounter)

	// Pass 1: Stroke outlines.
	for _, item := range layout.Items {
		if !item.HasStroke || item.UseOriginalColor {
			continue
		}
		physW := item.StrokeWidth * r.scaleFactor

		for i := item.GlyphStart; i < item.GlyphStart+item.GlyphCount; i++ {
			if i < 0 || i >= len(layout.Glyphs) {
				continue
			}
			g := layout.Glyphs[i]
			if (g.Index & PangoGlyphUnknownFlag) != 0 {
				continue
			}
			placement := placements[i]
			cg := r.getOrLoadStrokedGlyph(item, g, physW)
			r.touchPage(cg)
			if cg.Width > 0 && cg.Height > 0 && cg.Page >= 0 && cg.Page < len(r.atlas.Pages) {
				r.emitPlacedQuad(cg, placement, item.StrokeColor,
					float32(item.Ascent), float32(item.Descent),
					item.UseOriginalColor, float32(g.XAdvance))
			}
		}
	}

	// Pass 2: Fill glyphs.
	for _, item := range layout.Items {
		if item.HasStroke && item.Color.A == 0 {
			continue
		}
		c := item.Color
		if item.UseOriginalColor {
			c = Color{255, 255, 255, 255}
		}

		for i := item.GlyphStart; i < item.GlyphStart+item.GlyphCount; i++ {
			if i < 0 || i >= len(layout.Glyphs) {
				continue
			}
			g := layout.Glyphs[i]
			if (g.Index & PangoGlyphUnknownFlag) != 0 {
				continue
			}
			placement := placements[i]
			bin := r.computeSubpixelBin(placement.X)
			cg := r.getOrLoadGlyph(item, g, bin)
			r.touchPage(cg)
			if cg.Width > 0 && cg.Height > 0 && cg.Page >= 0 && cg.Page < len(r.atlas.Pages) {
				r.emitPlacedQuad(cg, placement, c,
					float32(item.Ascent), float32(item.Descent),
					item.UseOriginalColor, float32(g.XAdvance))
			}
		}
	}
}

func (r *Renderer) Atlas() *GlyphAtlas { return r.atlas }

// --- helper methods ---

func (r *Renderer) getOrLoadGlyph(item Item, g Glyph, bin int) CachedGlyph {
	key := winGlyphCacheKey(g.Codepoint, int(item.Ascent+item.Descent), bin,
		winStyleHash(item.Style))

	if cg, ok := r.cache[key]; ok {
		r.cacheAges[key] = r.atlas.FrameCounter
		return cg
	}

	// Evict if cache is full.
	if len(r.cache) >= r.maxCacheEntries {
		r.evictOldGlyphs()
	}

	cfg := LoadGlyphConfig{
		Index:        g.Index,
		Codepoint:    g.Codepoint,
		TargetHeight: int(item.Ascent + item.Descent),
		SubpixelBin:  bin,
		Style:        item.Style,
	}

	result, err := r.atlas.LoadGlyph(cfg, r.scaleFactor)
	if err != nil {
		return CachedGlyph{}
	}

	if result.ResetOccurred {
		r.handleAtlasReset(result.ResetPage)
	}

	r.cache[key] = result.Cached
	r.cacheAges[key] = r.atlas.FrameCounter
	if result.Cached.Page >= 0 {
		r.pageKeys[result.Cached.Page] = append(
			r.pageKeys[result.Cached.Page], key)
	}
	return result.Cached
}

func (r *Renderer) getOrLoadStrokedGlyph(item Item, g Glyph, physStrokeWidth float32) CachedGlyph {
	key := winStrokedGlyphCacheKey(g.Codepoint, int(item.Ascent+item.Descent),
		physStrokeWidth, winStyleHash(item.Style))

	if cg, ok := r.cache[key]; ok {
		r.cacheAges[key] = r.atlas.FrameCounter
		return cg
	}

	if len(r.cache) >= r.maxCacheEntries {
		r.evictOldGlyphs()
	}

	cfg := LoadGlyphConfig{
		Index:        g.Index,
		Codepoint:    g.Codepoint,
		TargetHeight: int(item.Ascent + item.Descent),
		Style:        item.Style,
	}

	result, err := r.atlas.LoadStrokedGlyph(cfg, physStrokeWidth, r.scaleFactor)
	if err != nil {
		return CachedGlyph{}
	}

	if result.ResetOccurred {
		r.handleAtlasReset(result.ResetPage)
	}

	r.cache[key] = result.Cached
	r.cacheAges[key] = r.atlas.FrameCounter
	if result.Cached.Page >= 0 {
		r.pageKeys[result.Cached.Page] = append(
			r.pageKeys[result.Cached.Page], key)
	}
	return result.Cached
}

func winStrokedGlyphCacheKey(codepoint uint32, targetH int,
	strokeWidth float32, styleHash uint64) uint64 {
	h := fnvOffsetBasis
	h = fnvHashU64(h, uint64(codepoint))
	h = fnvHashU64(h, uint64(targetH))
	h = fnvHashF32(h, strokeWidth)
	h = fnvHashU64(h, styleHash)
	h = fnvHashU64(h, 0x5354524F4B45) // "STROKE" sentinel
	return h
}

func winGlyphCacheKey(codepoint uint32, targetH, bin int, styleHash uint64) uint64 {
	h := fnvOffsetBasis
	h = fnvHashU64(h, uint64(codepoint))
	h = fnvHashU64(h, uint64(targetH))
	h = fnvHashU64(h, uint64(bin))
	h = fnvHashU64(h, styleHash)
	return h
}

// winStyleHash returns a hash of the font-affecting fields in a TextStyle.
func winStyleHash(s TextStyle) uint64 {
	h := fnvOffsetBasis
	h = fnvHashString(h, s.FontName)
	h = fnvHashF32(h, s.Size)
	h = fnvHashU64(h, uint64(s.Typeface))
	return h
}

func (r *Renderer) evictOldGlyphs() {
	threshold := r.atlas.FrameCounter - 60
	for key, age := range r.cacheAges {
		if age < threshold {
			if cg, ok := r.cache[key]; ok {
				r.removePageKey(cg.Page, key)
			}
			delete(r.cache, key)
			delete(r.cacheAges, key)
		}
	}
}

func (r *Renderer) handleAtlasReset(page int) {
	keys := r.pageKeys[page]
	for _, k := range keys {
		delete(r.cache, k)
		delete(r.cacheAges, k)
	}
	r.pageKeys[page] = nil
}

func (r *Renderer) removePageKey(page int, key uint64) {
	keys := r.pageKeys[page]
	for i, k := range keys {
		if k == key {
			keys[i] = keys[len(keys)-1]
			r.pageKeys[page] = keys[:len(keys)-1]
			return
		}
	}
}

func (r *Renderer) touchPage(cg CachedGlyph) {
	if cg.Page >= 0 && cg.Page < len(r.atlas.Pages) {
		r.atlas.Pages[cg.Page].Age = r.atlas.FrameCounter
	}
}

func (r *Renderer) computeSubpixelBin(x float32) int {
	physX := x * r.scaleFactor
	snapped := float32(math.Round(float64(physX)*4.0)) / 4.0
	frac := snapped - float32(math.Floor(float64(snapped)))
	return int(frac*float32(SubpixelBins)+0.1) & (SubpixelBins - 1)
}

func (r *Renderer) computeDrawOrigin(targetX, targetY float32) (float32, float32, int) {
	scale := r.scaleFactor
	physX := targetX * scale
	snappedX := float32(math.Round(float64(physX)*4.0)) / 4.0
	drawOriginX := float32(math.Floor(float64(snappedX)))
	fracX := snappedX - drawOriginX
	bin := int(fracX*float32(SubpixelBins)+0.1) & (SubpixelBins - 1)

	physY := targetY * scale
	drawOriginY := float32(math.Round(float64(physY)))

	return drawOriginX, drawOriginY, bin
}
