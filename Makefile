.PHONY: build vet lint test check run clean

BIN := bin/codefall

build:
	go build -o $(BIN) ./cmd/codefall

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./...

# What CI runs. Both halves of ADR-GO-02's verification are covered here: the compiler catches
# facade reach-arounds during `build`, depguard catches outward layer imports during `lint`.
check: build vet lint test

run:
	go run ./cmd/codefall

clean:
	rm -rf bin
