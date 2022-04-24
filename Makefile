SERVICE_NAME = user-app

#DB_URL = 'postgres://postgres:postgres@0.0.0.0:5432/postgres?sslmode=disable'

build:
	docker-compose build $(SERVICE_NAME)

run: 
	docker-compose up $(SERVICE_NAME)	

#migrate:
#	migrate -path ./server/schema -database $(DB_URL) up


.PHONY: build, run