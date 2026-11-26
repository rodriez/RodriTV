.PHONY: init downloads mocks sql

init: download mocks

download:
	go mod download

mocks:
	go install go.uber.org/mock/mockgen@latest
	go generate ./...

sql:
	sqlc generate -f scripts/db/sqlc.yaml