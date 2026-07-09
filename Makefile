.PHONY: build test test-integration test-bdd lint shellcheck update-golden

build:
	go build -o mydots .

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

test-bdd:
	godog ./internal/test/features/...

lint:
	golangci-lint run

shellcheck:
	shellcheck --shell=sh assets/scripts/*.sh

update-golden:
	go test -update ./...
