package mock

import "github.com/upfluence/goutils/stringcache"

type Cache struct {
	st map[string]bool
}

func NewCache() *Cache {
	return &Cache{st: make(map[string]bool)}
}

func (c *Cache) Has(k string) (bool, error) {
	_, e := c.st[k]

	return e, nil
}

func (c *Cache) Add(k string) error {
	c.st[k] = true

	return nil
}

func (c *Cache) Delete(k string) error {
	_, e := c.st[k]

	if !e {
		return stringcache.ErrNotFound
	}

	delete(c.st, k)

	return nil
}

func (c *Cache) Keys() []string {
	var r []string

	for k := range c.st {
		r = append(r, k)
	}

	return r
}
