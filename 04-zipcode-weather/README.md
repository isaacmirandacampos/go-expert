# Zipcode Weather

Este é um projeto que está hospedado no cloud run, você consegue testar usando o link:
```sh
curl -X GET 'https://my-weather-service-1033737360968.us-central1.run.app/weather?cep=$CEP'
```

## Como rodar o projeto

Instale as dependências do projeto:
```sh
go install
```

Rode o projeto:
```sh
WEATHER_API_KEY="" go run main.go
```
*Você consegue gerar seu WEATHER_API_KEY no site https://www.weatherapi.com/.*

## Como rodar os testes

```sh
go test .
```
