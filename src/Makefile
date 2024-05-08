BINARY_NAME := menucko
BINARY_PATH := ../tmp/${BINARY_NAME}


.PHONY: tidy
tidy:
	go mod tidy -v


.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...


.PHONY: build
build:
	go build -o=${BINARY_PATH}


.PHONY: run
run: build
	${BINARY_PATH}
