FROM golang:1.25.6-alpine

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    go install github.com/pressly/goose/v3/cmd/goose@latest

COPY _docker/goose/docker-goose.sh /docker-goose.sh
RUN chmod +x /docker-goose.sh

WORKDIR /app

ENTRYPOINT ["/docker-goose.sh"]