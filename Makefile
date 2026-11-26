.PHONY: init download mocks sql install uninstall

init: download mocks

download:
	go mod download

mocks:
	go install go.uber.org/mock/mockgen@latest
	go generate ./...

sql:
	sqlc generate -f scripts/db/sqlc.yaml

install:
	./scripts/setup.sh

uninstall:
	./scripts/purge.sh
