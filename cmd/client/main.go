package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"library/v1/pb"
	"library/v1/sample"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	laptopClient := testCreateLaptop(conn)

	//testSearchLaptop(laptopClient)

	//testUploadImage(laptopClient)

	testRatingLaptop(laptopClient)

}

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {
	// Create a laptopCreateRequest req
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

func testCreateLaptop(conn *grpc.ClientConn) pb.LaptopServiceClient {
	// Create laptop service client
	laptopClient := pb.NewLaptopServiceClient(conn)

	return laptopClient
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Println("search filter", filter)

	// Set request time out
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	// For Receive
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

func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient, sample.NewLaptop())
	}
	// Search laptop according to filter
	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 4},
	}
	searchLaptop(laptopClient, filter)
}

func UploadImage(laptopClient pb.LaptopServiceClient, laptopID string, imagePath string) {
	// 1.prepare image data
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()

	// 2.timeout handle
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 3.generate upload image stream
	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image", err)
	}

	// 4. construct a pb.UploadImageRequest
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	// 5. Send req by stream
	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send  image info", err)
	}

	// 6.stream method send image data
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to the buffer: ", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n], // attention buffer used
			},
		}

		//log.Println(len(buffer[:n]))

		err = stream.Send(req)
		if err != nil {
			err2 := stream.RecvMsg(nil)
			log.Fatal("cannot send image file data: ", err, err2)
		}
	}

	// 7. close and receive response, print response result
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}
	log.Printf("Upload image %s succeed, size is %d", res.GetId(), res.GetSize())
}

func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
	UploadImage(laptopClient, laptop.GetId(), "tmp/laptop.png")
}

func ratingLaptop(laptopClient pb.LaptopServiceClient, laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cnanot rate laptop %v", err)
	}

	// start goroutine receive stream response
	waitResponse := make(chan error)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Println("no more data")
				waitResponse <- nil
				return
			}
			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream response %v", err)
				return
			}

			log.Printf("receive rating-laptop with laptopID: %s, average score %.2f\n", res.GetLaptopId(), res.GetAverageRate())
		}
	}()

	// generate pb.RateLaptopRequest and send request
	for i, laptopID := range laptopIDs {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopID,
			Score:    scores[i],
		}

		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v", err)
		}
		log.Printf("Send request: %v", req)
	}
	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close stream %v", err)
	}

	err = <-waitResponse

	return err
}

func testRatingLaptop(laptopClient pb.LaptopServiceClient) {
	n := 3
	// create laptopIDS and createLaptop store in memoryStore
	laptopIDs := make([]string, n)
	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIDs[i] = laptop.GetId()
		createLaptop(laptopClient, laptop)
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

		err := ratingLaptop(laptopClient, laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}
