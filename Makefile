TEST_DB_NAME?=postgres
TEST_DB_HOST?=localhost
TEST_DB_USER?=postgres
TEST_DB_PWD?=pswd
TEST_DB_PORT?=5432
TEST_DB_DISABLE_SSL=true
TEST_POSTGRES_CONTAINER_NAME=service-catalog-test-db

build:
	CGO_ENABLED=0 go build -a -o ./bin/server ./cmd/app

run:
	CGO_ENABLED=0 go run cmd/app/main.go

fmt:
	gofmt -l -s -w ./
	goimports -l -w ./

tidy:
	go mod tidy

setup-test-db:
	docker run --name $(TEST_POSTGRES_CONTAINER_NAME) -e POSTGRES_PASSWORD=$(TEST_DB_PWD) -p $(TEST_DB_PORT):5432 -d postgres
	rm .env.test
	echo "POSTGRES_USER=$(TEST_DB_USER)" >> .env.test
	echo "POSTGRES_PASSWORD=$(TEST_DB_PWD)" >> .env.test
	echo "POSTGRES_HOST=$(TEST_DB_HOST)" >> .env.test
	echo "POSTGRES_PORT=$(TEST_DB_PORT)" >> .env.test
	echo "POSTGRES_DB_NAME=$(TEST_DB_NAME)" >> .env.test
	echo "POSTGRES_DISABLE_SSL=$(TEST_DB_DISABLE_SSL)" >> .env.test
	echo "JWT_SIGNING_KEY=test-key" >> .env.test
	echo "TOKEN_HOUR_LIFESPAN=1" >> .env.test

destroy-test-db:
	@if [[ $$(docker ps -a | grep $(TEST_POSTGRES_CONTAINER_NAME)) ]]; then \
		docker stop $(TEST_POSTGRES_CONTAINER_NAME); \
		docker rm $(TEST_POSTGRES_CONTAINER_NAME); \
	fi

test: destroy-test-db setup-test-db
	go test -v ./...
