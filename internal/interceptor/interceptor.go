package interceptor

import (
	"context"
	"log"

	"github.com/viquitorreis/my-grpc-proto/protogen/go/hello"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/resiliency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LogUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp any, err error) {
		log.Println("[LOGGED BY SERVER INTERCEPTOR]", req)

		return handler(ctx, req)
	}
}

func BasicUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp any, err error) {
		switch request := req.(type) {
		case *hello.HelloRequest:
			request.Name = "[MODIFIED BY SERVER INTERCEPTOR - 1]" + request.Name
		}

		responseMetadata, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			responseMetadata = metadata.New(nil)
		}

		responseMetadata.Append("my-response-metadata-key-1", "my-response-metadata-value-1")
		responseMetadata.Append("my-response-metadata-key-2", "my-response-metadata-value-2")

		ctx = metadata.NewOutgoingContext(ctx, responseMetadata)

		grpc.SetHeader(ctx, responseMetadata)

		// handle request with modified message
		res, err := handler(ctx, req)
		if err != nil {
			return res, err
		}

		// modify response
		switch response := res.(type) {
		case *hello.HelloResponse:
			response.Message = "[MODIFIED BY SERVER INTERCEPTOR - 2]" + response.Message
		case *resiliency.ResiliencyReponse:
			response.DummyString = "[MODIFIED BY SERVER INTERCEPTOR - 3]" + response.DummyString
		}

		return res, nil
	}
}

func LogStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
	) error {
		log.Println("[LOGGED BY SERVER INTERCEPTOR]", info)

		return handler(srv, ss)
	}
}

// STREAMING INTERCEPTOR NEEDS STRUCT
type InterceptedServerStream struct {
	grpc.ServerStream
}

func BasicStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
	) error {
		// intercept stream
		interceptedServerStream := &InterceptedServerStream{
			ServerStream: ss,
		}

		// add response metadata
		responseMetadata, ok := metadata.FromOutgoingContext(interceptedServerStream.Context())
		if !ok {
			responseMetadata = metadata.New(nil)
		}

		responseMetadata.Append("my-response-metadata-key-1", "my-response-metadata-value-1")
		responseMetadata.Append("my-response-metadata-key-2", "my-response-metadata-value-2")

		interceptedServerStream.SetHeader(responseMetadata)

		// handle request
		return handler(srv, interceptedServerStream)
	}
}

// modifying messages from stream
func (s *InterceptedServerStream) RecvMsg(msg interface{}) error {
	if err := s.ServerStream.RecvMsg(msg); err != nil {
		return err
	}

	switch request := msg.(type) {
	case *hello.HelloRequest:
		request.Name = "[MODIFIED BY SERVER INTERCEPTOR - 4]" + request.Name
	}

	return nil
}

func (s *InterceptedServerStream) SendMsg(msg interface{}) error {
	switch response := msg.(type) {
	case *hello.HelloResponse:
		response.Message = "[MODIFIED BY SERVER INTERCEPTOR - 5]" + response.Message
	case *resiliency.ResiliencyReponse:
		response.DummyString = "[MODIFIED BY SERVER INTERCEPTOR - 6]" + response.DummyString
	}

	return s.ServerStream.SendMsg(msg)
}
