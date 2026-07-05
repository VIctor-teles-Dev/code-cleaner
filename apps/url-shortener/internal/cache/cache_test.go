package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

func TestCacheHitMiss(t *testing.T) {
	c := New(10, time.Minute, time.Minute)
	if _, ok := c.Get("x"); ok {
		t.Fatal("miss esperado em cache vazio")
	}
	c.Put("x", shortener.Resolution{OriginalURL: "https://go.dev", Found: true})
	res, ok := c.Get("x")
	if !ok || res.OriginalURL != "https://go.dev" {
		t.Fatalf("Get = %+v, %v; want hit", res, ok)
	}
}

func TestCacheEviction(t *testing.T) {
	c := New(2, time.Minute, time.Minute)
	c.Put("a", shortener.Resolution{Found: true})
	c.Put("b", shortener.Resolution{Found: true})
	c.Get("a")                                    // "a" vira recente; "b" vira a cauda LRU
	c.Put("c", shortener.Resolution{Found: true}) // deve evictar "b"

	if _, ok := c.Get("b"); ok {
		t.Error("b deveria ter sido evictado")
	}
	if _, ok := c.Get("a"); !ok {
		t.Error("a deveria continuar no cache")
	}
	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2", c.Len())
	}
}

func TestCacheTTL(t *testing.T) {
	c := New(10, time.Minute, 30*time.Second)
	now := time.Now()
	c.now = func() time.Time { return now }

	c.Put("pos", shortener.Resolution{Found: true})
	c.Put("neg", shortener.Resolution{Found: false})

	now = now.Add(45 * time.Second) // passou do negativeTTL (30s), não do ttl (1min)
	if _, ok := c.Get("neg"); ok {
		t.Error("entrada negativa deveria expirar em 30s")
	}
	if _, ok := c.Get("pos"); !ok {
		t.Error("entrada positiva não deveria expirar em 45s")
	}

	now = now.Add(30 * time.Second) // total 75s > ttl (60s)
	if _, ok := c.Get("pos"); ok {
		t.Error("entrada positiva deveria expirar após o ttl")
	}
}

func TestCacheDelete(t *testing.T) {
	c := New(10, time.Minute, time.Minute)
	c.Put("x", shortener.Resolution{Found: true})
	c.Delete("x")
	if _, ok := c.Get("x"); ok {
		t.Error("x deveria ter sido removido")
	}
}

// TestCacheConcurrent roda sob -race para pegar corridas no LRU.
func TestCacheConcurrent(t *testing.T) {
	c := New(100, time.Minute, time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			slug := fmt.Sprintf("s%d", i%10)
			c.Put(slug, shortener.Resolution{OriginalURL: "u", Found: true})
			c.Get(slug)
			if i%3 == 0 {
				c.Delete(slug)
			}
		}(i)
	}
	wg.Wait()
}
