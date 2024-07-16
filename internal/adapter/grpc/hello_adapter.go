package grpc

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/viquitorreis/my-grpc-proto/protogen/go/hello"
)

func (a *GrpcAdapter) SayHello(ctx context.Context, req *hello.HelloRequest) (*hello.HelloResponse, error) {
	greet := a.helloService.GenerateHello(req.Name)

	// colocando a string hello como resposta gRPC
	return &hello.HelloResponse{
		Message: greet,
	}, nil
}

func (a *GrpcAdapter) SayManyHello(req *hello.HelloRequest, stream hello.HelloService_SayManyHelloServer) error {
	for i := 0; i < 100; i++ {
		greet := a.helloService.GenerateHello(req.Name)
		res := fmt.Sprintf("[%d] %s", i, greet)
		stream.Send(
			&hello.HelloResponse{
				Message: res,
			},
		)

		// rand.NewSource(time.Now().UnixNano())                        // Generates a new seed for the random number generator
		// randomDuration := rand.Intn(50) + 1                          // Generates a number between 1 and 50
		// time.Sleep(time.Duration(randomDuration) * time.Millisecond) // Use the random number as the sleep duration
	}

	return nil
}

func (a *GrpcAdapter) SayHelloToEveryone(stream hello.HelloService_SayHelloToEveryoneServer) error {
	res := ""

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(
				&hello.HelloResponse{
					Message: res,
				},
			)
		}

		if err != nil {
			log.Fatalln("Error receiving stream:", err)
		}

		greet := a.helloService.GenerateHello(req.Name)

		res += greet + "\n"
	}
}

func (a *GrpcAdapter) SayHelloContinuous(stream hello.HelloService_SayHelloContinuousServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalln("Error receiving stream:", err)
		}

		// para cada request vamos gerar uma resposta e retornar para o client gRPC imediatamente
		greet := a.helloService.GenerateHello(req.Name)
		err = stream.Send(
			&hello.HelloResponse{
				Message: greet,
			},
		)

		if err != nil {
			log.Fatalln("Error sending stream:", err)
		}
	}
}
