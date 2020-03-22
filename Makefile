ROOT_DIR = $(shell pwd)
BIN_DIR = $(ROOT_DIR)/bin

build: build-client build-server

build-client:
	go build cmd/client/client.go
	mkdir -p $(BIN_DIR)
	mv client $(BIN_DIR)/alcatraz

build-server:
	go build cmd/server/server.go
	mkdir -p $(BIN_DIR)
	mv server $(BIN_DIR)/alcatrazd

clean:
	rm -rf $(BIN_DIR)

test:
	go test ./...  --cover -p 1