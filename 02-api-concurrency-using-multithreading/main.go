package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	viaCepURL    = "https://viacep.com.br/ws/%s/json"
	brasilApiURL = "https://brasilapi.com.br/api/cep/v1/%s"
)

type BrasilApi struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type CepResult struct {
	ApiUsed string `json:"api_used"`
	Cep     string `json:"cep"`
	Estado  string `json:"estado"`
	Cidade  string `json:"cidade"`
	Bairro  string `json:"bairro"`
	Rua     string `json:"rua"`
}

func main() {
	if len(os.Args) < 2 {
		panic("Please provide a CEP to fetch")
	}

	cep := os.Args[1]
	fetchCEPConcurrentlyWithChannel(cep)
}

func fetchCEPConcurrentlyWithChannel(cep string) {
	cepResult := make(chan *CepResult, 2) // Using 2 since we're fetching from two APIs
	go fetch(cep, viaCepURL, "ViaCep", cepResult)
	go fetch(cep, brasilApiURL, "BrasilApi", cepResult)
	select {
	case res := <-cepResult:
		fmt.Printf("Fetched from %s: %+v\n", res.ApiUsed, res)
	case <-time.After(1 * time.Second):
		fmt.Println("Exceeded time limit")
		return
	}
}

func fetch(cep string, url string, apiName string, cepResult chan<- *CepResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(url, cep), nil)
	if err != nil {
		fmt.Printf("Error creating request to %s: %v\n", apiName, err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error fetching from %s: %v\n", apiName, err)
	}
	defer resp.Body.Close()
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}
	var result CepResult
	switch apiName {
	case "ViaCep":
		var viaCep ViaCep
		if err := json.Unmarshal(res, &viaCep); err != nil {
			fmt.Printf("Error unmarshaling ViaCep response: %v\n", err)
			return
		}
		result = CepResult{
			ApiUsed: apiName,
			Cep:     viaCep.Cep,
			Estado:  viaCep.Uf,
			Cidade:  viaCep.Localidade,
			Bairro:  viaCep.Bairro,
			Rua:     viaCep.Logradouro,
		}
	case "BrasilApi":
		var brasilApi BrasilApi
		if err := json.Unmarshal(res, &brasilApi); err != nil {
			fmt.Printf("Error unmarshaling BasilApi response: %v\n", err)
			return
		}
		result = CepResult{
			ApiUsed: apiName,
			Cep:     brasilApi.Cep,
			Estado:  brasilApi.State,
			Cidade:  brasilApi.City,
			Bairro:  brasilApi.Neighborhood,
			Rua:     brasilApi.Street,
		}
	}
	cepResult <- &result
}
