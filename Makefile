.PHONY: run build test lint clean air swagger test-api test-api-smoke test-api-load test-api-load-prod build-seed seed seed-init seed-drop seed-truncate seed-normalize benchmark benchmark-page benchmark-v0 benchmark-page-v0

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

test-api-load-prod:
	K6_BASE_URL=https://api.samaita.com/address-quality ./tests/api/run-k6.sh load-test tests/api/load-test.js

build-seed:
	go build -o bin/seeder ./cmd/seeder

# Seed into existing tables (update data). Pass seeder flags via ARGS, e.g.
#   make seed ARGS=--init
# Or use the dedicated targets below.
seed:
	go run ./cmd/seeder $(ARGS)

seed-init:
	go run ./cmd/seeder --init

seed-drop:
	go run ./cmd/seeder --drop

seed-truncate:
	go run ./cmd/seeder --truncate

seed-normalize:
	go run ./cmd/seeder --normalize

swagger:
	swag init -g cmd/server/main.go -o docs

benchmark:
	node tests/api/benchmark-test.js --source=kemendagri --csv=tests/api/cases/address-tagged.csv

benchmark-page:
	@printf 'benchmark_build: '; read b; [ -n "$$b" ] || { echo 'error: benchmark_build is required'; exit 1; }; node tests/api/page/build.js "$$b"

benchmark-v0:
	node tests/api/benchmark-test-v0.js --source=kemendagri --csv=tests/api/cases/address-tagged.csv

benchmark-page-v0:
	@printf 'benchmark_build: '; read b; [ -n "$$b" ] || { echo 'error: benchmark_build is required'; exit 1; }; BENCH_VER=v0 node tests/api/page/build.js "$$b"
