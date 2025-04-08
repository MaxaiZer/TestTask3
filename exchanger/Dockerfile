FROM scratch AS base

FROM golang:1.23 AS build
WORKDIR /app

COPY api/go.mod api/go.sum ./api/
COPY exchanger/go.mod exchanger/go.sum ./exchanger/

WORKDIR /app/api
RUN go mod download

WORKDIR /app/exchanger
RUN go mod download

WORKDIR /app
COPY api/ ./api/
COPY exchanger/cmd/ ./exchanger/cmd
COPY exchanger/internal/ ./exchanger/internal/

WORKDIR /app/exchanger/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/build/main main.go

FROM base AS final
WORKDIR /app
COPY exchanger/configs ./configs/
COPY exchanger/migrations ./migrations/
COPY --from=build /app/build/main .
CMD ["./main", "--config", "configs/config.env"]