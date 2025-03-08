# Usar la imagen oficial de Golang
FROM golang:1.22-alpine AS builder

# Establecer el directorio de trabajo
WORKDIR /app

# Copiar go.mod y go.sum
COPY go.mod ./
COPY go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN go build -o DiscMuteBot bot/main.go

# Crear imagen final
FROM alpine:latest

WORKDIR /app

# Instalar CA certificates para HTTPS
RUN apk --no-cache add ca-certificates

# Copiar ejecutable compilado
COPY --from=builder /app/DiscMuteBot .

# Crear directorio de logs
RUN mkdir -p logs

# Puerto para posibles expansiones (no es necesario para un bot de Discord)
EXPOSE 8080

# Ejecutar el bot
CMD ["./DiscMuteBot"]