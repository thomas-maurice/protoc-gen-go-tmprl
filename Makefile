all: gen-tprl build gen

.PHONY: gen-tprl
gen-tprl:
	buf generate --path proto/temporal

.PHONY: build
build:
	if ! [ -d bin ]; then mkdir bin; fi
	go build -o bin/protoc-gen-go-tmprl

.PHONY: gen
gen: build
	buf generate --path example/proto/example
