PROTO_DIR = api/proto
GEN_DIR = api/gen

proto:
	protoc --go_out=$(GEN_DIR) --go-grpc_out=$(GEN_DIR) $(PROTO_DIR)/exchange.proto

swagger_wallet:
	cd wallet && swag init -g ./internal/transport/http/server.go

test_wallet:
	@echo "Running tests for wallet..."
	cd wallet && go test -v ./...

test_exchanger:
	@echo "Running tests for exchanger..."
	cd exchanger && go test -v ./...

test: test_wallet test_exchanger
