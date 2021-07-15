package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"library/v1/pb"
	"library/v1/sample"
	"testing"
)

// TestLaptopServer ..
func TestLaptopServer(t *testing.T) {
	t.Parallel()

	// OK, not ID
	laptopNoID := sample.NewLaptop()
	laptopNoID.Id = ""

	// Failure invalid ID
	laptopInvalidID := sample.NewLaptop()
	laptopInvalidID.Id = "invalid-id"

	// Duplicate ID
	laptopDuplicateID := sample.NewLaptop()
	storeDuplicateID := NewInMemoryLaptopStore()
	err := storeDuplicateID.Save(laptopDuplicateID)
	require.Nil(t, err)

	// testCase
	testCase := []struct {
		name   string
		laptop *pb.Laptop
		store  LaptopStore
		code   codes.Code
	}{
		{
			name:   "success_with_id",
			laptop: sample.NewLaptop(),
			store:  NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "success_no_id",
			laptop: laptopNoID,
			store:  NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "failure_invalid_id",
			laptop: laptopInvalidID,
			store:  NewInMemoryLaptopStore(),
			code:   codes.InvalidArgument,
		},
		{
			name:   "failure_duplicate_id",
			laptop: laptopDuplicateID,
			store:  storeDuplicateID,
			code:   codes.AlreadyExists,
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := NewLaptopServer(tc.store)
			req := &pb.CreateLaptopRequest{
				Laptop: tc.laptop,
			}
			res, err := server.CreateLaptop(context.Background(), req)
			if tc.code == codes.OK {
				require.NoError(t, err)     // assert err is nil
				require.NotNil(t, res)      // assert res is not nil
				require.NotEmpty(t, res.Id) // test uuid generate
				if len(tc.laptop.Id) > 0 {
					require.Equal(t, tc.laptop.Id, res.Id) // test exists uuid case
				}
			} else {
				require.Error(t, err)           // assert err is not nil
				require.Nil(t, res)             // assert res is nil
				st, ok := status.FromError(err) // when err is nil or contain GRPC.status, ok is true, otherwise ok is false
				require.True(t, ok)             // err contain GRPC.status, so ok is true
				require.Equal(t, tc.code, st.Code())

			}
		})
	}

}
