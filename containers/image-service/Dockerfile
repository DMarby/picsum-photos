FROM docker.io/golang:1.19-alpine as gobuilder

RUN apk add --update --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips-dev && \
    apk add \
      pkgconfig \
      gcc \
      musl-dev

COPY go.mod go.sum /app/
WORKDIR /app
RUN --mount=type=cache,target=/go/pkg/mod go mod download -x

ADD . /app

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install ./cmd/image-service

FROM docker.io/alpine:3.16

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips ca-certificates

WORKDIR /app
COPY --from=gobuilder /go/bin/image-service .
CMD ["./image-service"]
