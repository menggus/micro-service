package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"library/v1/client"
	"library/v1/pb"
	"library/v1/sample"
	"log"
	"strings"
	"time"
)

func testCreateLaptop(conn *grpc.ClientConn) {
	// Create laptop service client
	client.NewLaptopClient(conn)
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	for i := 0; i < 10; i++ {
		laptopClient.CreateLaptop(sample.NewLaptop())
	}
	// Search laptop according to filter
	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 4},
	}
	laptopClient.SearchLaptop(filter)
}

func testUploadImage(laptopClient *client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.GetId(), "tmp/laptop.png")
}

func testRatingLaptop(laptopClient *client.LaptopClient) {
	n := 3
	// create laptopIDS and createLaptop store in memoryStore
	laptopIDs := make([]string, n)
	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIDs[i] = laptop.GetId()
		laptopClient.CreateLaptop(laptop)
	}
	// generate score list
	scores := make([]float64, n)
	for {
		fmt.Println("rate laptop y/n ?")
		var answer string
		fmt.Scan(&answer)

		if strings.ToLower(answer) != "y" {
			break
		}
		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}

		err := laptopClient.RatingLaptop(laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}

const (
	username        = "user1"
	password        = "123456"
	refreshDuration = 30 * time.Second
)

func authMethod() map[string]bool {
	const serverPath = "/techschool.proto.LaptopService/"
	return map[string]bool{
		serverPath + "CreateLaptop": true,
		serverPath + "UploadImage":  true,
		serverPath + "RatingLaptop": true,
	}
}

func main() {
	// Parse server address
	serverAddress := flag.String("address", "", "this is server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)

	// Start a grpc dial
	// WithInsecure 返回一个 DialOption，它禁用此 ClientConn 的传输安全。请注意，除非设置了 WithInsecure，否则需要传输安全性
	cc1, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	authClient := client.NewAuthClient(cc1, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, authMethod(), refreshDuration)
	if err != nil {
		log.Fatalf("cannot create a interceptor: %s", err)
	}

	cc2, err := grpc.Dial(
		*serverAddress,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	laptopClient := client.NewLaptopClient(cc2)
	testRatingLaptop(laptopClient)
}
