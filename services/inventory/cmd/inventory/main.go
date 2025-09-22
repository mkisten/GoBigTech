package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	inventorypb "inventory/v1" //github.com/mkisten/GoBigTech/services/
)

type server struct {
	inventorypb.UnimplementedInventoryServiceServer
}

func (s *server) GetStock(ctx context.Context, req *inventorypb.GetStockRequest) (*inventorypb.GetStockResponse, error) {
	return &inventorypb.GetStockResponse{ProductId: req.GetProductId(), Available: 42}, nil
}

func (s *server) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
	return &inventorypb.ReserveStockResponse{Success: req.GetQuantity() <= 42}, nil
}

func main() {
	l, err := net.Listen("tcp4", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(grpcSrv, &server{})
	log.Println("inventory gRPC listening on 127.0.0.1:50051")
	if err := grpcSrv.Serve(l); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
