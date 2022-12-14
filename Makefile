include .env
export $(shell sed 's/=.*//' .env)

.PHONY: run
run:
	go run cmd/main.go

.PHONY: migrate
migrate:
	go run cmd/migrate/main.go

.PHONY: migrate-down
migrate-down:
	go run cmd/migrate/main.go --down