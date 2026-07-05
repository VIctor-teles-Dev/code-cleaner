# url-shortener: imagem a partir do binário Go pré-compilado (bin/server).
# O binário é dinâmico (glibc, CGO p/ getaddrinfo) e exige no máximo GLIBC 2.34,
# então a base precisa de glibc >= 2.34 — debian:bookworm-slim (2.36) atende.
FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY bin/server /app/server
EXPOSE 8080
USER nobody
ENTRYPOINT ["/app/server"]
