HTTP_LISTEN_ADDR:=127.0.0.1
HTTP_LISTEN_PORT:=8080

build:
	mkdir build/
	go build -o build/weather-service main.go

test:
	go vet ./...
	go test -v ./...

run:
	go run main.go