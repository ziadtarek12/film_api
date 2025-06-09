include .env

current_time = $(shell date --iso-8601=seconds)

## help: print this help message
.PHONY: help
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## api/build: build the api
.PHONY: api/build
api/build:
	@echo "Buliding api..."
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -X main.buildTime=${current_time}' -o=./bin/api ./cmd/api/


## run/api : Run the api
.PHONY: 
api/run:
	./bin/api -db-dsn=${FILMAPI_DB_DSN} -port=${APP_PORT} -limiter-burst=${LIMITER_BURST} -limiter-rps=${LIMITER_RPS} -limiter-enabled=${LIMITER_ENABLED} -cors-trusted-origin=*


## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path=./migrations -database=${FILMAPI_DB_DSN} up


.PHONY: db/psql
## db/psql: connect to the database using psql
db/psql:
	PGPASSWORD=${DB_PASSWORD} && @psql --host=${DB_HOST} --port=${DB_PORT} --username=${DB_USER} --dbname=${DB_NAME} 
	

db/import:
	PGPASSWORD=${DB_PASSWORD} && @psql --host=${DB_HOST} --port=${DB_PORT} --username=${DB_USER} --dbname=${DB_NAME} -f import_films.sql


host/copy:
	scp ${name} zeyadomaro@ssh-zeyadomaro.alwaysdata.net:/home/zeyadomaro	


## db/migrations/new name=$1: create a new database migration
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=sql -dir=./migrations ${name}

# Create the new confirm target
confirm:
	@echo -n 'Are you sure [y/N] ' && read ans && [ $${ans:-N} = y ]






## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

	
.PHONY: vendor 
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor
	
