package db

import (
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math/rand"
	"storj.io/common/storj"
	"testing"
	"time"
)

func TestSatellites(t *testing.T) {
	rand.Seed(time.Now().Unix())
	dsn := "host=localhost user=postgres dbname=test port=5432 sslmode=disable"
	orm, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, orm.AutoMigrate(&Node{}, &Satellite{}, &SatelliteUsage{}))
	require.NoError(t, err)
	orm.Logger = &DBLog{}
	res := orm.Where("true=true").Delete(&SatelliteUsage{})
	require.NoError(t, res.Error)
	res = orm.Where("true=true").Delete(&Node{})
	require.NoError(t, res.Error)
	res = orm.Where("true=true").Delete(&Satellite{})
	require.NoError(t, res.Error)

	id := randomNodeID()

	sat1 := randomNodeID()
	sat2 := randomNodeID()

	address := "localhost:1234"
	orm.Create(&Satellite{
		ID:      sat2,
		Address: &address,
	})

	n := NewNodes(orm)
	node := Node{
		ID: id,
	}
	err = n.UpdateCheckin(node)
	require.NoError(t, err)

	err = n.UpdateSatellites(node, []HealthStatus{
		{
			SatelliteID: sat1,
		},
		{
			SatelliteID: sat2,
		},
	})
	require.NoError(t, err)

	satellites, err := n.GetUsedSatellites(id)
	require.NoError(t, err)
	require.Len(t, satellites, 2)
	require.Equal(t, satellites[0].SatelliteID.String(), sat1.String())

	withAddress := satellites[0]
	if satellites[0].SatelliteID != sat2 {
		withAddress = satellites[1]
	}
	require.Equal(t, "localhost:1234", *withAddress.Satellite.Address)

	list, err := n.SatelliteList()
	require.NoError(t, err)

	require.Len(t, list, 2)
}

func randomNodeID() NodeID {
	res := storj.NodeID{}
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	copy(res[:], randomBytes)
	return NodeID{
		NodeID: res,
	}
}
