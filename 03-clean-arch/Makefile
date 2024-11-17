include ./cmd/ordersystem/.env

start:
	go run cmd/ordersystem/main.go cmd/ordersystem/wire_gen.go

build:
	go build -o bin/ordersystem cmd/ordersystem/main.go

test:
	go test ./...

generate_graph:
	go run github.com/99designs/gqlgen generate

create_migration:
	migrate create -ext=sql -dir=migrations -seq $(n)
