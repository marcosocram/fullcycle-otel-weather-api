package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.26.0"
	"log"
	"net/http"
)

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

type Weather struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type CEPRequest struct {
	CEP string `json:"cep"`
}

func getWeather(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	ctx, span := otel.Tracer("service-a").Start(ctx, "getWeather")

	var cepReq CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&cepReq); err != nil || len(cepReq.CEP) != 8 {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	url := fmt.Sprintf("http://service_b:8082/weather?cep=%s", cepReq.CEP)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "error fetching weather", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var data Weather
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		http.Error(w, "error decoding weather data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)

	span.End()
}

func main() {
	// Inicializa o rastreamento
	shutdown := setupTracer("service-a")
	defer shutdown()

	r := mux.NewRouter()
	r.HandleFunc("/get-weather", getWeather).Methods("POST")

	log.Fatal(http.ListenAndServe(":8081", r))
}
