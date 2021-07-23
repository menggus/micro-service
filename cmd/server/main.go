package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"library/v1/pb"
	"library/v1/service"
	"log"
	net2 "net"
)

func main() {
	// Parse port param by flag package
	port := flag.Int("port", 0, "the is server port")
	flag.Parse()
	log.Printf("start server on port %d\n", *port)

	// Create a laptop service and  grpcServer

	// after, register laptop service in grpc server
	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	laptopServer := service.NewLaptopServer(laptopStore, imageStore)
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	// Create a net listener
	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net2.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start a listener: ", err)
	}

	// Start grpc server
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start a grpc server:", err)
	}

}
