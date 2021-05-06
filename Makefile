.PHONY: test build run db-api setup migrate-api

# APP_NAME is used as a naming convention for resources to the local environment
ifndef APP_NAME
APP_NAME := notification
endif
APP_PWD := "$(shell pwd)"

# Commands for running docker compose
COMPOSE := docker-compose -f docker-compose.yaml
RUN_COMPOSE = $(COMPOSE) run --rm --service-ports -w /${APP_NAME}-api -v $(APP_PWD):/${APP_NAME}-api go-api
GO_COMPOSE = $(COMPOSE) run --rm -w /${APP_NAME}-api -v $(APP_PWD):/${APP_NAME}-api go-api

setup: db-api sleep migrate-api

# test executes project tests in a golang container
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
test: setup
	@if $(GO_COMPOSE) env $(shell cat .env | egrep -v '^#|^DATABASE_URL' | xargs) make go-test; \
	then printf "\n\n$(OK_COLOR)[Test ok -- `date`]$(NO_COLOR)\n"; \
	else printf "\n\n$(ERROR_COLOR)[Test FAILED -- `date`]$(NO_COLOR)\n"; exit 1; fi

# go-test executes test for all packages
go-test:
	go test -coverprofile=c.out -failfast -timeout 5m ./...

# run starts the web server in a golang container
run: setup
	@$(RUN_COMPOSE) env $(shell cat .env | egrep -v '^#|^DATABASE_URL' | xargs) \
		go run cmd/serverd/main.go

coverage: test
	go tool cover -html=c.out

build: setup
	$(GO_COMPOSE) make go-build

go-build:
	go build -v ./cmd/serverd

# teardown stops and removes all containers and resources associated to docker-compose.yml
teardown:
	$(COMPOSE) down -v

db-api:
	$(COMPOSE) up -d db-api

# migrate-api runs the migrate service for API defined in the compose file
migrate-api:
	$(COMPOSE) run --rm -v $(APP_PWD)/data/migrations:/migrations migrate-api \
	sh -c './migrate -path /migrations -database $$DATABASE_URL up'

# sleep is to delay the test from running to ensure all services (i.e. mq, db, redis) are up
sleep:
	sleep 2

generate-models:
	$(shell rm -rf internal/models/bmodels/*)
	sqlboiler --no-tests psql
	goreturns -w internal/models/bmodels/*.go
