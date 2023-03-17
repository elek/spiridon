package endpoint

import "storj.io/common/pb"

type OrdersEndpoint struct {
}

func (o *OrdersEndpoint) SettlementWithWindow(stream pb.DRPCOrders_SettlementWithWindowStream) error {
	storagenodeSettled := map[int32]int64{}
	for {
		s, err := stream.Recv()
		if err != nil {
			return stream.SendAndClose(&pb.SettlementWithWindowResponse{
				Status:        pb.SettlementWithWindowResponse_ACCEPTED,
				ActionSettled: storagenodeSettled,
			})
		}
		storagenodeSettled[int32(s.Limit.Action)] += s.Order.Amount
	}

}
