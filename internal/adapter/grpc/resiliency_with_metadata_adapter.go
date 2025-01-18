package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/resiliency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func dummyRequestMetadata(ctx context.Context) {
	if requestMetadata, ok := metadata.FromIncomingContext(ctx); ok {
		log.Println("Request metadata: ")
		for k, v := range requestMetadata {
			log.Printf(" %v: %v\n", k, v)
		}
	} else {
		log.Println("Request metadata not found")
	}
}

func dummyResponseMetadata() metadata.MD {
	md := map[string]string{
		"grpc-server-time":     fmt.Sprint(time.Now().Format("15:04:05")),
		"grpc-server-location": "Uberlandia, Brazil",
		"grpc-response-uuid":   uuid.New().String(),
	}

	return metadata.New(md)
}

// UnaryResiliencyWithMetadata(context.Context, *ResiliencyRequest) (*ResiliencyReponse, error)
func (a *GrpcAdapter) UnaryResiliencyWithMetadata(ctx context.Context, req *resiliency.ResiliencyRequest) (*resiliency.ResiliencyReponse, error) {
	log.Println("UnaryResiliencyWithMetadata called")

	randomDelay := time.Duration(rand.Intn(int(req.MaxDelaySecond-req.MinDelaySecond))+int(req.MinDelaySecond)) * time.Second
	time.Sleep(randomDelay)

	// res, err := a.ResiliencyWithMetadataServiceServer.UnaryResiliencyWithMetadata(ctx, req)
	// if err != nil {
	// 	return nil, err
	// }

	header := dummyResponseMetadata()
	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set header: %v", err)
	}

	return &resiliency.ResiliencyReponse{
		DummyString: fmt.Sprintf("Response after %v seconds", randomDelay.Seconds()),
	}, nil
}

func (a *GrpcAdapter) ServerStreamResiliencyWithMetadata(req *resiliency.ResiliencyRequest, stream resiliency.ResiliencyWithMetadataService_ServerStreamResiliencyWithMetadataServer) error {
	log.Println("ServerStreamResiliencyWithMetadata called")
	context := stream.Context()

	dummyRequestMetadata(context)
	if err := stream.SendHeader(dummyResponseMetadata()); err != nil {
		log.Println("Error while sending response metadata: ", err)
	}

	for {
		select {
		case <-context.Done():
			log.Println("client cancelou o streaming")
			return nil
		default:
			str, sts := a.resiliencyService.GenerateResiliency(req.MinDelaySecond, req.MaxDelaySecond, req.StatusCodes)

			if err := generateErrStatus(sts); err != nil {
				return err
			}

			stream.Send(&resiliency.ResiliencyReponse{
				DummyString: str,
			})
		}
	}
}

func (a *GrpcAdapter) ClientStreamResiliencyWithMetadata(stream resiliency.ResiliencyWithMetadataService_ClientStreamResiliencyWithMetadataServer) error {
	log.Println("ClientStreamResiliencyWithMetadata called")

	i := 0

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("ClientStreamResiliency finalizado")

				res := resiliency.ResiliencyReponse{
					DummyString: fmt.Sprintf("Recebeu %v requisições do client", strconv.Itoa(i)),
				}

				// precisamos enviar o header antes de terminar o streaming
				if err := stream.SendHeader(dummyResponseMetadata()); err != nil {
					log.Println("Error while sendint response metadata: ", err)
				}

				return stream.SendAndClose(&res)
			}

			log.Printf("Error receiving stream: %v", err)
			return err
		}

		// processando metadados
		context := stream.Context()
		dummyRequestMetadata(context)

		if req != nil {
			_, sts := a.resiliencyService.GenerateResiliency(req.MinDelaySecond, req.MaxDelaySecond, req.StatusCodes)

			if err := generateErrStatus(sts); err != nil {
				log.Printf("Error generating error status: %v", err)
				return err
			}
		}

		i = i + 1
	}
}

func (a *GrpcAdapter) BidirectionalStreamResiliencyWithMetadata(stream resiliency.ResiliencyWithMetadataService_BidirectionalStreamResiliencyWithMetadataServer) error {
	log.Println("BidirectionalStreamResiliencyWithMetadata called")

	context := stream.Context()
	if err := stream.SendHeader(dummyResponseMetadata()); err != nil {
		log.Println("Error while sendint response metadata: ", err)
	}

	for {
		select {
		case <-context.Done():
			log.Println("client cancelou o streaming")
			return nil
		default:
			req, err := stream.Recv()

			if err == io.EOF {
				return nil
			}

			if err != nil {
				log.Fatalln("Erro ao ler stream do client:", err)
			}

			dummyRequestMetadata(context)

			str, sts := a.resiliencyService.GenerateResiliency(req.MinDelaySecond, req.MaxDelaySecond, req.StatusCodes)

			if err := generateErrStatus(sts); err != nil {
				return err
			}

			err = stream.Send(&resiliency.ResiliencyReponse{
				DummyString: str,
			})

			if err != nil {
				log.Fatalln("Erro ao enviar response da stream para o client:", err)
			}
		}
	}
}
