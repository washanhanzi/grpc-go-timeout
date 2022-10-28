package main

import (
	"context"
	"log"
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
// [ref](https://github.com/grpc/grpc-go/issues/5059)
func TestServerSideTimeout(t *testing.T) {
	//server will wait 5 seconds and timeout
	wait := func(ctx context.Context) (*pb.HelloReply, error) {
		// sleepCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// defer cancel()
		// select {
		// //if you don't deal with the parent ctx timeout, the timeout is ignored
		// case <-ctx.Done():
		// 	return &pb.HelloReply{}, status.Error(codes.DeadlineExceeded, "server timeout")
		// case <-sleepCtx.Done():
		_, ok := ctx.Deadline()
		log.Print(ok)
		time.Sleep(5 * time.Second)
		return &pb.HelloReply{Message: "world"}, nil

	}
	timeoutInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		return handler(ctx, req)
	}
	interceptor := grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(timeoutInterceptor))
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer(wait, interceptor)), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)
	_, err = client.SayHello(ctx, &pb.HelloRequest{Name: "hello"})
	if err != nil {
		if v, ok := status.FromError(err); ok {
			if v.Code() != codes.DeadlineExceeded {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal("request succeeded")
	}
}

// TestClientSideTimeoutWithGoCtx will set a timeout(deadline) on client side
func TestClientSideTimeoutWithGoCtx(t *testing.T) {
	//server will wait 5 seconds and make sure client will timeout
	h := func(ctx context.Context) (*pb.HelloReply, error) {
		_, ok := ctx.Deadline()
		log.Println(ok)
		time.Sleep(5 * time.Second)
		return &pb.HelloReply{Message: "world"}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer(h)), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)
	_, err = client.SayHello(ctx, &pb.HelloRequest{Name: "hello"})
	if err != nil {
		if v, ok := status.FromError(err); ok {
			if v.Code() != codes.DeadlineExceeded {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal("request succeeded")
	}

}

func TestClientSideTimeoutWithHeader(t *testing.T) {
	//server will wait 5 seconds and make sure client will timeout
	h := func(ctx context.Context) (*pb.HelloReply, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Println("haha")
		}
		log.Println(md)
		d, ok := ctx.Deadline()
		if !ok {
			return &pb.HelloReply{}, status.Error(codes.InvalidArgument, "deadline not set")
		}
		log.Println(d)
		time.Sleep(5 * time.Second)
		return &pb.HelloReply{Message: "world"}, nil
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "grpc-timeout", "1s")
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer(h)), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)
	_, err = client.SayHello(ctx, &pb.HelloRequest{Name: "hello"})
	if err != nil {
		if v, ok := status.FromError(err); ok {
			if v.Code() != codes.DeadlineExceeded {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal("request succeeded")
	}
}
