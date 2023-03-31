package db

func (n *Persistence) SaveTelemetry(t Telemetry) error {
	tx := n.db.Create(&t)
	return tx.Error
}
