package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/isaacmirandacampos/go-expert/05-zipcode-weather-observability/pkg"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var tracer = otel.Tracer("weather-service")

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
	Erro       bool   `json:"erro,omitempty"`
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

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	ctx, span := tracer.Start(ctx, "weather-service-handler")
	defer span.End()

	cep := r.URL.Query().Get("cep")

	if matched, _ := regexp.MatchString(`^\d{8}$`, cep); !matched {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid zipcode"})
		return
	}

	ctxViaCEP, spanViaCEP := tracer.Start(ctx, "Fetch ViaCEP")
	viacepURL := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	req, _ := http.NewRequestWithContext(ctxViaCEP, "GET", viacepURL, nil)
	client := http.Client{}
	resp, err := client.Do(req)
	spanViaCEP.End()

	if err != nil {
		http.Error(w, "error fetching viaCEP", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "error reading viaCEP response", http.StatusInternalServerError)
		return
	}

	var cepInfo ViaCEPResponse
	if err := json.Unmarshal(body, &cepInfo); err != nil {
		http.Error(w, "error decoding viaCEP response", http.StatusInternalServerError)
		return
	}
	if cepInfo.Erro {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "can not find zipcode"})
		return
	}
	ctxViaWeather, spanViaWeather := tracer.Start(ctx, "Fetch WEATHER")
	weatherURL := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s",
		os.Getenv("WEATHER_API_KEY"), cepInfo.Localidade)
	reqWeather, _ := http.NewRequestWithContext(ctxViaWeather, "GET", weatherURL, nil)
	weatherResp, err := client.Do(reqWeather)
	spanViaWeather.End()

	if err != nil {
		http.Error(w, "error fetching WeatherAPI", http.StatusInternalServerError)
		return
	}
	defer weatherResp.Body.Close()

	var weather WeatherResponse
	if err := json.NewDecoder(weatherResp.Body).Decode(&weather); err != nil {
		http.Error(w, "error decoding WeatherAPI response", http.StatusInternalServerError)
		return
	}

	result := Response{
		TempC: weather.Current.TempC,
		TempF: weather.Current.TempC*1.8 + 32,
		TempK: weather.Current.TempC + 273,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	shutdown := pkg.InitTracer("weather-service")
	defer shutdown()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(otelhttp.NewMiddleware("weather-service"))
	r.Get("/weather", weatherHandler)

	log.Println("Weather Service running on :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
