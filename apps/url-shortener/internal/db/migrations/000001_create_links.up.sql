-- Links encurtados: slug único (alias custom ou base62 gerado). O UNIQUE gera o
-- índice usado no lookup do redirect e a violação 23505 que dirige 409/retry.
CREATE TABLE links (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ  -- nulo = nunca expira
);
