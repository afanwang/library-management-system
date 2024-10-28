IMAGE_NAME ?= library-management
IMAGE_VERSION ?= latest
PLATFORM ?= linux/amd64
UNAME = $(shell uname -m)
TRGT = build_$(UNAME)

## Need to have Docker up and running to execute the below command
postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret  -d postgres:16.4-alpine3.20

docker_build:
	docker build --platform $(PLATFORM) -t $(IMAGE_NAME):$(IMAGE_VERSION) .

server:
	go run cmd/app/main.go --config configs/app.yaml

createdb:
	docker exec -it postgres createdb --username=root --owner=root library-management

dropdb:
	docker exec -it postgres dropdb --username=root library-management

migrate-up:
	migrate -path database/migration -database "postgresql://root:secret@localhost:5432/library-management?sslmode=disable" -verbose up

migrate-down:
	migrate -path internal/database/migration -database "postgresql://root:secret@localhost:5432/library-management?sslmode=disable" -verbose down

sqlc:
	sqlc generate
