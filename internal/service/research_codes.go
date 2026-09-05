package service

import (
	"context"
	"sync"

	"github.com/dovod-app/app/internal/storage"
)

// codeCacheLimit is a ceiling, not a working set. A research's code is looked
// up once and kept for the life of the process, so the map only grows with the
// number of researches ever touched; the reset is here so a very long-lived
// server cannot grow it without bound.
const codeCacheLimit = 4096

// ResearchCodes resolves a research id to its short code, and remembers.
//
// Events carry the id, but every link in the web UI is built from the code, so
// a page that only knows `R7` cannot match an event that only says
// `9f3c…`. Rather than making each event pay for a lookup, this caches: a code
// is assigned once at creation and never changes, so a hit is always correct
// and a stale entry is not a thing that can exist.
type ResearchCodes struct {
	researches *storage.ResearchRepository

	mu    sync.RWMutex
	cache map[string]string
}

func NewResearchCodes(researches *storage.ResearchRepository) *ResearchCodes {
	return &ResearchCodes{
		researches: researches,
		cache:      make(map[string]string),
	}
}

// Code returns the short code, or "" if it cannot be determined. The empty
// answer is not an error worth propagating: it costs the client one matching
// strategy out of two, and the id still works.
func (c *ResearchCodes) Code(ctx context.Context, researchID string) string {
	if c == nil || researchID == "" {
		return ""
	}

	c.mu.RLock()
	code, ok := c.cache[researchID]
	c.mu.RUnlock()
	if ok {
		return code
	}

	research, err := c.researches.FindByID(ctx, researchID)
	if err != nil || research == nil || research.Code == "" {
		return ""
	}

	c.mu.Lock()
	if len(c.cache) >= codeCacheLimit {
		c.cache = make(map[string]string, codeCacheLimit)
	}
	c.cache[researchID] = research.Code
	c.mu.Unlock()

	return research.Code
}
