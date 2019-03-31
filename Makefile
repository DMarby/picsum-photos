.PHONY: fmt test vet install integration
all: test vet install

fmt:
	go fmt ./...

test:
	go test ./...

integration:
	go test -tags integration ./...

vet:
	go vet ./...

install:
	go install ./...
