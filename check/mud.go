package check

import (
	"github.com/elek/spiridon/mud"
)

func Module(ball *mud.Ball) {
	mud.Provide[*Validator](ball, NewValidator)
}
