# fullcycle-otel-weather-api

Este projeto consiste em dois serviços desenvolvidos em Go (`service-a` e `service-b`) que trabalham em conjunto para fornecer informações sobre o clima atual de uma localidade com base em um CEP fornecido. Além disso, a aplicação implementa rastreamento distribuído usando **OpenTelemetry (OTEL)** e **Zipkin** para monitoramento.

---

## Arquitetura do Projeto

1. **Serviço A**: Responsável por receber uma requisição POST com um CEP no formato JSON, validar o CEP e encaminhá-lo ao Serviço B.
2. **Serviço B**: Responsável por:
   - Validar o CEP recebido.
   - Consultar a localização usando a API do ViaCEP.
   - Consultar as informações climáticas usando a API do WeatherAPI.
   - Retornar a cidade e as temperaturas em Celsius, Fahrenheit e Kelvin.
3. **OTEL Collector**: Coleta os spans gerados pelos serviços A e B e os envia para o Zipkin.
4. **Zipkin**: Interface de visualização para o rastreamento distribuído.

---

## Tecnologias Utilizadas

- **Go**: Linguagem de programação principal.
- **OpenTelemetry (OTEL)**: Para rastreamento distribuído.
- **Zipkin**: Para visualização do tracing.
- **Docker/Docker Compose**: Para containerização e gerenciamento dos serviços.
- **ViaCEP API**: Para consulta de localização.
- **WeatherAPI**: Para consulta de clima.

---

## Pré-requisitos

- Docker e Docker Compose instalados.
- API Key para o WeatherAPI. Cadastre-se [aqui](https://www.weatherapi.com/) para obter uma chave gratuita.

## Configuração e Execução com Docker Compose
1. Clone este repositório:

    ```bash
    git clone https://github.com/marcosocram/fullcycle-otel-weather-api.git
    cd fullcycle-otel-weather-api
    ```

2. Substitua a variável de ambiente **`WEATHER_API_KEY`** no `docker-compose.yml` com sua chave de API do WeatherAPI:

    ```yaml
    environment:
      WEATHER_API_KEY: "sua_chave_api_aqui"
     ```

3. Inicie o serviço com Docker Compose:

    ```bash
    docker-compose up --build
    ```
   Isso iniciará os seguintes serviços:

   * Serviço A: Disponível em http://localhost:8081
   * Serviço B: Disponível em http://localhost:8082
   * OTEL Collector: Conecta-se aos serviços para coletar spans.
   * Zipkin: Interface disponível em http://localhost:9411

## Testando o Serviço
Para verificar o funcionamento do serviço, você pode fazer uma requisição para o endpoint /weather com um CEP válido. Aqui estão alguns exemplos de como testar:

### Exemplo de Requisição de Sucesso
```bash
curl -X POST http://localhost:8081/get-weather -H "Content-Type: application/json" -d '{"cep":"88110606"}'
```

### Exemplo de Possíveis Respostas:
* Sucesso (200):
    ```json
    {
      "city": "São José",
      "temp_C": 15.6,
      "temp_F": 60.08,
      "temp_K": 288.75
    }
    ```

* CEP Inválido (422):
    ```plaintext
    invalid zipcode
    ```

* CEP Não Encontrado (404):
    ```plaintext
    can not find zipcode
    ```
  
### Verificar o Rastreamento no Zipkin
Acesse o Zipkin em http://localhost:9411. Você verá os spans das requisições do Serviço A e do Serviço B.

## Monitoramento e Debugging
* Ver logs do Serviço A:
    ```bash
    docker-compose logs -f service-a
    ```
* Ver logs do Serviço B:
    ```bash
    docker-compose logs -f service-b
    ```
* Ver logs do OTEL Collector:
    ```bash
    docker-compose logs -f otel-collector
    ```
* Ver logs do Zipkin:
    ```bash
    docker-compose logs -f zipkin
    ```

