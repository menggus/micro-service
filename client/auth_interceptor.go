package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

type AuthInterceptor struct {
	authClient  *AuthClient
	authMethod  map[string]bool
	accessToken string
}

func NewAuthInterceptor(authClient *AuthClient, authMethod map[string]bool, refreshDuration time.Duration) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient: authClient,
		authMethod: authMethod,
	}

	err := interceptor.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil, err
	}

	return interceptor, nil
}

func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		log.Printf("--> Unary interceptor: %s", method)

		if interceptor.authMethod[method] {
			return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (interceptor *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {
		log.Printf("--> Stream interceptor: %s", method)

		if interceptor.authMethod[method] {
			return streamer(interceptor.attachToken(ctx), desc, cc, method, opts...)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (interceptor *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	// Add metadata params
	return metadata.AppendToOutgoingContext(ctx, "authorization", interceptor.accessToken)
}

func (interceptor *AuthInterceptor) scheduleRefreshToken(duration time.Duration) error {
	err := interceptor.RefreshToken()
	if err != nil {
		return err
	}

	// for{}  refresh the token
	go func() {
		wait := duration
		for {
			time.Sleep(wait)
			err := interceptor.RefreshToken()
			if err != nil {
				wait = time.Second
			} else {
				wait = duration
			}
		}
	}()
	return nil
}

func (interceptor *AuthInterceptor) RefreshToken() error {
	accessToken, err := interceptor.authClient.Login()
	if err != nil {
		return err
	}

	interceptor.accessToken = accessToken
	log.Printf("token refresh: %v", accessToken)

	return nil
}
