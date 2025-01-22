package grpc

import (
	"fmt"
	"log"
	"net"

	"github.com/viquitorreis/my-grpc-go-server/internal/port"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/bank"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/hello"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/resiliency"
	"google.golang.org/grpc"
)

type GrpcAdapter struct {
	helloService      port.HelloServicePort
	bankService       port.BankServicePort
	resiliencyService port.ResiliencyServicePort
	grpcPort          int
	server            *grpc.Server
	hello.HelloServiceServer
	bank.BankServiceServer
	resiliency.ResiliencyServiceServer
	resiliency.ResiliencyWithMetadataServiceServer
}

func NewGrpcAdapter(helloService port.HelloServicePort, bankService port.BankServicePort, resServPort port.ResiliencyServicePort, grpcPort int) *GrpcAdapter {
	return &GrpcAdapter{
		helloService:      helloService,
		bankService:       bankService,
		resiliencyService: resServPort,
		grpcPort:          grpcPort,
	}
}

func (a *GrpcAdapter) Run() {
	var err error

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.grpcPort))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("gRPC server running on port %d", a.grpcPort)

	grpcServer := grpc.NewServer(
	// interceptor deve ficar dentro das options do server
	// grpc.ChainUnaryInterceptor(
	// 	interceptor.LogUnaryServerInterceptor(),
	// 	interceptor.BasicUnaryServerInterceptor(),
	// ),
	// grpc.ChainStreamInterceptor(
	// 	interceptor.LogStreamServerInterceptor(),
	// 	interceptor.BasicStreamServerInterceptor(),
	// ),
	)
	a.server = grpcServer

	hello.RegisterHelloServiceServer(grpcServer, a)
	bank.RegisterBankServiceServer(grpcServer, a)
	resiliency.RegisterResiliencyServiceServer(grpcServer, a)
	resiliency.RegisterResiliencyWithMetadataServiceServer(grpcServer, a)

	if err = grpcServer.Serve(listen); err != nil {
		log.Fatal(err)
	}
}

func (a *GrpcAdapter) Stop() {
	a.server.GracefulStop()
}
