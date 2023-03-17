package check

import (
	"github.com/elek/spiridon/db"
	"github.com/pkg/errors"
)

var CheckinFailed = errors.New("")
var CheckinWarning = errors.New("")

type Checker interface {
	Name() string
	Check(node db.Node) error
}
