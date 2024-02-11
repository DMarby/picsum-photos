GO ?= go
GOTOOLRUN = $(GO) run -modfile=./tools/go.mod

.PHONY: test
test:
	$(GO) test ./...

.PHONY: fixtures
fixtures: generate_fixtures
	docker run --rm -v $(PWD):/picsum-photos docker.io/golang:1.19-alpine sh -c 'apk add make && cd /picsum-photos && make docker_fixtures generate_fixtures'

.PHONY: generate_fixtures
generate_fixtures:
	GENERATE_FIXTURES=1 $(GO) test ./... -run '^(TestFixtures)$$'

.PHONY: docker_fixtures
docker_fixtures:
	apk add --update --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing vips-dev
	apk add \
		git \
		gcc \
		musl-dev

.PHONY: generate
generate: go.mod.sri

go.mod.sri: go.mod
	$(GO) mod vendor -o .tmp-vendor
	$(GOTOOLRUN) tailscale.com/cmd/nardump -sri .tmp-vendor >$@
	rm -rf .tmp-vendor

.PHONY: upgrade
upgrade:
# https://github.com/golang/go/issues/28424
	$(GO) list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all | xargs $(GO) get
	$(GO) mod tidy -v

.PHONY: upgradetools
upgradetools:
	cd tools && $(GO) list -e -f '{{range .Imports}}{{.}}@latest {{end}}' -tags tools | xargs $(GO) get
	cd tools && $(GO) mod tidy -v
