package check

import (
	"encoding/json"
	"github.com/elek/spiridon/db"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"time"
)

type Healthy struct {
	nodes *db.Persistence
}

func NewHealthy(nodes *db.Persistence) *Healthy {
	return &Healthy{
		nodes: nodes,
	}

}
func (p *Healthy) Name() string {
	return "health check"
}

func (p *Healthy) Check(node db.Node) error {
	resp, err := http.DefaultClient.Get("http://" + node.Address)
	if err != nil {
		return errors.Wrap(CheckinFailed, "Health information is not available on http://"+node.Address)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(CheckinFailed, "Health couldn't be read after a HTTP request")
	}
	hr := db.HealthyReport{}
	err = json.Unmarshal(raw, &hr)
	if err != nil {
		return errors.Wrap(CheckinFailed, "Health HTTP request didn't return with JSON: "+err.Error())
	}
	if hr.Statuses != nil && len(hr.Statuses) > 0 {
		err := p.nodes.UpdateSatellites(node, hr.Statuses)
		if err != nil {
			return err
		}

		var disqualified, suspended int
		minOnlineScore := float64(1)
		for _, status := range hr.Statuses {
			if status.Disqualified.After(time.Time{}) {
				disqualified++
			}
			if status.SuspendedAt.After(time.Time{}) {
				suspended++
			}
			if status.OnlineScore < minOnlineScore {
				minOnlineScore = status.OnlineScore
			}
		}
		if disqualified > 0 {
			return errors.Wrapf(CheckinFailed, "Storagenode is disqualified on %d (other) satellites.", disqualified)
		}
		if suspended > 0 {
			return errors.Wrapf(CheckinFailed, "Storagenode is suspended on %d (other) satellites.", suspended)
		}
		if minOnlineScore < 0.95 {
			return errors.Wrapf(CheckinWarning, "Other satellites maintain low online scores (%f) for this storagenode", minOnlineScore)
		}
	}
	if !hr.AllHealthy {
		return errors.Wrap(CheckinFailed, "One (or more) other (!) Satellites reported bad status (suspension, disqualification, etc.)")
	}
	return nil
}

var _ Checker = &Port{}
