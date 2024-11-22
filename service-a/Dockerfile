# Dockerfile

# Usar uma imagem base do Golang
FROM golang:1.23

# Definir o diretório de trabalho
WORKDIR /app

# Copiar o código do serviço para o contêiner
COPY . .

# Instalar dependências
RUN go mod download

# Compilar o binário
RUN go build -o main .

# Comando para iniciar o serviço
CMD ["./main"]
