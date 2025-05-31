
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