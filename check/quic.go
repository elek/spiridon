package check

import (
	"context"
	"github.com/elek/spiridon/db"
	"github.com/elek/spiridon/util"
	"github.com/pkg/errors"
	"storj.io/common/identity"
	"storj.io/common/pb"
	"storj.io/common/storj"
	"time"
)

type Quic struct {
	identity *identity.FullIdentity
}

func (p *Quic) Name() string {
	return "quic ping"
}

func (p *Quic) Check(node db.Node) error {
	ctx, done := context.WithTimeout(context.Background(), 1*time.Minute)
	defer done()
	dialer, err := util.GetDialerForIdentity(ctx, p.identity, false, true)
	if err != nil {
		return err
	}
	conn, err := dialer.DialNodeURL(ctx, storj.NodeURL{
		ID:      node.ID.NodeID,
		Address: node.Address,
	})
	if err != nil {
		return errors.Wrap(CheckinWarning, "Couldn't connect to the node with QUIC.")
	}
	client := pb.NewDRPCContactClient(conn)
	_, err = client.PingNode(ctx, &pb.ContactPingRequest{})
	if err != nil {
		return errors.Wrap(CheckinWarning, "Couldn't send PING using the QUIC connection.")
	}
	return nil
}

var _ Checker = &Port{}
