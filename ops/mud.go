package ops

import "github.com/elek/spiridon/mud"

func Module(ball *mud.Ball) {
	mud.Provide[*Debug](ball, NewDebug)
	mud.Provide[*Metrics](ball, NewMetrics)
}
