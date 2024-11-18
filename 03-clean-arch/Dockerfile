FROM golang:1.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o server cmd/ordersystem/main.go cmd/ordersystem/wire_gen.go

FROM scratch

WORKDIR /app
COPY --from=builder /app/cmd/ordersystem/.env .

COPY --from=builder /app/server .

EXPOSE 8080
EXPOSE 8000
EXPOSE 50051


CMD ["./server"]
