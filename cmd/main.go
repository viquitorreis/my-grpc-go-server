package main

import (
	"log"

	mygrpc "github.com/viquitorreis/my-grpc-go-server/internal/adapter/grpc"
	app "github.com/viquitorreis/my-grpc-go-server/internal/application"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{})

	hs := &app.HelloService{}
	grpcAdapter := mygrpc.NewGrpcAdapter(hs, 9090)
	grpcAdapter.Run()
}
