package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

// ClickStore implementa shortener.ClickStore (escrita em lote) e
// shortener.AnalyticsStore (leituras agregadas) sobre Postgres.
type ClickStore struct {
	DB *sql.DB
}

const clickCols = 8

// InsertClicks grava os eventos num único INSERT multi-linha.
func (s ClickStore) InsertClicks(ctx context.Context, events []shortener.ClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(`INSERT INTO clicks ` +
		`(slug, clicked_at, ip, country, user_agent, browser, device, referrer) VALUES `)
	args := make([]any, 0, len(events)*clickCols)
	for i, e := range events {
		if i > 0 {
			b.WriteString(", ")
		}
		n := i * clickCols
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8)
		args = append(args, e.Slug, e.ClickedAt, e.IP, e.Country,
			e.UserAgent, e.Browser, e.Device, e.Referrer)
	}

	_, err := s.DB.ExecContext(ctx, b.String(), args...)
	return err
}

// Stats agrega as métricas de clique de um slug. Todas as queries são
// escopadas por slug e servidas pelo índice (slug, clicked_at DESC).
func (s ClickStore) Stats(ctx context.Context, slug string) (shortener.Analytics, error) {
	a := shortener.Analytics{
		Slug:         slug,
		TimeSeries:   []shortener.DayCount{},
		TopCountries: []shortener.LabelCount{},
		TopReferrers: []shortener.LabelCount{},
		Browsers:     []shortener.LabelCount{},
		Devices:      []shortener.LabelCount{},
	}

	if err := s.DB.QueryRowContext(ctx,
		`SELECT count(*) FROM clicks WHERE slug = $1`, slug).Scan(&a.TotalClicks); err != nil {
		return shortener.Analytics{}, err
	}

	series, err := s.timeSeries(ctx, slug)
	if err != nil {
		return shortener.Analytics{}, err
	}
	a.TimeSeries = series

	for _, dim := range []struct {
		dest  *[]shortener.LabelCount
		query string
	}{
		{&a.TopCountries, `SELECT country, count(*) c FROM clicks
		   WHERE slug = $1 AND country <> '' GROUP BY country ORDER BY c DESC LIMIT 10`},
		{&a.TopReferrers, `SELECT referrer, count(*) c FROM clicks
		   WHERE slug = $1 AND referrer <> '' GROUP BY referrer ORDER BY c DESC LIMIT 10`},
		{&a.Browsers, `SELECT browser, count(*) c FROM clicks
		   WHERE slug = $1 AND browser <> '' GROUP BY browser ORDER BY c DESC`},
		{&a.Devices, `SELECT device, count(*) c FROM clicks
		   WHERE slug = $1 AND device <> '' GROUP BY device ORDER BY c DESC`},
	} {
		list, err := s.labelCounts(ctx, dim.query, slug)
		if err != nil {
			return shortener.Analytics{}, err
		}
		*dim.dest = list
	}

	return a, nil
}

func (s ClickStore) timeSeries(ctx context.Context, slug string) ([]shortener.DayCount, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT date_trunc('day', clicked_at) AS day, count(*)
		   FROM clicks
		  WHERE slug = $1 AND clicked_at >= now() - interval '30 days'
		  GROUP BY day ORDER BY day`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []shortener.DayCount{}
	for rows.Next() {
		var d shortener.DayCount
		if err := rows.Scan(&d.Day, &d.Count); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s ClickStore) labelCounts(ctx context.Context, query, slug string) ([]shortener.LabelCount, error) {
	rows, err := s.DB.QueryContext(ctx, query, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []shortener.LabelCount{}
	for rows.Next() {
		var lc shortener.LabelCount
		if err := rows.Scan(&lc.Label, &lc.Count); err != nil {
			return nil, err
		}
		out = append(out, lc)
	}
	return out, rows.Err()
}
