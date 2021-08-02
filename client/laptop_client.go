package client

import (
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"library/v1/pb"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LaptopClient is a client to call laptop service rpc
type LaptopClient struct {
	service pb.LaptopServiceClient
}

func NewLaptopClient(cc *grpc.ClientConn) *LaptopClient {
	service := pb.NewLaptopServiceClient(cc)
	return &LaptopClient{service: service}
}

func (laptopClient *LaptopClient) CreateLaptop(laptop *pb.Laptop) {
	// Create a laptopCreateRequest req
	req := &pb.CreateLaptopRequest{Laptop: laptop}

	// Call CreateLaptop service
	// Case 1: Set timeout 5s, so request will cancel
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	// Case 2: Ctrl + c

	res, err := laptopClient.service.CreateLaptop(ctx, req)
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

func (laptopClient *LaptopClient) SearchLaptop(filter *pb.Filter) {
	log.Println("search filter", filter)

	// Set request time out
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create Search Laptop Request
	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}

	// Call Grpc-stream Method
	stream, err := laptopClient.service.SearchLaptop(ctx, req)
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

func (laptopClient *LaptopClient) UploadImage(laptopID string, imagePath string) {
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
	stream, err := laptopClient.service.UploadImage(ctx)
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

func (laptopClient *LaptopClient) RatingLaptop(laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.service.RateLaptop(ctx)
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
