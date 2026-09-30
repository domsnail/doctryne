FROM golang:1.27 AS build

ARG GITLAB_TOKEN
ARG GITLAB_HOST
ARG GITLAB_PROJECT_ID

RUN go env -w GOPRIVATE="${CI_SERVER_HOST}/*" && \
    go env -w GONOSUMDB="${CI_SERVER_HOST}/*" && \
    go env -w GOPROXY="https://${GITLAB_HOST}/api/v4/projects/${GITLAB_PROJECT_ID}/packages/go,https://proxy.golang.org,direct"

RUN --mount=type=secret,id=ci-job-token \
    CI_JOB_TOKEN=$(cat /run/secrets/ci-job-token) && \
    git config --global --unset-all url."https://${CI_SERVER_HOST}".insteadOf || true && \
    git config --global url."https://gitlab-ci-token:${CI_JOB_TOKEN}@${CI_SERVER_HOST}".insteadOf "https://${CI_SERVER_HOST}" && \
    git config --global --unset-all url."https://gitlab-ci-token:${CI_JOB_TOKEN}@${CI_SERVER_HOST}".insteadOf

COPY .. /apps

WORKDIR /apps/pom2sbom
RUN go mod download && go mod verify

RUN apk add --no-cache git gcc musl-dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY ../.. .

RUN go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate

RUN CGO_ENABLED=1 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/doctryne .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

# удаление лишних файлов
RUN rm -rf /var/cache/* /tmp/*

COPY --from=build /out/doctryne /usr/local/bin/doctryne

# защищенные переменные среды
ENV GODEBUG="netdns=go http2server=0"
ENV PATH="/app:${PATH}"

EXPOSE 8080

ENTRYPOINT ["doctryne"]
CMD ["--server", "--host", "0.0.0.0", "--port", "8080"]
