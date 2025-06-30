include .env

DATABASE_DSN=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

LOCAL_BIN := bin
OUT_PATH := pkg
PROTOC_VERSION := 31.1
PROTOC_FILE := protoc-$(PROTOC_VERSION)-win64.zip
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_FILE)
PROTOC := $(LOCAL_BIN)/bin/protoc.exe

bin-deps:
	cmd /C if not exist "$(LOCAL_BIN)" mkdir "$(LOCAL_BIN)"
	curl -LO $(PROTOC_URL)
	powershell -Command "Expand-Archive -Path '$(PROTOC_FILE)' -DestinationPath '$(LOCAL_BIN)' -Force"
	del $(PROTOC_FILE)
	powershell -Command "$$env:GOBIN = Resolve-Path '$(LOCAL_BIN)'; go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
	powershell -Command "$$env:GOBIN = Resolve-Path '$(LOCAL_BIN)'; go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
	powershell -Command "$$env:GOBIN = Resolve-Path '$(LOCAL_BIN)'; go install github.com/envoyproxy/protoc-gen-validate@latest"
	powershell -Command "$$env:GOBIN = Resolve-Path '$(LOCAL_BIN)'; go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest"
	powershell -Command "$$env:GOBIN = Resolve-Path '$(LOCAL_BIN)'; go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest"



generate:
	cmd /C if not exist "$(OUT_PATH)" mkdir "$(OUT_PATH)"
	$(PROTOC) --proto_path=api \
		--proto_path=vendor.protogen \
		--go_out=$(OUT_PATH) \
		--go_opt=paths=source_relative \
		--plugin protoc-gen-go="bin\protoc-gen-go.exe" \
		--go-grpc_out=$(OUT_PATH) \
		--go-grpc_opt=paths=source_relative \
		--plugin protoc-gen-go-grpc="bin\protoc-gen-go-grpc.exe" \
		--validate_out="lang=go,paths=source_relative:$(OUT_PATH)" \
		--plugin protoc-gen-validate="bin\protoc-gen-validate.exe" \
		--grpc-gateway_out=$(OUT_PATH) \
		--grpc-gateway_opt=paths=source_relative \
		--plugin protoc-gen-grpc-gateway="bin\protoc-gen-grpc-gateway.exe" \
		--openapiv2_out=$(OUT_PATH) \
        --plugin=protoc-gen-openapiv2="bin\protoc-gen-openapiv2.exe" \
		api/pwz/pwz.proto
	go mod tidy

vendor-proto/validate:
	@powershell -NoProfile -Command \
	"$$ErrorActionPreference = 'Stop'; \
	if (Test-Path 'vendor.protogen/tmp') { Remove-Item -Recurse -Force 'vendor.protogen/tmp' }; \
	if (Test-Path 'vendor.protogen/validate') { Remove-Item -Recurse -Force 'vendor.protogen/validate' }; \
	git clone -b main --single-branch --depth=2 --filter=tree:0 https://github.com/bufbuild/protoc-gen-validate vendor.protogen/tmp; \
	cd vendor.protogen/tmp; \
	git sparse-checkout set --no-cone validate; \
	git checkout; \
	mkdir -Force '..\validate' | Out-Null; \
	Move-Item -Path 'validate' -Destination '..\validate' -Force; \
	cd ..; \
	Remove-Item -Recurse -Force 'tmp'"

vendor-proto/google-api:
	@powershell -NoProfile -Command \
	"$$ErrorActionPreference = 'Stop'; \
	if (Test-Path 'vendor.protogen/googleapis') { Remove-Item -Recurse -Force 'vendor.protogen/googleapis' }; \
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 https://github.com/googleapis/googleapis vendor.protogen/googleapis; \
	Set-Location vendor.protogen/googleapis; \
	git sparse-checkout set --no-cone google/api; \
	git checkout; \
	Set-Location ../..; \
	if (-not (Test-Path 'vendor.protogen/google')) { New-Item -ItemType Directory -Path 'vendor.protogen/google' | Out-Null }; \
	Move-Item -Path 'vendor.protogen/googleapis/google/api' -Destination 'vendor.protogen/google' -Force; \
	Remove-Item -Recurse -Force 'vendor.protogen/googleapis'"

vendor-proto/openapiv2-options:
	@powershell -NoProfile -Command \
	"$$ErrorActionPreference = 'Stop'; \
	if (Test-Path 'vendor.protogen/tmp') { Remove-Item -Recurse -Force 'vendor.protogen/tmp' }; \
	if (Test-Path 'vendor.protogen/protoc-gen-openapiv2') { Remove-Item -Recurse -Force 'vendor.protogen/protoc-gen-openapiv2' }; \
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 https://github.com/grpc-ecosystem/grpc-gateway vendor.protogen/tmp; \
	Set-Location vendor.protogen/tmp; \
	git sparse-checkout set --no-cone protoc-gen-openapiv2/options; \
	git checkout; \
	New-Item -ItemType Directory -Path '../protoc-gen-openapiv2' -Force | Out-Null; \
	Move-Item -Path 'protoc-gen-openapiv2/options' -Destination '../protoc-gen-openapiv2' -Force; \
	Set-Location ../..; \
	Remove-Item -Recurse -Force 'vendor.protogen/tmp'"

run:
	go run ./cmd/main.go
update:
	go mod tidy
start:
	./homework.exe
build:
	go build -o homework.exe ./cmd
linter:
	golangci-lint run ./...


up:
	docker-compose up -d

down:
	docker-compose down

restart: down up

goose-add:
	goose -dir ./migrations postgres "$(DATABASE_DSN)" create $(NAME) sql

goose-up:
	goose -dir ./migrations postgres "$(DATABASE_DSN)" up

goose-down:
	goose -dir ./migrations postgres "$(DATABASE_DSN)" down

goose-status:
	goose -dir ./migrations postgres "$(DATABASE_DSN)" status

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

e2e:
	go test ./test1_test.go