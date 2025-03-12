mocks:
	go generate ./...

swag:
	swag init --md ./docs --parseInternal  --parseDependency --parseDepth 2 

.PHONY: mocks swag