# ---------- build stage ----------
FROM golang:1.25.0 AS builder

# Рабочая директория внутри контейнера
WORKDIR /app

# build args
ARG VERSION=unknown

# Оптимизация кеша зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код
COPY . .

# Сборка бинаря
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config.Version=${VERSION}" \
    -o checkout ./cmd/checkout
    
# ---------- run stage ----------
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Копируем только бинарь из builder
COPY --from=builder /app/checkout .

# По умолчанию порт (для документации, но expose не обязателен)
EXPOSE 8081

# Запуск
CMD ["./checkout"]