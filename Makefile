.PHONY: build test fmt lint ci clean gobankcli

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

gobankcli: build
	./bin/gobankcli $(filter-out $@,$(MAKECMDGOALS))

%:
	@:
