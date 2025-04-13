FROM golang:1.23.4-alpine

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    go install github.com/vektra/mockery/v2@latest

WORKDIR /app

ENTRYPOINT ["mockery"]