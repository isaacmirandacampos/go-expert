# Zipcode Weather Observability

É aconselhável que tente rodar o projeto através do docker-compose.
```sh
WEATHER_API_KEY="SUA API KEY DO WEATHER" docker-compose up
```

## Como testar o projeto

Rode o comando
```sh
curl -X POST http://localhost:8080/weather \
     -H "Content-Type: application/json" \
     -d '{"cep": "29164063"}'
```

## visualizar os traces

visite a url http://localhost:9411
