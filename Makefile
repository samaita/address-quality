.PHONY: run build test lint clean air test-api test-api-smoke test-api-load build-seed seed

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/ tmp/ db/*.db

air:
	air

test-api: test-api-smoke test-api-load

test-api-smoke:
	./tests/api/run-k6.sh smoke-test tests/api/smoke-test.js

test-api-load:
	./tests/api/run-k6.sh load-test tests/api/load-test.js

build-seed:
	go build -o bin/seeder ./cmd/seeder

seed:
	go run ./cmd/seeder --drop
