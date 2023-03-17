package check

import (
	"github.com/elek/spiridon/db"
	"github.com/pkg/errors"
	"time"
)

type Checkin struct {
}

func (p *Checkin) Name() string {
	return "recent checkin "
}

func (p *Checkin) Check(node db.Node) error {
	if time.Since(node.LastCheckIn) > 3*time.Hour {
		return errors.WithMessagef(CheckinFailed, "Last checkin is older than 3 hours")
	}
	return nil

}

var _ Checker = &Port{}
