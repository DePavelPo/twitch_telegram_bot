include .env
export $(shell sed 's/=.*//' .env)

.PHONY: run
run:
	go run cmd/main.go