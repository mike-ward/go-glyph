//go:build windows

package glyph

import "math"

// LoadGlyphConfig holds parameters for Windows glyph rasterization.
type LoadGlyphConfig struct {
	Index        uint32
	Codepoint    uint32
	ClusterText  string
	TargetHeight int
	SubpixelBin  int
	Style        TextStyle
}

// LoadGlyphResult holds the output of a glyph load operation.
type LoadGlyphResult struct {
	Cached        CachedGlyph
	ResetOccurred bool
	ResetPage     int
}

// LoadGlyph rasterizes a glyph via GDI and inserts it into the atlas.
func (atlas *GlyphAtlas) LoadGlyph(cfg LoadGlyphConfig, scaleFactor float32) (LoadGlyphResult, error) {
	if cfg.ClusterText == "" {
		return LoadGlyphResult{}, nil
	}

	gdi := getGDI()
	gdi.mu.Lock()
	gdi.selectFont(winFontParams(cfg.Style, scaleFactor))
	bmp, left, top := gdi.renderGlyphBitmap(cfg.ClusterText, cfg.TargetHeight)
	gdi.mu.Unlock()

	if bmp.Width == 0 || bmp.Height == 0 || len(bmp.Data) == 0 {
		return LoadGlyphResult{}, nil
	}

	cached, resetOccurred, resetPage, err := atlas.InsertBitmap(bmp, left, top)
	if err != nil {
		return LoadGlyphResult{}, err
	}

	return LoadGlyphResult{
		Cached:        cached,
		ResetOccurred: resetOccurred,
		ResetPage:     resetPage,
	}, nil
}

// LoadStrokedGlyph rasterizes a glyph via GDI and dilates the alpha channel
// to create a stroke/outline effect.
func (atlas *GlyphAtlas) LoadStrokedGlyph(cfg LoadGlyphConfig,
	physStrokeWidth, scaleFactor float32) (LoadGlyphResult, error) {

	if cfg.ClusterText == "" {
		return LoadGlyphResult{}, nil
	}

	gdi := getGDI()
	gdi.mu.Lock()
	gdi.selectFont(winFontParams(cfg.Style, scaleFactor))
	bmp, left, top := gdi.renderGlyphBitmap(cfg.ClusterText, cfg.TargetHeight)
	gdi.mu.Unlock()

	if bmp.Width == 0 || bmp.Height == 0 || len(bmp.Data) == 0 {
		return LoadGlyphResult{}, nil
	}

	radius := int(math.Ceil(float64(physStrokeWidth * 0.5)))
	if radius < 1 {
		radius = 1
	}
	bmp, left, top = dilateGlyphBitmap(bmp, left, top, radius)

	cached, resetOccurred, resetPage, err := atlas.InsertBitmap(bmp, left, top)
	if err != nil {
		return LoadGlyphResult{}, err
	}

	return LoadGlyphResult{
		Cached:        cached,
		ResetOccurred: resetOccurred,
		ResetPage:     resetPage,
	}, nil
}

// dilateGlyphBitmap expands the alpha channel of a glyph bitmap by radius
// pixels in all directions to create a stroke/outline effect.
func dilateGlyphBitmap(src Bitmap, left, top, radius int) (Bitmap, int, int) {
	newW := src.Width + radius*2
	newH := src.Height + radius*2
	newData := make([]byte, newW*newH*4)
	r2 := radius * radius

	for y := 0; y < newH; y++ {
		sy := y - radius
		for x := 0; x < newW; x++ {
			sx := x - radius
			var maxA byte
			for dy := -radius; dy <= radius; dy++ {
				oy := sy + dy
				if oy < 0 || oy >= src.Height {
					continue
				}
				for dx := -radius; dx <= radius; dx++ {
					if dx*dx+dy*dy > r2 {
						continue
					}
					ox := sx + dx
					if ox < 0 || ox >= src.Width {
						continue
					}
					if a := src.Data[(oy*src.Width+ox)*4+3]; a > maxA {
						maxA = a
					}
				}
			}
			if maxA > 0 {
				idx := (y*newW + x) * 4
				newData[idx+0] = 255
				newData[idx+1] = 255
				newData[idx+2] = 255
				newData[idx+3] = maxA
			}
		}
	}

	return Bitmap{
		Width:    newW,
		Height:   newH,
		Channels: 4,
		Data:     newData,
	}, left - radius, top + radius
}
