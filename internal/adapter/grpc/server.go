package grpc

import (
	"fmt"
	"log"
	"net"

	"github.com/viquitorreis/my-grpc-go-server/internal/port"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/bank"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/hello"
	"google.golang.org/grpc"
)

type GrpcAdapter struct {
	helloService port.HelloServicePort
	bankService  port.BankServicePort
	grpcPort     int
	server       *grpc.Server
	hello.HelloServiceServer
	bank.BankServiceServer
}

func NewGrpcAdapter(helloService port.HelloServicePort, bankService port.BankServicePort, grpcPort int) *GrpcAdapter {
	return &GrpcAdapter{
		helloService: helloService,
		bankService:  bankService,
		grpcPort:     grpcPort,
	}
}

func (a *GrpcAdapter) Run() {
	var err error

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.grpcPort))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("gRPC server running on port %d", a.grpcPort)

	grpcServer := grpc.NewServer()
	a.server = grpcServer

	hello.RegisterHelloServiceServer(grpcServer, a)
	bank.RegisterBankServiceServer(grpcServer, a)

	if err = grpcServer.Serve(listen); err != nil {
		log.Fatal(err)
	}
}

func (a *GrpcAdapter) Stop() {
	a.server.GracefulStop()
}
