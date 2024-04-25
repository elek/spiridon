package dns

import "github.com/elek/spiridon/mud"

func Module(ball *mud.Ball) {
	mud.Provide[*Server](ball, NewServer)
	mud.Provide[*Service](ball, NewService)
}
