gen:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

clean:
	rm pb/*.go

server1:
	go run cmd/server/main.go -port 50051

server2:
	go run cmd/server/main.go -port 50052

server-tls1:
	go run cmd/server/main.go -port 50051 -tls

server-tls2:
	go run cmd/server/main.go -port 50052 -tls

server:
	go run cmd/server/main.go -port 8080

client-tls:
	go run cmd/client/main.go -address 0.0.0.0:8080 -tls

client:
	go run cmd/client/main.go -address 0.0.0.0:8080

test:
	go test -cover -race ./...

cert:
	cd certificate; ./gen.sh; cd ..


.PHONY: gen clean server client test cert