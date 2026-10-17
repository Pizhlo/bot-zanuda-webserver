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
# Общий runner: исходники зависимостей можно выбрать явно, не редактируя сабмодули.
TESTKIT_DIR ?= $(CUR_DIR)/testkit
PYTESTS_DIR ?= $(abspath $(CUR_DIR)/../pytests)
export PYTESTS_DIR
export INTEGRATION_ROOT := $(CUR_DIR)
export INTEGRATION_CONTEXT := webserver
export INTEGRATION_SERVICE := webserver
export WEBSERVER_MODE WEBSERVER_IMAGE TEST_MODE PYTEST_CONFIG SEED_DATA TESTDATA_PATH TESTDATA_DATABASE_URL

.PHONY: integration-up integration-run integration-seed integration-test integration-down integration-logs integration-rabbitmq-setup
integration-up integration-run integration-seed integration-test integration-down integration-logs integration-rabbitmq-setup:
	@$(MAKE) -C "$(TESTKIT_DIR)" $@

# Только зависимости IDE: процесс приложения запускает разработчик.
.PHONY: containers-up containers-down
containers-up containers-down:
	@$(MAKE) -C "$(TESTKIT_DIR)" webserver-$@

# Короткие алиасы управления зависимостями локального сервиса.
.PHONY: up down
up: containers-up

down: containers-down
