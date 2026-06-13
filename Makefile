.PHONY: build test fmt lint ci clean docs-commands docs-test docs-site docs-site-test docs-site-clean gobankcli

build:
	go build -o bin/gobankcli ./cmd/gobankcli

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

lint:
	go vet ./...

ci: fmt lint docs-test docs-site-test test

docs-commands:
	go run ./cmd/gobankcli docs-command-reference | node scripts/gen-command-reference.mjs

docs-test:
	node --test scripts/*.test.mjs

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
