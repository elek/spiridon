package endpoint

import (
	"context"
	"github.com/elek/spiridon/db"
	"github.com/ethereum/go-ethereum/common"
	"storj.io/common/identity"
	"storj.io/common/pb"
	"storj.io/common/rpc/rpcstatus"
	"time"
)

type NodeEndpoint struct {
	pb.DRPCNodeUnimplementedServer
	Db *db.Persistence
}

func (s *NodeEndpoint) GetTime(context.Context, *pb.GetTimeRequest) (*pb.GetTimeResponse, error) {
	return &pb.GetTimeResponse{
		Timestamp: time.Now(),
	}, nil
}

func (s *NodeEndpoint) CheckIn(ctx context.Context, req *pb.CheckInRequest) (*pb.CheckInResponse, error) {
	peerID, err := identity.PeerIdentityFromContext(ctx)
	if err != nil {
		return nil, rpcstatus.Error(rpcstatus.Unknown, "Failed to find identity in the context")
	}

	n := db.Node{
		ID: db.NodeID{
			NodeID: peerID.ID,
		},
		LastCheckIn:    time.Now(),
		FreeDisk:       req.Capacity.FreeDisk,
		Address:        req.Address,
		Version:        req.Version.Version,
		CommitHash:     req.Version.CommitHash,
		Timestamp:      req.Version.Timestamp,
		OperatorWallet: common.HexToAddress(req.Operator.Wallet).Hex(),
	}
	err = s.Db.UpdateCheckin(n)
	if err != nil {
		return nil, err
	}
	return &pb.CheckInResponse{
		PingNodeSuccess: true,
	}, nil
}
