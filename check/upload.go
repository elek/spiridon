package check

import (
	"context"
	"crypto/rand"
	"github.com/elek/spiridon/db"
	"github.com/elek/spiridon/util"
	"github.com/pkg/errors"
	"github.com/zeebo/errs"
	"storj.io/common/identity"
	"storj.io/common/pb"
	"storj.io/common/signing"
	"storj.io/common/storj"
	"time"
)

type Upload struct {
	identity *identity.FullIdentity
}

func (p *Upload) Name() string {
	return "upload"
}

func (p *Upload) Check(node db.Node) error {
	err := p.tryUpload(node)
	if err != nil {
		return errors.Wrap(CheckinFailed, "Couldn't upload piece: "+err.Error())
	}
	return nil
}

func (p *Upload) tryUpload(node db.Node) error {
	var data []byte
	ctx, done := context.WithTimeout(context.Background(), 60*time.Second)
	defer done()
	pieceID := storj.PieceID{}
	copy(pieceID[:], []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	dialer, err := util.GetDialerForIdentity(ctx, p.identity, false, false)
	if err != nil {
		return err
	}
	conn, err := dialer.DialNodeURL(ctx, storj.NodeURL{
		ID:      node.ID.NodeID,
		Address: node.Address,
	})
	if err != nil {
		return err
	}

	client := pb.NewDRPCReplaySafePiecestoreClient(conn)

	stream, err := client.Upload(ctx)
	if err != nil {
		return errs.Wrap(err)
	}
	defer stream.Close()

	signer := signing.SignerFromFullIdentity(p.identity)

	var serial pb.SerialNumber
	pub, pk, err := storj.NewPieceKey()
	if err != nil {
		return err
	}
	_, err = rand.Read(serial[:])
	if err != nil {
		return err
	}

	limit := &pb.OrderLimit{
		PieceId:         pieceID,
		SerialNumber:    serial,
		SatelliteId:     p.identity.ID,
		StorageNodeId:   node.ID.NodeID,
		Action:          pb.PieceAction_PUT,
		Limit:           int64(len(data)),
		OrderCreation:   time.Now(),
		OrderExpiration: time.Now().Add(24 * time.Hour),
		UplinkPublicKey: pub,
	}
	limit, err = signing.SignOrderLimit(ctx, signer, limit)
	if err != nil {
		return err
	}

	err = stream.Send(&pb.PieceUploadRequest{
		Limit:         limit,
		HashAlgorithm: pb.PieceHashAlgorithm_SHA256,
	})
	if err != nil {
		return err
	}

	order := &pb.Order{
		SerialNumber: serial,
		Amount:       int64(len(data)),
	}

	order, err = signing.SignUplinkOrder(ctx, pk, order)
	if err != nil {
		return err
	}

	h := pb.NewHashFromAlgorithm(pb.PieceHashAlgorithm_SHA256)

	err = stream.Send(&pb.PieceUploadRequest{
		Order: order,
		Chunk: &pb.PieceUploadRequest_Chunk{
			Offset: int64(0),
			Data:   data,
		},
		HashAlgorithm: pb.PieceHashAlgorithm_SHA256,
	})
	if err != nil {
		return err
	}

	_, err = h.Write(data)
	if err != nil {
		return err
	}

	uplinkHash, err := signing.SignUplinkPieceHash(ctx, pk, &pb.PieceHash{
		PieceId:       pieceID,
		PieceSize:     int64(len(data)),
		Hash:          h.Sum(nil),
		Timestamp:     limit.OrderCreation,
		HashAlgorithm: pb.PieceHashAlgorithm_SHA256,
	})
	if err != nil {
		return err
	}

	err = stream.Send(&pb.PieceUploadRequest{
		Done: uplinkHash,
	})
	if err != nil {
		return err
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		return err
	}
	return nil
}
