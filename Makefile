mocks:
	go generate ./...

swag:
	swag init --md ./docs --parseInternal  --parseDependency --parseDepth 2 

lint:
	go vet ./...
	staticcheck ./...

test:
	go test -gcflags="-l" -race -v ./...

all:
    make lint
	make test

.PHONY: mocks swag lint test all