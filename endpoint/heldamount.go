package endpoint

import (
	"context"
	"storj.io/common/pb"
	"storj.io/common/storj"
	"time"
)

type HeldAmountEndpoint struct {
	nodeID storj.NodeID
}

func (h HeldAmountEndpoint) GetPayStub(ctx context.Context, request *pb.GetHeldAmountRequest) (*pb.GetHeldAmountResponse, error) {
	return &pb.GetHeldAmountResponse{
		Period:    request.Period,
		CreatedAt: time.Now(),
	}, nil
}

func (h HeldAmountEndpoint) GetAllPaystubs(ctx context.Context, request *pb.GetAllPaystubsRequest) (*pb.GetAllPaystubsResponse, error) {
	return &pb.GetAllPaystubsResponse{
		Paystub: []*pb.GetHeldAmountResponse{},
	}, nil
}

func (h HeldAmountEndpoint) GetPayment(ctx context.Context, request *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	return &pb.GetPaymentResponse{
		NodeId:    h.nodeID,
		CreatedAt: time.Now(),
		Period:    request.Period,
		Amount:    0,
	}, nil
}

func (h HeldAmountEndpoint) GetAllPayments(ctx context.Context, request *pb.GetAllPaymentsRequest) (*pb.GetAllPaymentsResponse, error) {
	return &pb.GetAllPaymentsResponse{
		Payment: []*pb.GetPaymentResponse{},
	}, nil
}
