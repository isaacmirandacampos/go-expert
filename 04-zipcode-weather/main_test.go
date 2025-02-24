package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInvalidCEP(t *testing.T) {
	req, err := http.NewRequest("GET", "/weather?cep=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(weatherHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("código de status incorreto: obtido %v, esperado %v", status, http.StatusUnprocessableEntity)
	}

	var res map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if res["message"] != "invalid zipcode" {
		t.Errorf("mensagem incorreta: obtida %s, esperada 'invalid zipcode'", res["message"])
	}
}

// Para testes adicionais (como simular as chamadas às APIs externas), recomenda-se utilizar mocks ou servidores de teste.
