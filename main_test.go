package main

import (
	"context"
	"testing"
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"

	pb "github.com/washanhanzi/grpc-go-timeout/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TestServerSideTimeout will set a 3 seconds timeout in server interceptor
// but we are not dealing with the parent context timeout
// [ref](https://github.com/grpc/grpc-go/issues/5059)
func TestServerSideTimeout(t *testing.T) {
	wait := func(ctx context.Context) (*pb.HelloReply, error) {
		_, ok := ctx.Deadline()
		//check deadline
		if !ok {
			t.Fatal("deadline should be set")
		}
		//the request will wait 5 seconds
		time.Sleep(5 * time.Second)
		return &pb.HelloReply{Message: "World"}, nil

	}
	//an unary interceptor to set server timeout for every request
	timeoutInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		return handler(ctx, req)
	}
	interceptor := grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(timeoutInterceptor))
	ctx := context.Background()
	resp, err := hello(ctx, dialer(wait, interceptor))
	if err != nil {
		if v, ok := status.FromError(err); ok {
			if v.Code() != codes.DeadlineExceeded {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	}
	//request should success
	if resp.Message != "World" {
		t.Fail()
	}
}

// TestServerSideTimeoutWithEffort is same as TestServerSideTimeout
// but we are dealing with the parent context timeout
func TestServerSideTimeoutWithEffort(t *testing.T) {
	//server will wait 5 seconds and timeout
	wait := func(ctx context.Context) (*pb.HelloReply, error) {
		_, ok := ctx.Deadline()
		//check deadline
		if !ok {
			t.Fatal("deadline should be set")
		}
		//a 5 seconds operation
		sleepCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		select {
		//we are dealing with the server timeout
		case <-ctx.Done():
			return &pb.HelloReply{}, status.Error(codes.DeadlineExceeded, "server timeout")
		case <-sleepCtx.Done():
			return &pb.HelloReply{Message: "World"}, nil
		}
	}
	timeoutInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		return handler(ctx, req)
	}
	interceptor := grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(timeoutInterceptor))
	ctx := context.Background()
	_, err := hello(ctx, dialer(wait, interceptor))
	if err != nil {
		if v, ok := status.FromError(err); ok {
			if v.Code() != codes.DeadlineExceeded {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal("request should timeout on server side")
	}
}

// TestClientSideTimeout will set a timeout(deadline) on client side
func TestClientSideTimeout(t *testing.T) {
	//server will wait 5 seconds and make sure client will timeout
	h := func(ctx context.Context) (*pb.HelloReply, error) {
		//the client set deadline should be propagated to the request context
		_, ok := ctx.Deadline()
		if !ok {
			t.Fatal("deadline should be set")
		}
		//server will wait 5 seconds, but client's 1 second timeout should be triggered first
		time.Sleep(5 * time.Second)
		return &pb.HelloReply{Message: "world"}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := hello(ctx, dialer(h))
	if err != nil {
		if v, ok := status.FromError(err); ok {
			if v.Code() != codes.DeadlineExceeded {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal("request should timeout on client side")
	}

}

// TestClientSideTimeoutWithMetadata will incorrectly set a timeout by client using metadata
func TestClientSideTimeoutWithMetadata(t *testing.T) {
	h := func(ctx context.Context) (*pb.HelloReply, error) {
		_, ok := ctx.Deadline()
		if ok {
			t.Fatal("deadline should not exist")
		}
		time.Sleep(5 * time.Second)
		return &pb.HelloReply{Message: "World"}, nil
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "grpc-timeout", "1s")
	resp, err := hello(ctx, dialer(h))
	if err != nil {
		if v, ok := status.FromError(err); ok {
			if v.Code() != codes.DeadlineExceeded {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	}
	//request should success
	if resp.Message != "World" {
		t.Fail()
	}
}
