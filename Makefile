SQLC ?= go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0

.PHONY: up db-up db-down db-kill sqlc seed-wikidata fmt-check vet build ci

up:
	docker compose up -d --wait postgres
	go build -o kinosearch ./cmd/main.go
	./kinosearch

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-kill:
	docker compose down -v

sqlc:
	$(SQLC) generate

seed-wikidata:
	go run ./cmd/seed-wikidata -limit 1000

fmt-check:
	test -z "$$(gofmt -l .)"

vet:
	GOCACHE=/tmp/go-build go vet ./...

build:
	GOCACHE=/tmp/go-build go build ./...

ci: fmt-check vet build
