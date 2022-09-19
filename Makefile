.PHONY: fmt test vet install integration package publish fixtures generate_fixtures docker_fixtures
all: test vet install

fmt:
	go fmt ./...

test:
	go test ./...

integration:
	go test -tags integration ./...

integration_services:
	docker run --rm -p 5433:5432 -e POSTGRES_PASSWORD=postgres postgres & \
	docker run --rm -p 6380:6379 redis

vet:
	go vet ./...

install:
	go install ./...

package:
	docker build . -t dmarby/picsum-photos:latest

publish: package
	docker push dmarby/picsum-photos:latest

fixtures: generate_fixtures
	docker run --rm -v $(PWD):/picsum-photos golang:1.14-alpine sh -c 'apk add make && cd /picsum-photos && make docker_fixtures generate_fixtures'

generate_fixtures:
	GENERATE_FIXTURES=1 go test ./... -run '^(TestFixtures)$$'

docker_fixtures:
	apk add --update --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips-dev
	apk add \
		git \
		gcc \
		musl-dev
