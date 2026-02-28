package glyph

// FontMetricsEntry stores cached font metrics in Pango units.
type FontMetricsEntry struct {
	Ascent  int // Pango units.
	Descent int // Pango units.
	LineGap int // Pango units (0 if not available).
}

// metricsCache is an LRU cache for font metrics keyed by
// (face pointer XOR size) tuple.
type metricsCache struct {
	entries     map[uint64]FontMetricsEntry
	accessOrder []uint64 // Most recent at end.
	capacity    int
}

func newMetricsCache(capacity int) metricsCache {
	return metricsCache{
		entries:     make(map[uint64]FontMetricsEntry, capacity),
		accessOrder: make([]uint64, 0, capacity),
		capacity:    capacity,
	}
}

func (c *metricsCache) get(key uint64) (FontMetricsEntry, bool) {
	entry, ok := c.entries[key]
	if !ok {
		return FontMetricsEntry{}, false
	}
	// Move to end (most recent).
	c.moveToEnd(key)
	return entry, true
}

func (c *metricsCache) put(key uint64, entry FontMetricsEntry) {
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.capacity {
		// Evict oldest (first in accessOrder).
		if len(c.accessOrder) > 0 {
			evictKey := c.accessOrder[0]
			delete(c.entries, evictKey)
			c.accessOrder = c.accessOrder[1:]
		}
	}
	c.entries[key] = entry
	c.moveToEnd(key)
}

func (c *metricsCache) moveToEnd(key uint64) {
	for i, k := range c.accessOrder {
		if k == key {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			break
		}
	}
	c.accessOrder = append(c.accessOrder, key)
}
