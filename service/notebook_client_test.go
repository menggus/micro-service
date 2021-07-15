package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"library/v1/pb"
	"library/v1/sample"
	"library/v1/serializer"
	"net"
	"testing"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	// Start a grpc server
	laptopServer, serveAddress := startTestLaptopServe(t)
	laptopClient := newTestLaptopClient(t, serveAddress)

	// Create some laptop object
	laptop := sample.NewLaptop()
	expectedID := laptop.Id
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	// Test laptopClient rpc request
	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedID, res.Id)

	// Check that the laptop is saved to the store
	other, err := laptopServer.Store.Find(res.Id)
	require.NoError(t, err)
	require.NotNil(t, other)

	// Check that the saved laptop is the same as the one we send
	requireSameLaptop(t, laptop, other)

}

// Create a grpc server
func startTestLaptopServe(t *testing.T) (*LaptopServer, string) {
	// 1. prepare customer Server
	laptopServer := NewLaptopServer(NewInMemoryLaptopStore())

	// 2. New a grpc Server and register
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	// 3. start a net.Listen protocol+port
	listener, err := net.Listen("tcp", ":0") // random availble prot
	require.NoError(t, err)

	// todo handle grpc server error
	// 4. start grpc server
	go grpcServer.Serve(listener)

	return laptopServer, listener.Addr().String()
}

// Create a grpc client
func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	// Create a grpc Dial
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure()) // which disables transport security for this ClientConn
	require.NoError(t, err)

	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1, laptop2 *pb.Laptop) {
	json1, err := serializer.ProtobufToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}
