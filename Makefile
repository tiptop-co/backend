ifneq (,$(wildcard ./.env))
    include .env
    export
endif


testgen:
	./bin/testgen -path "$(CURDIR)/$(PATH)"

run:
	env $(cat .env | xargs) go run ./cmd/service/main.go

migrate-up:
	goose -dir db/migrations postgres "host=${POSTGRES_HOST} port=${POSTGRES_PORT} user=${POSTGRES_USER} \
		password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DATABASE} sslmode=${POSTGRES_SSLMODE}" up

migrate-down:
	goose -dir db/migrations postgres "host=${POSTGRES_HOST} port=${POSTGRES_PORT} user=${POSTGRES_USER} \
		password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DATABASE} sslmode=${POSTGRES_SSLMODE}" down

fill-db:
	go run ./cmd/seed
