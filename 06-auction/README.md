# Auction

Esse repositório é uma atividade da pós-graduação e grande parte do código fonte foi retirado do repositório: [labs-auction-goexpert](https://github.com/devfullcycle/labs-auction-goexpert.git)

## Descrição

Conforme a proposta que foi levantada na atividade, segue as edições realizadas nos arquivos:
- 06-auction/internal/infra/database/auction/create_auction.go
- 06-auction/internal/infra/database/auction/create_auction_test.go

## Execução

Para executar o projeto, basta rodar o comando abaixo:

```bash
docker-compose up
```

## Endpoints

Com os dois endpoints abaixo você já consegue visualizar o funcionamento do projeto.

POST /auctions
```json
{
	"product_name": "Um produto bonzao",
  "category": "tecnologico",
  "description": "selo de aprovado do Inmetro"
}
```
GET /auctions - Lista todos os leilões

POST /bid
```json
{
	"amount": 2000,
  "user_id": "", // Pode ser qualquer valor
  "auction_id": "" // liste os auctions e pegue o id
}
```

GET /auction/winner/auction_id - Retorna o vencedor do leilão

## Testes

Para rodar os testes, basta rodar o comando abaixo:

```bash
go test ./...
```
