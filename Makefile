SERVICE_NAME = user-app
VOLUMES = pdb
ENV = ./server/.env

export GO111MODULE=on

build:
	docker-compose -f ./docker-compose.yaml --env-file $(ENV) build

run:
	docker-compose -f ./docker-compose.yaml --env-file $(ENV) up

stop:
	docker-compose -f ./docker-compose.yaml --env-file ${ENV} down

generate:
	@protoc --go_out=. --go_opt=paths=source_relative \
         --go-grpc_out=. --go-grpc_opt=paths=source_relative userpb/user.proto
	@echo "generate done"

migrate:
	goose -dir server/migrations  \
      postgres "user=postgres password=postgres host=localhost port=5432 database=postgres sslmode=disable" \
      status

#fclean: stop
#	docker system prune -a
#	docker volume rm -f $(VOLUMES)

# kcat -b localhost:9095 -G users users

rm_db_data:
	sudo rm -rf ./.database/postgres/data/*

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: build, run, stop, fclean, generate, rm_db_data