# ─── Stage 1: build ──────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Instala dependencias del sistema necesarias para cgo (lib/pq lo requiere)
RUN apk add --no-cache git ca-certificates tzdata

# Copia archivos de workspace y módulos primero para aprovechar la caché de capas
COPY go.work go.work.sum ./
COPY go.mod go.sum ./

# Copia go.mod de cada submódulo pkg/
COPY pkg/config/go.mod       pkg/config/
COPY pkg/constants/go.mod    pkg/constants/
COPY pkg/databases/go.mod    pkg/databases/
COPY pkg/errors/go.mod       pkg/errors/
COPY pkg/http_status/go.mod  pkg/http_status/
COPY pkg/logging/go.mod      pkg/logging/
COPY pkg/openapi/go.mod      pkg/openapi/
COPY pkg/pagination/go.mod   pkg/pagination/
COPY pkg/responses/go.mod    pkg/responses/
COPY pkg/validators/go.mod   pkg/validators/

# Descarga dependencias (se cachean si los .mod no cambian)
RUN go work sync && go mod download

# Copia el resto del código fuente
COPY . .

# Compila el binario con optimizaciones para producción
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -o /app/bin/docsy-main \
    ./cmd/

# ─── Stage 2: runtime ────────────────────────────────────────────────────────
FROM scratch

# Certificados TLS y zona horaria desde el builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Binario compilado
COPY --from=builder /app/bin/docsy-main /docsy-main

# Migraciones SQL incluidas en la imagen
COPY --from=builder /app/migrations /migrations

# Usuario no-root para producción (uid 65532 = nonroot en distroless/scratch)
USER 65532:65532

EXPOSE 8100

ENTRYPOINT ["/docsy-main"]
