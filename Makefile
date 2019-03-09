.PHONY: fmt test vet install
all: test vet install

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

install:
	go install ./...
