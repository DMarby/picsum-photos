.PHONY: fmt test vet install
all: test vet install

fmt:
	go fmt ./...

test:
	go test -short ./...

vet:
	go vet ./...

install:
	go install ./...
