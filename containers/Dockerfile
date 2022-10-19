FROM golang:1.19-alpine as gobuilder

RUN apk add --update --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips-dev && \
    apk add \
      git \
      make \
      gcc \
      musl-dev

ADD . /picsum-photos
WORKDIR /picsum-photos

RUN go install ./...

FROM node:16.14-alpine as nodebuilder

ADD ./web /web
WORKDIR /web

RUN npm ci && npm run-script build


FROM alpine:3.16

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips ca-certificates

WORKDIR /app
COPY --from=gobuilder /go/bin/picsum-photos .
COPY --from=gobuilder /go/bin/image-service .
COPY --from=nodebuilder /web/dist dist
CMD ["./picsum-photos"]
