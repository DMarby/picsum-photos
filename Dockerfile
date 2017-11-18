FROM golang:alpine

# Install libvips
RUN apk add --update --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips-tools && \
    apk add git

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
