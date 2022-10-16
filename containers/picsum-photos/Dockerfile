FROM golang:1.19-alpine as gobuilder

RUN apk add \
      git \
      make \
      gcc \
      musl-dev

ADD . /picsum-photos
WORKDIR /picsum-photos

RUN go install ./cmd/picsum-photos


FROM node:16.14-alpine as nodebuilder

ADD ./web /web
WORKDIR /web

RUN npm ci && npm run-script build


FROM alpine:3.16

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing ca-certificates

WORKDIR /app
COPY --from=gobuilder /go/bin/picsum-photos .
COPY --from=nodebuilder /web/dist dist
CMD ["./picsum-photos"]
