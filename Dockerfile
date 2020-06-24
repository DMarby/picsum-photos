# First stage for building the app
FROM golang:1.14-alpine as gobuilder

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
FROM node:12.16-alpine as nodebuilder

# Add the project
ADD . /picsum-photos
WORKDIR /picsum-photos

# Install dependencies and run the build
RUN npm install && npm run-script build

# Third stage with only the things needed for the app to run
FROM alpine:3.11

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips ca-certificates

WORKDIR /app
COPY --from=gobuilder /go/bin/picsum-photos .
COPY --from=gobuilder /go/bin/image-service .
COPY --from=gobuilder /picsum-photos/migrations migrations
COPY --from=nodebuilder /picsum-photos/dist dist
CMD ["./picsum-photos"]
