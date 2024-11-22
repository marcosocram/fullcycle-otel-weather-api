package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.26.0"
	"log"
	"net/http"
	"net/url"
	"os"
)

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func setupTracer(serviceName string) func() {
	ctx := context.Background()

	//exporter, _ := zipkin.New("http://zipkin:9411/api/v2/spans")
	// Configurar o exportador OTLP
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}
	bsp := trace.NewBatchSpanProcessor(exporter)
	provider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(bsp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Função para encerrar o rastreamento
	return func() {
		if err := provider.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	}
}

// Estrutura para deserializar a resposta da API viaCEP
type ViaCEPResponse struct {
	Localidade string `json:"localidade"` // Cidade
}

// Consulta a API viaCEP e retorna a cidade para um CEP dado
func fetchCity(cep string) (string, error) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("não foi possível obter a cidade")
	}

	var viaCEP ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&viaCEP); err != nil {
		return "", err
	}

	if viaCEP.Localidade == "" {
		return "", errors.New("cidade não encontrada para o CEP informado")
	}

	return viaCEP.Localidade, nil
}

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

// Consulta a WeatherAPI para obter a temperatura em Celsius de uma cidade específica
func fetchTemperature(city string) (float64, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, url.QueryEscape(city))
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("não foi possível obter a temperatura")
	}

	var weatherResponse WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		return 0, err
	}

	return weatherResponse.Current.TempC, nil
}

// Converte Celsius para Fahrenheit e Kelvin
func calculateTemperatures(tempC float64) (float64, float64) {
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15
	return tempF, tempK
}

// Formata e retorna os dados climáticos, incluindo as conversões de temperatura
func getWeatherData(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	ctx, span := otel.Tracer("service-b").Start(ctx, "getWeatherData")

	cep := r.URL.Query().Get("cep")

	if len(cep) != 8 {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	ctx, spanCity := otel.Tracer("service-b").Start(ctx, "fetchCity")
	city, err := fetchCity(cep)
	if err != nil {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}
	spanCity.End()

	ctx, spanTemperature := otel.Tracer("service-b").Start(ctx, "fetchTemperature")
	tempC, err := fetchTemperature(city)
	if err != nil {
		http.Error(w, "temperature service unavailable", http.StatusInternalServerError)
		return
	}
	spanTemperature.End()

	tempF, tempK := calculateTemperatures(tempC)
	weatherResponse := WeatherResponse{City: city, TempC: tempC, TempF: tempF, TempK: tempK}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weatherResponse)

	span.End()
}

func main() {
	// Inicializa o rastreamento
	shutdown := setupTracer("service-b")
	defer shutdown()

	http.HandleFunc("/weather", getWeatherData)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
