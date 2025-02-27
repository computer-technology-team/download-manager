SQLC_VERSION := v1.28.0

.PHONY: build
build:
	go build -o download-manager

.PHONY: clean
clean:
	find . -type f -name "*.go"  -exec grep -qE "// Code generated by (mockery|sqlc|protoc-gen-go|protoc-gen-go-kvstore).*\. DO NOT EDIT\." {} \; -delete

generate:  clean gen-sqlc gen-go
	@go mod tidy

gen-go:
	@go generate ./...

gen-sqlc: check-sqlc
	sqlc generate -f ./configs/sqlc.yaml

check-sqlc:
	@if ! command -v sqlc ; then \
		echo "sqlc could not be found"; \
		echo "Installing sqlc $(SQLC_VERSION)"; \
		go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION); \
	elif ! sqlc version 2> /dev/null | grep -q $(SQLC_VERSION); then \
		echo "Incorrect version of sqlc found"; \
		echo "Installing sqlc $(SQLC_VERSION)"; \
		go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION); \
	else \
		echo "Required sqlc version $(SQLC_VERSION) is already installed"; \
	fi
