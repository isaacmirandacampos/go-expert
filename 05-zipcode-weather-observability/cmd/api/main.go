package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/isaacmirandacampos/go-expert/05-zipcode-weather-observability/pkg"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("api-service")

type WeatherResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type RequestBody struct {
	CEP string `json:"cep"`
}

func isValidCEP(cep string) bool {
	match, _ := regexp.MatchString(`^[0-9]{8}$`, cep)
	return match
}

func apiWeatherHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "api-service-handler")
	defer span.End()

	var req RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !isValidCEP(req.CEP) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		http.Error(w, "API_URL environment variable is not set", http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("%s/weather?cep=%s", apiURL, req.CEP)

	reqCtx, reqSpan := tracer.Start(ctx, "fetch_weather_service")

	start := time.Now()
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	request, _ := http.NewRequestWithContext(reqCtx, "GET", url, nil)

	resp, err := client.Do(request)
	reqSpan.End()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch weather data: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	log.Printf("Weather service response time: %v", time.Since(start))

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func main() {
	shutdown := pkg.InitTracer("api-service")
	defer shutdown()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(otelhttp.NewMiddleware("api-service"))
	r.Post("/weather", apiWeatherHandler)

	log.Println("API Service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
