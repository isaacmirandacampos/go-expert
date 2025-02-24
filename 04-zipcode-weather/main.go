package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
	Erro        bool   `json:"erro,omitempty"`
}

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type Response struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func main() {
	http.HandleFunc("/weather", weatherHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor iniciando na porta %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	cep := r.URL.Query().Get("cep")
	// Validação: CEP deve ter exatamente 8 dígitos numéricos
	matched, _ := regexp.MatchString(`^\d{8}$`, cep)
	if !matched {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid zipcode"})
		return
	}

	// Consulta à API viaCEP
	viacepURL := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	resp, err := http.Get(viacepURL)
	if err != nil {
		http.Error(w, "erro ao consultar viaCEP", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "erro ao ler resposta viaCEP", http.StatusInternalServerError)
		return
	}

	var cepInfo ViaCEPResponse
	if err := json.Unmarshal(body, &cepInfo); err != nil {
		http.Error(w, "erro ao decodificar resposta viaCEP", http.StatusInternalServerError)
		return
	}
	if cepInfo.Erro {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "can not find zipcode"})
		return
	}

	// Consulta à WeatherAPI usando a cidade obtida
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		http.Error(w, "chave da WeatherAPI não configurada", http.StatusInternalServerError)
		return
	}
	weatherURL := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, cepInfo.Localidade)
	weatherResp, err := http.Get(weatherURL)
	if err != nil {
		http.Error(w, "erro ao consultar WeatherAPI", http.StatusInternalServerError)
		return
	}
	defer weatherResp.Body.Close()

	weatherBody, err := ioutil.ReadAll(weatherResp.Body)
	if err != nil {
		http.Error(w, "erro ao ler resposta WeatherAPI", http.StatusInternalServerError)
		return
	}

	var weather WeatherResponse
	if err := json.Unmarshal(weatherBody, &weather); err != nil {
		http.Error(w, "erro ao decodificar resposta WeatherAPI", http.StatusInternalServerError)
		return
	}

	// Conversões das temperaturas:
	tempC := weather.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273

	result := Response{
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
