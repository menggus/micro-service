package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"library/v1/pb"
	"library/v1/service"
	"log"
	net2 "net"
	"net/http"
	"time"
)

const (
	secretKey    = "secret"
	timeDuration = 5 * time.Minute

	serverCertPem = "certificate/server-cert.pem"
	serverKeyPem  = "certificate/server-key.pem"
	CACertPem     = "certificate/ca-cert.pem"
)

func seedUser(store service.UserStore) error {
	err := CreateUser(store, "user1", "123456", "admin")
	if err != nil {
		return err
	}

	return CreateUser(store, "user2", "123456", "user")
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

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// load certificate of the CA who signed client's certificate
	pemClientCA, err := ioutil.ReadFile(CACertPem)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(serverCertPem, serverKeyPem)
	if err != nil {
		return nil, err
	}
	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool, // use verify client certificate set
	}

	return credentials.NewTLS(config), nil
}

func rungRPCServer(
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	jwtManager *service.JWTManager,
	enableTLS bool,
	listener net2.Listener,
) error {

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	serverOption := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}
	if enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			return fmt.Errorf("cannot load TLS credentials: %v", err)
		}

		serverOption = append(serverOption, grpc.Creds(tlsCredentials))
	}

	grpcServer := grpc.NewServer(serverOption...)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

	// Start grpc server
	log.Printf("Start gRPC server at %s, TLS=%t", listener.Addr().String(), enableTLS)
	return grpcServer.Serve(listener)
}

func runRESTServer(
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	jwtManager *service.JWTManager,
	enableTLS bool,
	listener net2.Listener,
) error {
	mux := runtime.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pb.RegisterAuthServiceHandlerServer(ctx, mux, authServer)
	if err != nil {
		return err
	}

	err = pb.RegisterLaptopServiceHandlerServer(ctx, mux, laptopServer)
	if err != nil {
		return err
	}

	log.Printf("Start REST server at %s, TLS=%t", listener.Addr().String(), enableTLS)

	if enableTLS {
		return http.ServeTLS(listener, mux, serverCertPem, serverKeyPem)
	}

	return http.Serve(listener, mux)
}

func main() {
	// Parse port param by flag package
	port := flag.Int("port", 0, "the server port")
	enableTLS := flag.Bool("tls", false, "enable SSL/TLS")
	serverType := flag.String("type", "grpc", "type of server (grpc/rest)")
	flag.Parse()
	log.Printf("start server on port %d, TLS=%t\n", *port, *enableTLS)

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

	// Create a net listener
	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net2.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start a listener: ", err)
	}
	if *serverType == "grpc" {
		err = rungRPCServer(authServer, laptopServer, jwtManager, *enableTLS, listener)
		if err != nil {
			log.Fatal("not start gRPC server")
		}
	} else {
		err = runRESTServer(authServer, laptopServer, jwtManager, *enableTLS, listener)
		if err != nil {
			log.Fatal("not start rest server")
		}
	}

}
