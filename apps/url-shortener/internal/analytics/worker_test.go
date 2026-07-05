package analytics

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

type fakeClickStore struct {
	mu       sync.Mutex
	inserted []shortener.ClickEvent
}

func (f *fakeClickStore) InsertClicks(_ context.Context, events []shortener.ClickEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.inserted = append(f.inserted, events...)
	return nil
}

func (f *fakeClickStore) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.inserted)
}

func (f *fakeClickStore) first() shortener.ClickEvent {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.inserted[0]
}

func TestWorkerFlushesOnShutdown(t *testing.T) {
	store := &fakeClickStore{}
	w := NewWorker(store, nil) // geo nil -> país vazio
	go w.Run()

	for i := 0; i < 10; i++ {
		w.Enqueue(shortener.ClickEvent{Slug: "s", UserAgent: "curl/8", IP: "1.2.3.4"})
	}
	w.Shutdown() // fecha o canal e faz o flush final do batch parcial

	if store.count() != 10 {
		t.Fatalf("inseriu %d eventos, want 10", store.count())
	}
	// O enriquecimento (parse de UA) aconteceu no worker.
	if got := store.first(); got.Browser == "" || got.Device == "" {
		t.Errorf("evento não enriquecido: browser=%q device=%q", got.Browser, got.Device)
	}
}

func TestEnqueueDropsWhenFull(t *testing.T) {
	w := NewWorker(&fakeClickStore{}, nil)
	// Não inicia Run(): o canal (buffer bufferSize) nunca é drenado.
	for i := 0; i < bufferSize; i++ {
		w.Enqueue(shortener.ClickEvent{Slug: "s"})
	}

	// Com o buffer cheio, os próximos Enqueue descartam sem bloquear.
	done := make(chan struct{})
	go func() {
		w.Enqueue(shortener.ClickEvent{Slug: "overflow"})
		w.Enqueue(shortener.ClickEvent{Slug: "overflow"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Enqueue bloqueou com o buffer cheio")
	}
	if w.Dropped() < 2 {
		t.Errorf("Dropped = %d, want >= 2", w.Dropped())
	}
}

func TestWorkerTickerFlush(t *testing.T) {
	store := &fakeClickStore{}
	w := NewWorker(store, nil)
	go w.Run()
	defer w.Shutdown()

	w.Enqueue(shortener.ClickEvent{Slug: "s"}) // batch parcial (< batchSize)

	deadline := time.Now().Add(3 * flushInterval)
	for time.Now().Before(deadline) {
		if store.count() >= 1 {
			return // o ticker fez o flush
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("ticker não fez flush do batch parcial (count=%d)", store.count())
}
