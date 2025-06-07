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
	@echo "linting..."
	make lint
	@echo "testing..."
	make test
	@echo "successfully finished"

run:
	go run main.go

.PHONY: mocks swag lint test all run