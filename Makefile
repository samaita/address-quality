.PHONY: run build test lint clean air test-api test-api-smoke test-api-load

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/ tmp/ address.db

air:
	air

test-api: test-api-smoke test-api-load

test-api-smoke:
	k6 run tests/api/smoke-test.js

test-api-load:
	k6 run tests/api/load-test.js
