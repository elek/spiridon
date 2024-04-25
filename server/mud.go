package satellite

import (
	"github.com/elek/spiridon/config"
	"github.com/elek/spiridon/db"
	"github.com/elek/spiridon/endpoint"
	"github.com/elek/spiridon/mud"
	"github.com/pkg/errors"
	"storj.io/common/identity"
	"storj.io/common/pb"
)

func Module(ball *mud.Ball) {
	mud.Provide[*RPCServer](ball, func(cfg config.Config, ident *identity.FullIdentity, persistence *db.Persistence) (*RPCServer, error) {
		rpc, err := NewRPCServer(ident, cfg.DrpcPort)
		if err != nil {
			return rpc, err
		}

		err = pb.DRPCRegisterHeldAmount(rpc.Mux, endpoint.HeldAmountEndpoint{})
		if err != nil {
			return rpc, errors.WithStack(err)
		}

		err = pb.DRPCRegisterNode(rpc.Mux, &endpoint.NodeEndpoint{
			Db: persistence,
		})
		if err != nil {
			return rpc, errors.WithStack(err)
		}

		err = pb.DRPCRegisterNodeStats(rpc.Mux, &endpoint.NodeStatEndpoint{})
		if err != nil {
			return rpc, errors.WithStack(err)
		}

		err = pb.DRPCRegisterOrders(rpc.Mux, &endpoint.OrdersEndpoint{})
		if err != nil {
			return rpc, errors.WithStack(err)
		}
		return rpc, nil
	})

}
