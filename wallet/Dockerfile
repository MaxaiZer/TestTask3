FROM scratch AS base

FROM golang:1.23 AS build
WORKDIR /app

COPY api/go.mod api/go.sum ./api/
COPY wallet/go.mod wallet/go.sum ./wallet/

WORKDIR /app/api
RUN go mod download

WORKDIR /app/wallet
RUN go mod download

WORKDIR /app
COPY api/ ./api/
COPY wallet/cmd/ ./wallet/cmd
COPY wallet/docs/ ./wallet/docs/
COPY wallet/internal/ ./wallet/internal/

WORKDIR /app/wallet/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/build/main main.go

FROM base AS final
WORKDIR /app
COPY wallet/configs ./configs/
COPY wallet/migrations ./migrations/
COPY --from=build /app/build/main .
CMD ["./main"]