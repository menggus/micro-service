package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"library/v1/pb"
	"log"
	"time"
)

type LaptopServer struct {
	Store                               LaptopStore
	pb.UnimplementedLaptopServiceServer // 必须嵌入以具有向前兼容的实现
}

// NewLaptopServer create *LaptopServer
func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{Store: store}
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
		id, err := uuid.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	// todo something
	// Set timeout test
	time.Sleep(6 * time.Second)

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
	err := server.Store.Save(laptop)
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
