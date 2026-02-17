APP := redisolar-go
PORT := 8081

.PHONY: build run dev test load clean frontend deps

all: deps test

deps:
	go mod tidy
	cd frontend && npm install

build:
	go build -o bin/server ./cmd/server
	go build -o bin/loader ./cmd/loader

test:
	go test ./...

frontend:
	cd frontend && npm run build
	rm -rf static
	cp -r frontend/dist/static static
	cp frontend/dist/index.html static/

load:
	go run ./cmd/loader

dev: frontend
	SERVER_PORT=$(PORT) go run ./cmd/server

run: build frontend
	SERVER_PORT=$(PORT) ./bin/server

clean:
	rm -rf bin/
	rm -rf frontend/dist
	rm -rf frontend/node_modules
