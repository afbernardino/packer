.PHONY:fmt
fmt:
	go fmt ./...

.PHONY:vet
vet: fmt
	go vet ./...

.PHONY:build
build: vet
	go build -trimpath -v -o build/api ./cmd/main.go

.PHONY:test
test:
	go test ./...

.PHONY:integration_test
integration_test:
	go test -tags integration_test ./...
