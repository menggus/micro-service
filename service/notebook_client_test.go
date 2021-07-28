package service

import (
	"bufio"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"io"
	"library/v1/pb"
	"library/v1/sample"
	"library/v1/serializer"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// TestClientCreateLaptop test create laptop service
func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	// Start a grpc server
	laptopStore := NewInMemoryLaptopStore()
	serveAddress := startTestLaptopServe(t, laptopStore, nil)
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
	other, err := laptopStore.Find(res.Id)
	require.NoError(t, err)
	require.NotNil(t, other)

	// Check that the saved laptop is the same as the one we send
	requireSameLaptop(t, laptop, other)

}

// TestSearchLaptop  test search laptop service
func TestSearchLaptop(t *testing.T) {
	t.Parallel()

	// Create filter
	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   3.0,
		MinRam:      &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 4},
	}

	// Create store
	laptopStore := NewInMemoryLaptopStore()

	// Mock data
	exceptIDs := make(map[string]bool)
	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()
		switch i {
		case 0:
			laptop.PriceUsd = 3002
		case 1:
			laptop.Cpu.NumCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &pb.Memory{Unit: pb.Memory_MEGABYTE, Value: 2048}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumCores = 4
			laptop.Cpu.MinGhz = 3.2
			laptop.Cpu.MaxGhz = 4.5
			laptop.Ram = &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 16}
			exceptIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumCores = 6
			laptop.Cpu.MinGhz = 3.1
			laptop.Cpu.MaxGhz = 5.0
			laptop.Ram = &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 64}
			exceptIDs[laptop.Id] = true
		}
		err := laptopStore.Save(laptop)
		require.NoError(t, err)
		log.Println(exceptIDs)
	}

	// Create server and client
	serverAddress := startTestLaptopServe(t, laptopStore, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	// Create request
	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	// Receive stream-data and match exceptIDs
	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Contains(t, exceptIDs, res.GetLaptop().GetId())

		found += 1
	}

	require.Equal(t, len(exceptIDs), found)
}

// Create a grpc server
func startTestLaptopServe(t *testing.T, laptopStore LaptopStore, imageStore ImageStore) string {
	// 1. prepare customer Server
	laptopServer := NewLaptopServer(laptopStore, imageStore)

	// 2. New a grpc Server and register
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	// 3. start a net.Listen protocol+port
	listener, err := net.Listen("tcp", ":0") // random availble prot
	require.NoError(t, err)

	// todo handle grpc server error
	// 4. start grpc server
	go grpcServer.Serve(listener)

	return listener.Addr().String()
}

// Create a grpc client
func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	// Create a grpc Dial
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure()) // which disables transport security for this ClientConn
	require.NoError(t, err)

	return pb.NewLaptopServiceClient(conn)
}

// Check that the saved laptop is the same as the one we send
func requireSameLaptop(t *testing.T, laptop1, laptop2 *pb.Laptop) {
	json1, err := serializer.ProtobufToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}

// TestUploadImageClient test upload image
func TestUploadImageClient(t *testing.T) {
	t.Parallel()

	testImageFolder := "../tmp"

	laptopStore := NewInMemoryLaptopStore()
	imageStore := NewDiskImageStore(testImageFolder)

	// Require Save laptop no error
	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddress := startTestLaptopServe(t, laptopStore, imageStore)
	laptopClient := newTestLaptopClient(t, serverAddress)

	// file test
	filePath := fmt.Sprintf("%s/laptop.png", "../tmp")
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	//
	stream, err := laptopClient.UploadImage(context.Background())
	require.NoError(t, err)

	// 4. construct a pb.UploadImageRequest
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptop.GetId(),
				ImageType: filepath.Ext(filePath),
			},
		},
	}

	// 5. Send req by stream
	err = stream.Send(req)
	require.NoError(t, err)

	// 6.stream method send image data
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n], // attention buffer used
			},
		}

		size += len(buffer[:n])
		err = stream.Send(req)
		require.NoError(t, err)
	}

	res, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotZero(t, res.GetId())
	require.EqualValues(t, res.GetSize(), size)

	saveImagePath := fmt.Sprintf("%s/%s%s", testImageFolder, res.GetId(), filepath.Ext(filePath))
	require.FileExists(t, saveImagePath)
}
