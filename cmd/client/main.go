package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"library/v1/pb"
	"library/v1/sample"
	"log"
	"time"
)

func main() {
	// Parse server address
	serverAddress := flag.String("address", "", "this is server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)

	// Start a grpc dial
	// WithInsecure 返回一个 DialOption，它禁用此 ClientConn 的传输安全。请注意，除非设置了 WithInsecure，否则需要传输安全性
	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	// Create laptop service client
	laptopClient := pb.NewLaptopServiceClient(conn)
	for i:=0;i<10;i++ {
		createLaptop(laptopClient)
	}

	// Search laptop
	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz: 2.5,
		MinRam: &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 4},
	}
	searchLaptop(laptopClient, filter)
}

func createLaptop(laptopClient pb.LaptopServiceClient)  {
	// Create a laptopCreateRequest req
	laptop := sample.NewLaptop()
	laptop.Id = ""
	req := &pb.CreateLaptopRequest{Laptop: laptop}

	// Call CreateLaptop service
	// Case 1: Set timeout 5s, so request will cancel
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	// Case 2: Ctrl + c

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Println("laptop already exists")
		} else {
			log.Fatal("cannot create laptop ", err)
		}
		return
	}
	log.Printf("Created laptop with id: %s\n", res.Id)
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Println("search filter", filter)

	// Set request time out
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	// Create Search Laptop Request
	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}

	// Call Grpc-stream Method
	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop: ", err)
	}

	// For Recv
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}

		laptop := res.GetLaptop()
		log.Println("- found: ", laptop.GetId())
		log.Println("+ brand: ", laptop.GetBrand())
		log.Println("+ name: ", laptop.GetName())
		log.Println("+ cpu cores: ", laptop.GetCpu().GetNumCores())
		log.Println("+ cpu min ghz: ", laptop.GetCpu().GetMinGhz())
		log.Println("+ ram: ", laptop.GetRam())
		log.Println("+ price: ", laptop.GetPriceUsd(), "USD")
	}
}