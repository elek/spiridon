package endpoint

import (
	"context"
	"storj.io/common/pb"
)

type NodeStatEndpoint struct {
}

func (n *NodeStatEndpoint) DailyStorageUsage(ctx context.Context, request *pb.DailyStorageUsageRequest) (*pb.DailyStorageUsageResponse, error) {
	return &pb.DailyStorageUsageResponse{}, nil
}

func (n *NodeStatEndpoint) PricingModel(ctx context.Context, request *pb.PricingModelRequest) (*pb.PricingModelResponse, error) {
	return &pb.PricingModelResponse{
		EgressBandwidthPrice: 1,
	}, nil
}

func (n *NodeStatEndpoint) GetStats(context.Context, *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	return &pb.GetStatsResponse{
		OnlineScore: 1,

		UptimeCheck: &pb.ReputationStats{
			ReputationScore: 1,
		},
		AuditCheck: &pb.ReputationStats{
			SuccessCount:           0,
			ReputationScore:        1,
			ReputationAlpha:        1,
			ReputationBeta:         0,
			UnknownReputationScore: 1,
			UnknownReputationAlpha: 1,
			UnknownReputationBeta:  0,
		},
	}, nil
}
