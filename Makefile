.PHONY: run build test lint clean air

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
