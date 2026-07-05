# Build a partir de apps/url-shortener:
#   docker build -t ccl/url-shortener:dev apps/url-shortener

FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
ENV PORT=8080
# A base GeoLite2 (.mmdb) NÃO é embutida na imagem (licença + tamanho): monte-a
# via volume e aponte GEOIP_DB_PATH. Sem ela, o país fica vazio (degradação graciosa).
COPY --from=builder /out/server /usr/local/bin/server
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/server"]
