# ---------- build stage ----------
FROM golang:1.25.0 AS builder

# Рабочая директория внутри контейнера
WORKDIR /app

# build args
ARG VERSION=unknown
ENV GO111MODULE=on CGO_ENABLED=0

# Оптимизация кеша зависимостей
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Копируем остальной код
COPY . .

# Сборка бинаря
RUN --mount=type=cache,target=/go/pkg/mod \
     GOOS=linux go build -mod=readonly -trimpath -buildvcs=false\
    -ldflags="-s -w -X github.com/EgorLis/MicroserviceExampleGo/provider/internal/config.Version=${VERSION}" \
    -o /app/bin/provider ./cmd/provider
    
# ---------- run stage ----------
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# бинарь
COPY --from=builder /app/bin/provider /app/provider

# по умолчанию путь к конфигу (можно переопределить в compose)
ENV CONFIG_PATH=/app/config/config.yaml

# По умолчанию порт (для документации, но expose не обязателен)
EXPOSE 7081
USER nonroot:nonroot
ENTRYPOINT ["/app/provider"]