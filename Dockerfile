# First stage for building the app
FROM golang:1.12-alpine3.9 as gobuilder

# Install libvips
RUN apk add --update --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips-dev && \
    apk add \
      git \
      make \
      gcc \
      musl-dev

# Add the project
ADD . /picsum-photos
WORKDIR /picsum-photos

# Run tests and build
RUN make

# Second stage with only the things needed for the app to run
FROM alpine:3.9

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips

WORKDIR /app
COPY --from=gobuilder /go/bin/picsum-photos .
COPY --from=gobuilder /picsum-photos/static static
CMD ["./picsum-photos"]
