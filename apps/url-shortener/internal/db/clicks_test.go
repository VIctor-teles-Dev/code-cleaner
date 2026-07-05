package db

import (
	"context"
	"testing"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

func TestClickStoreInsertAndStats(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	if _, err := pool.ExecContext(ctx, "TRUNCATE clicks, links RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("TRUNCATE: %v", err)
	}

	links := LinkStore{DB: pool}
	if err := links.Create(ctx, shortener.Link{Slug: "abc", OriginalURL: "https://go.dev"}); err != nil {
		t.Fatalf("Create link: %v", err)
	}

	clicks := ClickStore{DB: pool}
	now := time.Now().UTC()
	events := []shortener.ClickEvent{
		{Slug: "abc", ClickedAt: now, Country: "BR", Browser: "Chrome", Device: "Desktop", Referrer: "https://t.co"},
		{Slug: "abc", ClickedAt: now, Country: "BR", Browser: "Firefox", Device: "Mobile", Referrer: "https://t.co"},
		{Slug: "abc", ClickedAt: now, Country: "US", Browser: "Chrome", Device: "Desktop"},
	}
	if err := clicks.InsertClicks(ctx, events); err != nil {
		t.Fatalf("InsertClicks: %v", err)
	}

	stats, err := clicks.Stats(ctx, "abc")
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalClicks != 3 {
		t.Errorf("TotalClicks = %d, want 3", stats.TotalClicks)
	}
	if len(stats.TopCountries) != 2 || stats.TopCountries[0].Label != "BR" || stats.TopCountries[0].Count != 2 {
		t.Errorf("TopCountries = %+v, want BR=2 no topo", stats.TopCountries)
	}
	if len(stats.TimeSeries) == 0 {
		t.Error("TimeSeries vazio")
	}
	for _, r := range stats.TopReferrers {
		if r.Label == "" {
			t.Error("referrer vazio não deveria aparecer no top")
		}
	}
}

func TestInsertClicksEmpty(t *testing.T) {
	pool := testPool(t)
	if err := (ClickStore{DB: pool}).InsertClicks(context.Background(), nil); err != nil {
		t.Errorf("InsertClicks(nil) = %v, want nil", err)
	}
}

func TestInsertClicksForeignKey(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	if _, err := pool.ExecContext(ctx, "TRUNCATE clicks, links RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("TRUNCATE: %v", err)
	}
	// Clique para um slug inexistente viola a FK.
	err := (ClickStore{DB: pool}).InsertClicks(ctx,
		[]shortener.ClickEvent{{Slug: "fantasma", ClickedAt: time.Now().UTC()}})
	if err == nil {
		t.Error("InsertClicks para slug inexistente deveria falhar (FK)")
	}
}
