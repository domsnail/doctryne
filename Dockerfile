# syntax=docker/dockerfile:1

# =============================================================================
# Build stage
# =============================================================================
FROM golang:1.27-alpine AS build

# CGO is required: gorm's sqlite driver links against mattn/go-sqlite3.
RUN apk add --no-cache git gcc musl-dev

WORKDIR /src

# Cache module downloads first so incremental builds stay fast.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Regenerate templ templates. The *_templ.go files are gitignored, so they may
# be missing in a clean clone; this produces them deterministically from the
# .templ sources before compiling.
RUN go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate

RUN CGO_ENABLED=1 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/doctryne .

# =============================================================================
# Runtime stage
# =============================================================================
FROM alpine:3.21

# ca-certificates: outbound HTTPS (GitHub, NVD, npm, GigaChat, …).
# tzdata: timezone-aware scheduling and database timezone handling.
RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/doctryne /usr/local/bin/doctryne

EXPOSE 8080

ENTRYPOINT ["doctryne"]
CMD ["--server", "--host", "0.0.0.0", "--port", "8080"]
