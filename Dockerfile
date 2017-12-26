FROM golang:alpine

# TODO: Can we get rid of pkgconf?
# TODO: Do we need these extra packages for this or just for debug?
# Install libvips
RUN apk add --update --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips-dev && \
    apk add \
          git \
          pkgconf \
          fftw-dev

# Install needed tools
RUN go get \
      github.com/golang/dep/cmd/dep \
      github.com/onsi/ginkgo/ginkgo

# Add the project
ADD . /go/src/github.com/DMarby/picsum-photos
WORKDIR /go/src/github.com/DMarby/picsum-photos

# Install dependencies
RUN dep ensure

# Run tests
RUN ginkgo -r -p -noColor

# Build
RUN go install

# Copy binary
RUN cp /go/bin/picsum-photos /usr/bin/picsum-photos

# Clean up
RUN apk del git &&\
    rm -rf \
      /var/cache/apk/* \
      $GOPATH

CMD ["/usr/bin/picsum-photos"]
# TODO: Multi-stage?
