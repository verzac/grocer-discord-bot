package ingredients

import (
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

const pendingDefaultTTL = 5 * time.Minute

type pendingIngredients struct {
	Ingredients []string
	GuildID     string
	AuthorID    string
	ListLabel   string
}

type pendingCache struct {
	c *cache.Cache
}

func newPendingCache() *pendingCache {
	return &pendingCache{
		c: cache.New(pendingDefaultTTL, 10*time.Minute),
	}
}

func (p *pendingCache) set(data *pendingIngredients) string {
	key := data.GuildID + ":" + uuid.NewString()
	p.c.Set(key, data, pendingDefaultTTL)
	return key
}

func (p *pendingCache) peek(key string) (*pendingIngredients, bool) {
	v, ok := p.c.Get(key)
	if !ok {
		return nil, false
	}
	out, ok := v.(*pendingIngredients)
	if !ok {
		return nil, false
	}
	return out, true
}

func (p *pendingCache) take(key string) (*pendingIngredients, bool) {
	v, ok := p.c.Get(key)
	if !ok {
		return nil, false
	}
	p.c.Delete(key)
	out, ok := v.(*pendingIngredients)
	if !ok {
		return nil, false
	}
	return out, true
}

func (p *pendingCache) delete(key string) {
	p.c.Delete(key)
}
