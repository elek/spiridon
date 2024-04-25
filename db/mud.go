package db

import (
	_ "embed"
	"github.com/elek/spiridon/config"
	"github.com/elek/spiridon/mud"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"storj.io/common/storj"
	"strings"
)

//go:embed satellites.txt
var knownSatellites []byte

func Module(ball *mud.Ball) {
	mud.Provide[*Refresher](ball, NewRefresher)
	mud.Provide[*gorm.DB](ball, func(cfg config.Config) (*gorm.DB, error) {
		orm, err := gorm.Open(postgres.Open(cfg.Db), &gorm.Config{})
		if err != nil {
			return orm, err
		}
		err = orm.AutoMigrate(&Node{}, &Status{}, &Subscription{}, &Satellite{}, &SatelliteUsage{}, &Wallet{}, &Telemetry{})
		if err != nil {
			return orm, err
		}
		return orm, nil
	})

	mud.Provide[*Persistence](ball, func(orm *gorm.DB) (*Persistence, error) {
		persistence := NewPersistence(orm)
		err := persistence.Init()
		if err != nil {
			return persistence, err
		}

		err = InitSatellites(orm)
		if err != nil {
			return persistence, err
		}
		return persistence, nil
	})

	mud.Provide[*Subscriptions](ball, NewSubscriptions)

}

func InitSatellites(orm *gorm.DB) error {
	for _, sat := range strings.Split(string(knownSatellites), "\n") {
		if sat == "" {
			continue
		}
		parts := strings.SplitN(sat, " ", 2)
		url, err := storj.ParseNodeURL(parts[0])
		if err != nil {
			return errors.WithStack(err)
		}
		res := orm.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"address", "description"}),
		}).Create(&Satellite{
			ID: NodeID{
				NodeID: url.ID,
			},
			Address:     &url.Address,
			Description: &parts[1],
		})
		if res.Error != nil {
			return errors.WithStack(res.Error)
		}
	}
	return nil
}
