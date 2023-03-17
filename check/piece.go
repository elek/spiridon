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

type Piece struct {
	identity *identity.FullIdentity
}

func (p *Piece) Name() string {
	return "piece upload"
}

func (p *Piece) Check(node db.Node) error {
	ctx, done := context.WithTimeout(context.Background(), 1*time.Minute)
	defer done()
	dialer, err := util.GetDialerForIdentity(ctx, p.identity, false, false)
	if err != nil {
		return err
	}
	conn, err := dialer.DialNodeURL(ctx, storj.NodeURL{
		ID:      node.ID.NodeID,
		Address: node.Address,
	})
	if err != nil {
		return errors.Wrap(CheckinFailed, "Couldn't connect to the node")
	}

	client := pb.NewDRPCPiecestoreClient(conn)
	_, err = client.Upload(ctx)
	if err != nil {
		return errors.Wrap(CheckinFailed, "Couldn't send PING on the connection")
	}
	return nil
}

var _ Checker = &Port{}
