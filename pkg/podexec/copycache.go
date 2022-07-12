package podexec

func NewMgctlCopyCache() *mgctlCopyCache {
	return &mgctlCopyCache{cache: make(map[string]bool)}
}

func (c *mgctlCopyCache) setCache(podId string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[podId] = true
}

func (c *mgctlCopyCache) isCached(podId string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, cached := c.cache[podId]
	return cached
}
