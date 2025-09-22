package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	paymentpb "payment/v1" //github.com/alatair8/GoBigTech/services/
)

type server struct {
	paymentpb.UnimplementedPaymentServiceServer
}

func (s *server) ProcessPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
	return &paymentpb.ProcessPaymentResponse{Success: true, TransactionId: "tx_123456"}, nil
}

func main() {
	l, err := net.Listen("tcp4", "127.0.0.1:50052")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(grpcSrv, &server{})
	log.Println("payment gRPC listening on 127.0.0.1:50052")
	if err := grpcSrv.Serve(l); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
