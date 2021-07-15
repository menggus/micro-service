gen:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

clean:
	rm pb/*.go

server:
	go run cmd/server/main.go

client:
	go run cmd/client/main.go

test:
	go test -cover -race ./...