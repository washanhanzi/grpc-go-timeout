package main

import (
	"context"
	"log"
	"net"

	pb "github.com/washanhanzi/grpc-go-timeout/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func main() {

}

type handlerFunc func(context.Context) (*pb.HelloReply, error)

type server struct {
	customHandler handlerFunc
	pb.UnimplementedGreeterServer
}

// NewServer returns a GreeterServer with a custom handler
func NewServer(fn handlerFunc) *server {
	return &server{customHandler: fn}
}

// SayHello implements GreeterServer.SayHello, and execute custom handler
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return s.customHandler(ctx)
}

// dialer return a dialer for grpc.WithContextDialer
func dialer(fn handlerFunc, opts ...grpc.ServerOption) func(context.Context, string) (net.Conn, error) {
	ctrl := NewServer(fn)
	list := bufconn.Listen(1204 * 1024)
	s := grpc.NewServer(opts...)
	pb.RegisterGreeterServer(s, ctrl)
	go func() {
		if err := s.Serve(list); err != nil {
			log.Fatal(err)
		}
	}()
	return func(context.Context, string) (net.Conn, error) {
		return list.Dial()
	}
}

// hello make the client grpc request
func hello(ctx context.Context, dialer func(context.Context, string) (net.Conn, error)) (*pb.HelloReply, error) {
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)
	return client.SayHello(ctx, &pb.HelloRequest{Name: "Hello"})
}
