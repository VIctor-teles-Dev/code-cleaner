-- Eventos de clique (série temporal). Referencia links pelo slug para o worker
-- inserir em lote sem precisar traduzir slug -> id no caminho de escrita.
-- country/browser/device são preenchidos pelo worker; podem ficar vazios.
CREATE TABLE clicks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug TEXT NOT NULL REFERENCES links (slug) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ip TEXT,
    country TEXT,
    user_agent TEXT,
    browser TEXT,
    device TEXT,
    referrer TEXT
);

-- Analytics agregam por slug e por tempo: um índice composto cobre total,
-- série temporal e todos os "top N" (WHERE slug = $1 GROUP BY ...).
CREATE INDEX clicks_slug_clicked_at_idx ON clicks (slug, clicked_at DESC);
