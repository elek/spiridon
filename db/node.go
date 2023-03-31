package db

import (
	"database/sql/driver"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"storj.io/common/storj"
	"time"
)

type Persistence struct {
	db *gorm.DB
}

func NewPersistence(db *gorm.DB) *Persistence {
	return &Persistence{db: db}
}

type NodeID struct {
	storj.NodeID
}

func (n *NodeID) GormDataType() string {
	return "text"
}

func (n *NodeID) Scan(value interface{}) error {
	id, err := storj.NodeIDFromString(value.(string))
	n.NodeID = id
	return err
}

// Value return json value, implement driver.Valuer interface
func (n *NodeID) Value() (driver.Value, error) {
	return n.String(), nil
}

func (n *Persistence) UpdateCheckin(node Node) error {
	p := Node{}
	result := n.db.First(&p, "id=?", node.ID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		node.FirstCheckIn = time.Now()
		result := n.db.Create(node)
		if result.Error != nil {
			return result.Error
		}
		return nil
	} else if result.Error != nil {
		return result.Error
	}
	res := n.db.Model(&node).Omit("first_check_in", "health").Updates(node)
	return res.Error
}

func (n *Persistence) Get(id NodeID) (Node, error) {
	node := Node{}
	result := n.db.Model(&Node{}).First(&node, "id=?", id.String())
	return node, result.Error
}

func (n *Persistence) ListNodes() ([]Node, error) {
	var nodes []Node
	n.db.Select([]string{"id", "first_check_in", "last_check_in", "free_disk", "address", "version", "commit_hash", "timestamp", "release", "health"}).Find(&nodes)
	return nodes, nil
}

func (n *Persistence) ListNodesInternal() ([]Node, error) {
	var nodes []Node
	n.db.Find(&nodes)
	return nodes, nil
}

func (n *Persistence) GetStatus(id NodeID) (map[string]Status, error) {
	var res []Status
	result := n.db.Where("id = ?", id).Find(&res)

	ret := map[string]Status{}
	for _, v := range res {
		ret[v.Check] = v
	}
	return ret, result.Error
}

func (n *Persistence) GetUsedSatellites(id NodeID) ([]SatelliteUsage, error) {
	res := []SatelliteUsage{}
	result := n.db.Preload("Satellite").Where("node_id = ?", id).Find(&res)
	return res, result.Error
}

func (n *Persistence) UpdateStatus(id NodeID, checked map[string]CheckResult) error {
	failed := false
	warning := false
	for k, c := range checked {
		s := Status{
			ID:          id,
			Check:       k,
			LastChecked: c.Time,
			Error:       c.Error,
			Duration:    c.Duration,
			Warning:     c.Warning,
		}
		res := n.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "check"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"last_checked": time.Now(), "error": c.Error, "duration": c.Duration, "warning": c.Warning}),
		}).Create(&s)

		if c.Error != "" {
			if c.Warning {
				warning = true
			} else {
				failed = true
			}
		}
		if res.Error != nil {
			return res.Error
		}
	}

	health := "healthy"
	if failed {
		health = "failed"
	} else if warning {
		health = "warning"
	}
	res := n.db.Model(&Node{}).Where("id = ?", id).Select("health").Updates(Node{Health: health})
	return res.Error
}

func (n *Persistence) UpdateSatellites(node Node, statuses []HealthStatus) error {
	n.db.Where("node_id = ?", node.ID).Delete(&SatelliteUsage{})
	for _, st := range statuses {
		res := n.db.Create(&SatelliteUsage{
			NodeID: node.ID,
			Satellite: Satellite{
				ID: st.SatelliteID,
			},
		})
		if res.Error != nil {
			return errors.WithStack(res.Error)
		}
	}
	return nil
}

type UsedSatellite struct {
	Satellite Satellite
	Count     int
}

func (n *Persistence) SatelliteList() ([]UsedSatellite, error) {
	res := []UsedSatellite{}
	rows, err := n.db.Raw("select satellites.id, satellites.address, satellites.description,count(satellite_usages) FROM satellite_usages JOIN satellites ON satellite_usages.satellite_id = satellites.id group by satellites.id order by count desc").Rows()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	for rows.Next() {
		us := UsedSatellite{}
		err := rows.Scan(&us.Satellite.ID, &us.Satellite.Address, &us.Satellite.Description, &us.Count)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		res = append(res, us)
	}
	return res, nil
}

func (n *Persistence) GetWalletWithNodes(walletAddress common.Address) (Wallet, []Node, error) {
	var nodes []Node
	n.db.Where("operator_wallet = ?", walletAddress.String()).Select([]string{"id", "first_check_in", "last_check_in", "free_disk", "address", "version", "commit_hash", "timestamp", "release", "health"}).Find(&nodes)

	var wallet Wallet
	n.db.First(&wallet, "address = ?", walletAddress.String())
	if wallet.Address == "" {
		wallet.Address = walletAddress.String()
	}
	return wallet, nodes, nil
}

func (n *Persistence) SaveWallet(w Wallet) error {
	res := n.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"ntfy_channel": w.NtfyChannel}),
		}).Create(w)
	return res.Error
}

func (n *Persistence) GetWallet(walletAddress common.Address) (Wallet, error) {
	var wallet Wallet
	n.db.First(&wallet, "address = ?", walletAddress.String())
	if wallet.Address == "" {
		wallet.Address = walletAddress.String()
	}
	return wallet, nil
}
