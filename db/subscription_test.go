package db

import (
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"storj.io/common/storj"
	"testing"
)

func TestRearWriteSubscriptions(t *testing.T) {
	dsn := "host=localhost user=postgres dbname=test port=5432 sslmode=disable"
	orm, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, orm.AutoMigrate(&Subscription{}))
	require.NoError(t, err)

	id := storj.NodeID{}
	channelId := "1234"

	s := NewSubscriptions(orm)
	list, err := s.ListSubscriptions(TelegramSubscription, channelId)
	require.NoError(t, err)
	require.Len(t, list, 0)

	sub := Subscription{
		Destination:     channelId,
		DestinationType: TelegramSubscription,
		Base:            id.String(),
		BaseType:        NodeSubcription,
	}
	err = s.Subscribe(sub)
	require.NoError(t, err)

	list, err = s.ListSubscriptions(TelegramSubscription, channelId)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.Equal(t, id.String(), list[0].Base)

	err = s.Unsubscribe(sub)
	require.NoError(t, err)

	list, err = s.ListSubscriptions(TelegramSubscription, channelId)
	require.NoError(t, err)
	require.Len(t, list, 0)

}
