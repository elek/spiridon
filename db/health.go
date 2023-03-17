package db

import (
	"time"
)

type HealthyReport struct {
	AllHealthy bool
	Statuses   []HealthStatus
}

type HealthStatus struct {
	SatelliteID  NodeID
	OnlineScore  float64
	Disqualified time.Time
	SuspendedAt  time.Time
}
