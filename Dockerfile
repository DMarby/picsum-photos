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

# Second stage for the frontend
FROM node:10.15.3-alpine as nodebuilder

# Add the project
ADD . /picsum-photos
WORKDIR /picsum-photos

# Install dependencies and run the build
RUN npm install && npm run-script build

# Third stage with only the things needed for the app to run
FROM alpine:3.9

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips ca-certificates

WORKDIR /app
COPY --from=gobuilder /go/bin/picsum-photos .
COPY --from=nodebuilder /picsum-photos/static static
CMD ["./picsum-photos"]
