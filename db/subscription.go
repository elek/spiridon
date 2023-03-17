package db

import (
	"gorm.io/gorm"
)

type Subscriptions struct {
	db *gorm.DB
}

func NewSubscriptions(db *gorm.DB) *Subscriptions {
	return &Subscriptions{db: db}
}

func (s *Subscriptions) Subscribe(subscription Subscription) error {
	res := s.db.Create(subscription)
	return res.Error
}

func (s *Subscriptions) Unsubscribe(subscription Subscription) error {
	res := s.db.Delete(subscription)
	return res.Error
}

func (s *Subscriptions) ListSubscriptions(destType int, dest string) (subs []Subscription, err error) {
	res := s.db.Where("destination_type = ? AND destination = ?", destType, dest).Find(&subs)
	return subs, res.Error
}

func (s *Subscriptions) GetTargets(subscriptionType int, id string) (subs []Subscription, err error) {
	res := s.db.Where("base_type = ? AND base = ?", subscriptionType, id).Find(&subs)
	return subs, res.Error
}
