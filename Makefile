gen:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

clean:
	rm pb/*.go

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go -address localhost:8080

test:
	go test -cover -race ./...