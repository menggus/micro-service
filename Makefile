gen:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

clean:
	rm pb/*.go

run:
	go run main.go

test:
	go test -cover -race ./...