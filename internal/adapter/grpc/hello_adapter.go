package grpc

import (
	"context"

	"github.com/viquitorreis/my-grpc-proto/protogen/go/hello"
)

func (a *GrpcAdapter) SayHello(ctx context.Context, req *hello.HelloRequest) (*hello.HelloResponse, error) {
	greet := a.helloService.GenerateHello(req.Name)

	// colocando a string hello como resposta gRPC
	return &hello.HelloResponse{
		Message: greet,
	}, nil
}
