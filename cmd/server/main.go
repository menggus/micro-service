package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"library/v1/pb"
	"library/v1/service"
	"log"
	net2 "net"
	"time"
)

const (
	secretKey    = "secret"
	timeDuration = 5 * time.Minute
)

func seedUser(store service.UserStore) error {
	err := CreateUser(store, "admin", "admin", "role")
	if err != nil {
		return err
	}

	return CreateUser(store, "user1", "user1", "user")
}

func CreateUser(userStore service.UserStore, username string, password string, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}

	return userStore.Save(user)
}

func accessibleRoles() map[string][]string {
	const serverPath = "/techschool.proto.LaptopService/"
	return map[string][]string{
		serverPath + "CreateLaptop": {"admin"},
		serverPath + "UploadImage":  {"admin"},
		serverPath + "RatingLaptop": {"admin", "user"},
	}
}

func main() {
	// Parse port param by flag package
	port := flag.Int("port", 0, "the is server port")
	flag.Parse()
	log.Printf("start server on port %d\n", *port)

	// create user

	// Create a laptop service and  grpcServer
	// after, register laptop service in grpc server
	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	userStore := service.NewInMemoryUserStore()
	err := seedUser(userStore)
	if err != nil {
		log.Fatalf("cannot seed user: %s", err)
	}
	jwtManager := service.NewJWTManager(secretKey, timeDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

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
