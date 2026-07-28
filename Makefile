all: test-unit gen-tprl build gen verify-examples

.PHONY: gen-tprl
gen-tprl: build
	@PATH="$$PWD/bin:$$PATH" buf generate --path proto/temporal/v1/temporal.proto

.PHONY: build
build:
	@echo "Building plugin binary..."
	@if ! [ -d bin ]; then mkdir bin; fi
	@go build -o bin/protoc-gen-go-tmprl

.PHONY: gen
gen: build
	@echo "Regenerating example code..."
	@PATH="$$PWD/bin:$$PATH" buf generate --path example/proto/example

.PHONY: verify-examples
verify-examples:
	@echo "Building and verifying generated example code..."
	@if ! [ -d bin ]; then mkdir bin; fi
	@go build -o bin/example-worker ./example/worker
	@go build -o bin/example-client ./example/client
	@echo "All examples built successfully in bin/"

.PHONY: test
test: test-unit

.PHONY: test-e2e
test-e2e:
	@echo "Starting the temporal dev stack..."
	@docker compose up -d --wait
	@echo "Running e2e tests against the dev stack..."
	@go test -race -count=1 -tags e2e -v ./example/e2e/

.PHONY: test-unit
test-unit:
	@echo "Running unit tests with race detection and coverage..."
	@go test -race -cover ./internal/... ./gen/... ./example/...

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin
	@echo "All build artifacts removed"

.PHONY: bufpush
bufpush:
	buf push proto
