SHELL := /bin/bash

.PHONY: up down seed status reset run

up:
	./scripts/migrate.sh up

down:
	./scripts/migrate.sh down 1

seed:
	./scripts/migrate.sh seed

status:
	./scripts/migrate.sh status

reset:
	./scripts/migrate.sh down 999
	./scripts/migrate.sh up
	./scripts/migrate.sh seed

run:
	set -a; [ -f .env ] && source .env; set +a; \
	lsof -ti :8080 | xargs -r kill -9 || true; \
	go build -ldflags="-linkmode=external" -o ./bin/server ./cmd/server; \
	codesign -s - ./bin/server || true; \
	./bin/server
