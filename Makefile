.PHONY: test integration package publish fixtures generate_fixtures docker_fixtures

test:
	go test ./...

integration:
	go test -tags integration ./...

integration_services:
	docker run --rm -p 6380:6379 redis

package:
	docker build . -f containers/Dockerfile -t registry.digitalocean.com/picsum-registry/picsum-photos:latest

publish: package
	docker push registry.digitalocean.com/picsum-registry/picsum-photos:latest

fixtures: generate_fixtures
	docker run --rm -v $(PWD):/picsum-photos docker.io/golang:1.19-alpine sh -c 'apk add make && cd /picsum-photos && make docker_fixtures generate_fixtures'

generate_fixtures:
	GENERATE_FIXTURES=1 go test ./... -run '^(TestFixtures)$$'

docker_fixtures:
	apk add --update --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips-dev
	apk add \
		git \
		gcc \
		musl-dev
