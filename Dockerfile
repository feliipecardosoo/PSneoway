# Use a imagem oficial do Golang como base
FROM golang:1.22.4-alpine

# Defina o diretório de trabalho dentro do contêiner
WORKDIR /app

# Copie o código fonte para dentro do contêiner
COPY . .

# Construa o aplicativo Go
RUN go build -o modulo .

# Comando padrão a ser executado ao iniciar o contêiner
CMD ["./modulo"]
