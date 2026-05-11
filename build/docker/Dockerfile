LABEL authors="lysak.yaroslav00@yandex.ru"

FROM golang:1.24.3-alpine3.21 as modules

COPY ../go.mod go.sum /modules/

WORKDIR /modules

RUN go mod download

FROM golang:1.24.3-alpine3.21 as builder

COPY --from=modules /go/pkg /go/pkg
COPY .. /app

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /bin/app ./cmd/app

FROM scratch

COPY --from=builder /app/cfg /cfg
COPY --from=builder /bin/app /app

CMD ["/app"]
