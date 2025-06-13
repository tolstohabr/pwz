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

generate:
	cmd /C if not exist "$(OUT_PATH)" mkdir "$(OUT_PATH)"
	$(PROTOC) --proto_path=api --go_out=$(OUT_PATH) --go_opt=paths=source_relative --plugin protoc-gen-go="bin\protoc-gen-go.exe" --go-grpc_out=$(OUT_PATH) --go-grpc_opt=paths=source_relative --plugin protoc-gen-go-grpc="bin\protoc-gen-go-grpc.exe" api/order/order.proto
	go mod tidy


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