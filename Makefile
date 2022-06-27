include .env
export

.PHONY: env-test
env-test:
	.env

.PHONY: run
run:
	.env go run cmd/main.go