FROM golang:1.27-alpine AS build

RUN apk add --no-cache git gcc musl-dev

WORKDIR /src

COPY ../../go.mod go.sum ./
RUN go mod download

COPY ../.. .

RUN go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate

RUN CGO_ENABLED=1 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/doctryne .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/doctryne /usr/local/bin/doctryne

EXPOSE 8080

ENTRYPOINT ["doctryne"]
CMD ["--server", "--host", "0.0.0.0", "--port", "8080"]
