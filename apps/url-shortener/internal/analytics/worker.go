// Package analytics coleta os cliques de forma assíncrona: o redirect apenas
// enfileira (não-bloqueante) e um worker em background resolve GeoIP + user
// agent e grava em lote no Postgres, sem atrasar o redirecionamento.
package analytics

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/geoip"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/ua"
)

const (
	bufferSize    = 10_000
	batchSize     = 100
	flushInterval = 2 * time.Second
)

// Worker recebe ClickEvents por um canal bufferizado e os grava em lote.
type Worker struct {
	store   shortener.ClickStore
	geo     *geoip.Resolver
	ch      chan shortener.ClickEvent
	done    chan struct{}
	dropped atomic.Int64
}

// NewWorker cria o worker. store nil desabilita a gravação (eventos são
// descartados) — mantém o serviço de pé mesmo sem banco.
func NewWorker(store shortener.ClickStore, geo *geoip.Resolver) *Worker {
	return &Worker{
		store: store,
		geo:   geo,
		ch:    make(chan shortener.ClickEvent, bufferSize),
		done:  make(chan struct{}),
	}
}

// Enqueue tenta enfileirar o evento sem bloquear. Canal cheio -> descarta e
// loga esporadicamente; o redirect NUNCA espera pela analytics.
func (w *Worker) Enqueue(e shortener.ClickEvent) {
	select {
	case w.ch <- e:
	default:
		if n := w.dropped.Add(1); n%1000 == 1 {
			log.Printf("analytics: buffer cheio, evento descartado (total=%d)", n)
		}
	}
}

// Dropped é a contagem de eventos descartados por buffer cheio (para métricas/testes).
func (w *Worker) Dropped() int64 { return w.dropped.Load() }

// Run processa o canal até ele fechar, fazendo flush por tamanho ou por tempo.
func (w *Worker) Run() {
	defer close(w.done)
	batch := make([]shortener.ClickEvent, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case e, ok := <-w.ch:
			if !ok {
				w.flush(batch)
				return
			}
			batch = append(batch, e)
			if len(batch) >= batchSize {
				w.flush(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				w.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

// Shutdown fecha o canal e espera o flush final do batch pendente.
func (w *Worker) Shutdown() {
	close(w.ch)
	<-w.done
}

func (w *Worker) flush(batch []shortener.ClickEvent) {
	if len(batch) == 0 || w.store == nil {
		return
	}

	// Enriquecimento (GeoIP + user agent) acontece aqui, fora do redirect.
	enriched := make([]shortener.ClickEvent, len(batch))
	copy(enriched, batch)
	for i := range enriched {
		enriched[i].Country = w.geo.Country(enriched[i].IP)
		enriched[i].Browser, enriched[i].Device = ua.Parse(enriched[i].UserAgent)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.store.InsertClicks(ctx, enriched); err != nil {
		log.Printf("analytics: falha ao gravar lote de %d cliques: %v", len(enriched), err)
	}
}
