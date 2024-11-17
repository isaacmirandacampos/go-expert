# Contexto do desáfio:
O código cria pedidos e já implementa graphql, grpc e web server. O desafio, é criar uma listagem das ordens implementando as boas práticas da clean arch com cada implementação, grpc, graphql e webserver.

Esta listagem precisa ser feita com:
- Endpoint REST (GET /order)
- Service ListOrders com GRPC
- Query ListOrders GraphQL

Não esqueça de criar as migrações necessárias e o arquivo api.http com a request para criar e listar as orders.

Para a criação do banco de dados, utilize o Docker (Dockerfile / docker-compose.yaml), com isso ao rodar o comando docker compose up tudo deverá subir, preparando o banco de dados.
Inclua um README.md com os passos a serem executados no desafio e a porta em que a aplicação deverá responder em cada serviço.


## Como rodar:
Levante a infraestrutura executando:
```bash
docker compose up -d
```

entre na pasta:
```bash
cd cmd/ordersystem
```

execute o projeto:
```bash
go run main.go wire_gen.go
```

## Testando o web server
Use o plugin do vscode api rest ou a funciondade do goland de executar http requests

Endpoints disponíveis:
-   POST create_order http://localhost:8000/order
-   GET list_orders http://localhost:8000/orders

## Testando o graphql
Abra o navegador e vá para http://localhost:8080 

Para criar uma ordem:

```graphql
mutation CreateOrder {
  createOrder(input:{ id: "bb", Price: 10.20, Tax: 10}) {
    id
    Price
    Tax
    FinalPrice
  }
}
```

Para listar as ordens:

```graphql
query ListOrder {
  listOrders {
    id
    Price
    Tax
    FinalPrice
  }
}
```
