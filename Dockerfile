ARG IMAGE_PREFIX=docker.io/library

FROM ${IMAGE_PREFIX}/golang:1.27-alpine3.24 AS build

ARG GOPROXY="https://proxy.golang.org,direct"
ENV GOPROXY="${GOPROXY}"

RUN apk add --no-cache git build-base musl-dev && \
    rm -rf /var/cache/* /tmp/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ../.. .

RUN go tool templ generate

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc \
    go build -trimpath -tags="netgo,osusergo" -ldflags="-s -w -linkmode external -extldflags '-static'" -o /out/doctryne .

FROM ${IMAGE_PREFIX}/alpine:3.24

RUN apk add --no-cache ca-certificates tzdata && rm -rf /var/cache/* /tmp/*

COPY --from=build /out/doctryne /usr/local/bin/doctryne

# защищенные переменные среды
ENV GODEBUG="netdns=go http2server=0"
ENV PATH="/usr/local/bin:${PATH}"

EXPOSE 8080

ENTRYPOINT ["doctryne"]
CMD ["--server", "--use-secrets", "--host", "0.0.0.0", "--port", "8080"]
