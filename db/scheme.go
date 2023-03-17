package db

import "time"

type Node struct {
	ID           NodeID `gorm:"primaryKey"`
	FirstCheckIn time.Time
	LastCheckIn  time.Time
	FreeDisk     int64
	Address      string

	OperatorWallet string `json:"OperatorWallet,omitempty"`

	Version    string
	CommitHash string
	Timestamp  time.Time
	Release    bool

	Health string
}

type SatelliteUsage struct {
	NodeID       NodeID
	Satellite    Satellite
	SatelliteID  NodeID
	Disqualified time.Time
	Suspended    time.Time
	OnlineScore  float64
}

type Satellite struct {
	ID          NodeID `gorm:"primaryKey"`
	Address     *string
	Description *string
}

const (
	NodeSubcription    int = 0
	WalletSubscription int = 1

	TelegramSubscription int = 0
	NtfySubscription     int = 1
)

type Subscription struct {
	Destination     string `gorm:"primaryKey"`
	DestinationType int    `gorm:"primaryKey"`
	Base            string `gorm:"primaryKey"`
	BaseType        int    `gorm:"primaryKey"`
}

type Status struct {
	ID          NodeID `gorm:"primaryKey"`
	Check       string `gorm:"primaryKey"`
	LastChecked time.Time
	Error       string
	Duration    time.Duration
	Warning     bool
}

type CheckResult struct {
	Error    string
	Time     time.Time
	Duration time.Duration
	Warning  bool
}
