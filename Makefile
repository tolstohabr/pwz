LOCAL_BIN := bin
OUT_PATH := pkg
PROTOC_VERSION := 31.1
PROTOC_FILE := protoc-$(PROTOC_VERSION)-win64.zip
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_FILE)
PROTOC := $(LOCAL_BIN)/bin/protoc.exe

bin-deps:
	curl -LO $(PROTOC_URL)
	powershell -Command "Expand-Archive -Path '$(PROTOC_FILE)' -DestinationPath '$(LOCAL_BIN)' -Force"
	del $(PROTOC_FILE)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest

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