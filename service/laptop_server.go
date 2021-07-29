package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"library/v1/pb"
	"log"
)

const MaxImageSize = 1 << 20

type LaptopServer struct {
	laptopStore                         LaptopStore
	imageStore                          ImageStore
	ratingStore                         RatingStore
	pb.UnimplementedLaptopServiceServer // 必须嵌入以具有向前兼容的实现
}

// NewLaptopServer create *LaptopServer
func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{
		laptopStore: laptopStore,
		imageStore:  imageStore,
		ratingStore: ratingStore,
	}
}

// CreateLaptop creae Laptop and save in the store
func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	// Get params
	laptop := req.GetLaptop()
	log.Printf("receive a create-laptop request with id: %s", laptop.Id)

	// laptop ID check or generate
	// if ID exists, it need checked
	// or ID not exists, it need generate
	if len(laptop.Id) > 0 {
		// check if it's a valid uuid
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "cannot save laptop to the store")
		}
	} else {
		//id, err := uuid.NewUUID()  // version 1 uuid
		id, err := uuid.NewRandom() // version 4 uuid
		if err != nil {
			return nil, fmt.Errorf("cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	// todo something
	// Set timeout test
	// time.Sleep(6 * time.Second)

	if ctx.Err() == context.DeadlineExceeded { // deadline cancel all exceeded
		log.Println("deadline exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
	}

	// Ctrl+C
	if ctx.Err() == context.Canceled {
		log.Println("request cancel")
		return nil, status.Error(codes.Canceled, "request cancel")
	}

	// Save the laptop to store
	// example, the laptop is saved in memory-store
	// Normally it is saved in DB
	err := server.laptopStore.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Error(code, "cannot save laptop to the store")
	}
	log.Printf("saved laptop wiht id: %s", laptop.Id)

	// Return response
	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}
	return res, nil
}

func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	// Get Filter
	filter := req.GetFilter()
	log.Printf("receive a search-laptop request with filter: %v", filter)

	// according filter to search laptop and callback func
	err := server.laptopStore.Search(
		stream.Context(),
		filter,
		func(laptop *pb.Laptop) error {
			// Construct pb.SearchLaptopResponse
			res := &pb.SearchLaptopResponse{Laptop: laptop}

			// Use stream method to send it
			err := stream.Send(res)
			if err != nil {
				return err
			}
			log.Printf("sent laptop with id: %s", laptop.Id)
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}

// UploadImage is client stream RPC to upload laptop image
func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Error(codes.Internal, "cannot receive image info"))
	}

	// Get upload information of laptop image
	laptopId := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("reveive a image upload request for laptop[%s] with image type %s", laptopId, imageType)

	// Checked laptop if exists
	laptop, err := server.laptopStore.Find(laptopId)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot find laptop %v", err))
	}
	if laptop == nil {
		return logError(status.Errorf(codes.NotFound, "laptop[%s] doesn't exists", laptopId))
	}

	// Get image data, stores into image store
	imageData := bytes.Buffer{}
	imageSize := 0
	for {
		log.Println("wait receive more data")

		if err = contextError(stream.Context()); err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		// Start receive chunk data
		chunkData := req.GetChunkData()
		size := len(chunkData)
		log.Println("chunk with seize: ", size)

		imageSize += size

		if imageSize > MaxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image is too large: [%d > %d]", imageSize, MaxImageSize))
		}

		_, err = imageData.Write(chunkData)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "data cannot write: %v", err))
		}
	}

	imageID, err := server.imageStore.Save(laptopId, imageType, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image to the store: %v", err))
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Printf("save image with laptop imageID: %s , size: %d", imageID, imageSize)
	return nil
}

func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		// timeout or cancel handle
		if err := contextError(stream.Context()); err != nil {
			return err
		}

		// start receive stream data
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.DataLoss, "cannot receive request: %v", err))
		}

		// Get data and log
		laptopId := req.GetLaptopId()
		laptopScore := req.GetScore()
		log.Printf("receive a rat-laptop stream with laptopID: %s, Score: %.2f", laptopId, laptopScore)

		// search data from server
		found, err := server.laptopStore.Find(laptopId)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot find laptop: %v", err))
		}
		if found == nil {
			return logError(status.Error(codes.NotFound, "laptopID is not find"))
		}

		rating, err := server.ratingStore.Add(laptopId, laptopScore)
		if err != nil {
			return logError(status.Error(codes.Internal, "not save laptop rat"))
		}

		res := &pb.RateLaptopResponse{
			LaptopId:    laptopId,
			RateCount:   rating.Count,
			AverageRate: rating.Sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "send stream data Failed: %v", err))
		}
	}
	return nil
}

func logError(err error) error {
	if err != nil {
		log.Println(err)
	}
	return err
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is cancel"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "request is DeadlineExceeded"))
	default:
		return nil
	}
}
