package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/analytics"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/cache"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/db"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/geoip"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/handler"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

func main() {
	port := getenv("PORT", "8080")
	baseURL := os.Getenv("BASE_URL") // prefixo do short_url; vazio -> retorna /slug

	var linkStore shortener.Store
	var statsStore shortener.AnalyticsStore
	var clickStore shortener.ClickStore
	var pinger handler.Pinger
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		pool, err := db.Open(databaseURL)
		if err != nil {
			log.Fatalf("invalid DATABASE_URL: %v", err)
		}
		defer pool.Close()

		if err := db.Migrate(pool); err != nil {
			log.Fatalf("migrations: %v", err)
		}
		log.Print("database migrations applied")

		pinger = pool
		linkStore = db.LinkStore{DB: pool}
		clicks := db.ClickStore{DB: pool}
		statsStore = clicks
		clickStore = clicks
	} else {
		log.Print("DATABASE_URL not set; redirect/analytics disabled, /readyz won't check the database")
	}

	geo := geoip.FromEnv()
	if geo != nil {
		log.Print("GeoIP habilitado; país resolvido no worker de analytics")
		defer geo.Close()
	} else {
		log.Print("GeoIP desabilitado (GEOIP_DB_PATH ausente ou arquivo não encontrado)")
	}

	adminToken := os.Getenv("URL_SHORTENER_ADMIN_TOKEN")
	if adminToken == "" {
		log.Print("URL_SHORTENER_ADMIN_TOKEN não definido; POST /api/v1/urls e analytics desabilitados")
	}

	linkCache := cache.New(10_000, 5*time.Minute, 30*time.Second)

	worker := analytics.NewWorker(clickStore, geo)
	go worker.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Health)
	mux.Handle("GET /readyz", handler.Ready(pinger))
	mux.Handle("POST /api/v1/urls", handler.CreateURL(linkStore, linkCache, adminToken, baseURL))
	mux.Handle("GET /api/v1/analytics/{slug}", handler.GetAnalytics(linkStore, statsStore, adminToken))
	// Catch-all do redirect: a rota literal (/healthz, /api/...) sempre vence o
	// wildcard {slug} na precedência do ServeMux (Go 1.22).
	mux.Handle("GET /{slug}", handler.Redirect(linkStore, linkCache, worker))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("url-shortener listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Print("shutting down...")

	// Para de aceitar requests e deixa as em andamento terminarem, depois fecha
	// o canal do worker e espera o flush final do batch pendente.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
	worker.Shutdown()
	log.Print("bye")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
