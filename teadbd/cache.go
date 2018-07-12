package main

import (
	"log"

	"gitlab.com/hokiegeek.net/teadb"
)

type cachedAllTeas struct {
	Valid bool
	Item  []teadb.Tea
}

type cachedTea struct {
	Valid bool
	Item  teadb.Tea
}

/*
type CachedTeaEntry struct {
	checksum string
	valid bool
	item teadb.TeaEntry
}

func (c *CachedTeaEntry) invalidate() {
	c.valid = false
}
*/

// Invalidate invalidates everything cached
func (c *Cache) Invalidate() {
	log.Println("Invalidated cache")
	c.teas.Valid = false
	for _, v := range c.teaByID {
		v.Valid = false
	}
}

// Cache is a cache struct
type Cache struct {
	teas    *cachedAllTeas
	teaByID map[int]*cachedTea
}

// AllTeasValid returns if all teas is valid
func (c *Cache) AllTeasValid() bool {
	return c.teas.Valid
}

// GetAllTeas returns the teas
func (c *Cache) GetAllTeas() []teadb.Tea {
	return c.teas.Item
}

// CacheAllTeas caches teas array
func (c *Cache) CacheAllTeas(t []teadb.Tea) {
	c.teas = new(cachedAllTeas)
	c.teas.Valid = true
	c.teas.Item = t
}

// IsTeaValid returns if the given tea cache exist and is valid
func (c *Cache) IsTeaValid(id int) bool {
	tea, ok := c.teaByID[id]
	return ok && tea.Valid
}

// GetTea returns a cached tea
func (c *Cache) GetTea(id int) teadb.Tea {
	return c.teaByID[id].Item
}

// SetTea returns a cached tea
func (c *Cache) SetTea(tea teadb.Tea) {
	ct := new(cachedTea)
	ct.Valid = true
	ct.Item = tea

	c.teaByID[tea.ID] = ct
}

// NewCache creates a new Cache object
func NewCache() (*Cache, error) {
	c := new(Cache)

	c.teas = new(cachedAllTeas)
	c.teas.Valid = false

	c.teaByID = make(map[int]*cachedTea)

	return c, nil
}
