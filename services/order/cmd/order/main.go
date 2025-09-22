package main

import (
	//"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	inventorypb "inventory/v1" //github.com/mkisten/GoBigTech/services/
	orderapi "order/api"       //github.com/mkisten/GoBigTech/services/
	paymentpb "payment/v1"     //github.com/mkisten/GoBigTech/services/
)

type orderServer struct {
	invClient inventorypb.InventoryServiceClient
	payClient paymentpb.PaymentServiceClient
}

func (s *orderServer) PostOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body orderapi.PostOrdersJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if body.UserId == "" || body.Items == nil || len(body.Items) == 0 ||
		body.Items[0].ProductId == "" || body.Items[0].Quantity <= 0 {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	first := body.Items[0]
	productID := first.ProductId
	qty := int32(first.Quantity)
	userID := body.UserId

	// Вызов Inventory
	if _, err := s.invClient.ReserveStock(ctx, &inventorypb.ReserveStockRequest{ProductId: productID, Quantity: qty}); err != nil {
		http.Error(w, "inventory error: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Вызов Payment
	if _, err := s.payClient.ProcessPayment(ctx, &paymentpb.ProcessPaymentRequest{
		OrderId: "order-123", UserId: userID, Amount: 100.0, Method: "card",
	}); err != nil {
		http.Error(w, "payment error: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Успех
	id := "order-123"
	status := "paid"
	resp := orderapi.Order{Id: id, UserId: userID, Status: status, Items: body.Items}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetOrders handles GET /orders
func (s *orderServer) GetOrders(w http.ResponseWriter, r *http.Request) {
	// For now, return an empty array of orders
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]orderapi.Order{})
}

func (s *orderServer) GetOrdersId(w http.ResponseWriter, r *http.Request, id string) {
	status := "paid"
	userID := "u1"
	items := []orderapi.OrderItem{{"p1", 2}}
	resp := orderapi.Order{Id: id, UserId: userID, Status: status, Items: items}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func main() {
	// Подключаемся к Inventory
	connInv, err := grpc.Dial("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial inventory: %v", err)
	}
	defer connInv.Close()

	// Подключаемся к Payment
	connPay, err := grpc.Dial("127.0.0.1:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial payment: %v", err)
	}
	defer connPay.Close()

	s := &orderServer{
		invClient: inventorypb.NewInventoryServiceClient(connInv),
		payClient: paymentpb.NewPaymentServiceClient(connPay),
	}

	r := chi.NewRouter()
	orderapi.HandlerFromMux(s, r)

	log.Println("order HTTP listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
