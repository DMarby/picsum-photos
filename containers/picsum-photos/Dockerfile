FROM docker.io/golang:1.19-alpine as gobuilder

ADD https://github.com/tailwindlabs/tailwindcss/releases/download/v3.2.4/tailwindcss-linux-x64 /usr/bin/tailwindcss
RUN chmod +x /usr/bin/tailwindcss

COPY go.mod go.sum /app/
WORKDIR /app
RUN --mount=type=cache,target=/go/pkg/mod go mod download -x

ADD . /app

RUN tailwindcss -c ./internal/api/web/tailwind.config.js -i ./internal/api/web/style.css -o ./internal/api/web/embed/assets/css/style.css --minify
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install ./cmd/picsum-photos

FROM docker.io/alpine:3.16

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=gobuilder /go/bin/picsum-photos .
CMD ["./picsum-photos"]
