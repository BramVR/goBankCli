.PHONY: build test fmt lint ci clean docs-site docs-site-test docs-site-clean gobankcli

build:
	go build -o bin/gobankcli ./cmd/gobankcli

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

lint:
	go vet ./...

ci: fmt lint test

clean:
	rm -rf bin

docs-site:
	node scripts/build-docs-site.mjs

docs-site-test:
	node --test scripts/build-docs-site.test.mjs

docs-site-clean:
	rm -rf dist/docs-site

gobankcli: build
	./bin/gobankcli $(filter-out $@,$(MAKECMDGOALS))

%:
	@:
