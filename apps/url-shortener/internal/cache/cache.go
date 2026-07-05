// Package cache é um LRU thread-safe de resoluções slug->destino para o
// caminho quente do redirect. Guarda hits positivos e negativos (404) com TTLs
// distintos, mas nunca a decisão 302/410 — a expiração é sempre recalculada.
package cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

type entry struct {
	slug       string
	res        shortener.Resolution
	insertedAt time.Time
}

// Cache é um LRU com capacidade fixa e TTL por entrada.
type Cache struct {
	mu          sync.Mutex
	capacity    int
	ttl         time.Duration
	negativeTTL time.Duration
	ll          *list.List
	items       map[string]*list.Element
	now         func() time.Time // injetável nos testes
}

// New cria o cache. ttl aplica aos hits positivos; negativeTTL (curto) aos
// negativos, para que um alias recém-criado fique acessível rápido.
func New(capacity int, ttl, negativeTTL time.Duration) *Cache {
	return &Cache{
		capacity:    capacity,
		ttl:         ttl,
		negativeTTL: negativeTTL,
		ll:          list.New(),
		items:       make(map[string]*list.Element),
		now:         time.Now,
	}
}

// Get retorna a resolução e true num hit fresco; false em miss ou entrada
// expirada (que é removida).
func (c *Cache) Get(slug string) (shortener.Resolution, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[slug]
	if !ok {
		return shortener.Resolution{}, false
	}
	e := el.Value.(*entry)

	ttl := c.ttl
	if !e.res.Found {
		ttl = c.negativeTTL
	}
	if c.now().Sub(e.insertedAt) > ttl {
		c.removeElement(el)
		return shortener.Resolution{}, false
	}

	c.ll.MoveToFront(el)
	return e.res, true
}

// Put insere ou atualiza a resolução do slug, evitando a cauda LRU se preciso.
func (c *Cache) Put(slug string, res shortener.Resolution) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[slug]; ok {
		e := el.Value.(*entry)
		e.res = res
		e.insertedAt = c.now()
		c.ll.MoveToFront(el)
		return
	}

	el := c.ll.PushFront(&entry{slug: slug, res: res, insertedAt: c.now()})
	c.items[slug] = el
	if c.capacity > 0 && c.ll.Len() > c.capacity {
		if back := c.ll.Back(); back != nil {
			c.removeElement(back)
		}
	}
}

// Delete remove o slug do cache (defensivo após criar um link).
func (c *Cache) Delete(slug string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[slug]; ok {
		c.removeElement(el)
	}
}

// Len é o número de entradas — usado nos testes.
func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

func (c *Cache) removeElement(el *list.Element) {
	c.ll.Remove(el)
	delete(c.items, el.Value.(*entry).slug)
}
