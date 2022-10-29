SERVICE_NAME = user-app
VOLUMES = pdb
ENV = ./server/.env

.PHONY: build
build:
	docker-compose -f ./docker-compose.yaml --env-file $(ENV) build

.PHONY: run
run:
	docker-compose -f ./docker-compose.yaml --env-file $(ENV) up

.PHONY: stop
stop:
	docker-compose -f ./docker-compose.yaml --env-file ${ENV} down

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: generate-proto
generate-proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
         --go-grpc_out=. --go-grpc_opt=paths=source_relative userpb/user.proto
	@echo "generate done"

.PHONY: migrate
migrate:
	goose -dir server/migrations  \
      postgres "user=postgres password=postgres host=localhost port=5432 database=postgres sslmode=disable" \
      status

#fclean: stop
#	docker system prune -a
#	docker volume rm -f $(VOLUMES)

# kcat -b localhost:9095 -G users users

.PHONY: rm_db_data
rm_db_data:
	sudo rm -rf ./.database/postgres/data/*



.PHONY: build, run, stop, fclean, generate, rm_db_data