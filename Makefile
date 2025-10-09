MAKEFILE_PATH := $(abspath $(firstword $(MAKEFILE_LIST)))
CUR_DIR := $(patsubst %/,%, $(dir $(MAKEFILE_PATH)))
BUILD_DIR := $(CUR_DIR)/.build
APP_EXECUTABLE_DIR := $(BUILD_DIR)/bin

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

build:
	@echo " > building..."
	@mkdir -p "$(BUILD_DIR)/bin"
	@VERSION=$$(git describe --tags --always --dirty); \
	BUILD_DATE=$$(date -u +%Y%m%d-%H%M%SZ); \
	GIT_COMMIT=$$(git rev-parse --short HEAD); \
	go build -trimpath \
	-ldflags "-s -w -X main.Version=$$VERSION -X main.BuildDate=$$BUILD_DATE -X main.GitCommit=$$GIT_COMMIT" \
	-o "$(BUILD_DIR)/bin/" ./cmd/...
	@echo " > successfully built"

run:
	@make build
	$(APP_EXECUTABLE_DIR)/app

image:
	docker build -f Dockerfile -t pizhlo/bot-zanuda-webserver:latest .

push:
	docker push pizhlo/bot-zanuda-webserver:latest

.PHONY: mocks swag lint test all run image push