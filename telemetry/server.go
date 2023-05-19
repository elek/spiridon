package telemetry

import (
	"context"
	"github.com/elek/spiridon/db"
	"github.com/rs/zerolog/log"
	"storj.io/common/storj"
	"storj.io/common/telemetry"
	"strings"
	"time"
)

type Telemetry struct {
	db *db.Persistence
}

func NewTelemetry(db *db.Persistence) (Telemetry, error) {
	return Telemetry{
		db: db,
	}, nil
}

func (t Telemetry) Run(ctx context.Context) error {
	listen, err := telemetry.Listen("0.0.0.0:9000")
	if err != nil {
		return err
	}
	return listen.Serve(ctx, telemetry.HandlerFunc(func(application, instance string, key []byte, val float64) {

		parts := strings.SplitN(string(key), " ", 2)
		nodeID, err := storj.NodeIDFromString(instance)
		if err != nil {
			log.Err(err).Send()
			return
		}

		// todo: filter interesting metrics
		err = t.db.SaveTelemetry(db.Telemetry{
			NodeID: db.NodeID{
				NodeID: nodeID,
			},
			Key:      parts[0],
			Field:    parts[1],
			Received: time.Now(),
			Value:    val,
		})
		if err != nil {
			log.Err(err).Send()
		}
	}))
}
