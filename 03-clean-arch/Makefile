include ./cmd/ordersystem/.env

build:
	go build -o bin/ordersystem cmd/ordersystem/main.go cmd/ordersystem/wire_gen.go

test:
	go test ./...

generate_graph:
	go run github.com/99designs/gqlgen generate

create_migration:
	migrate create -ext=sql -dir=migrations -seq $(n)
