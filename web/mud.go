package web

import (
	"github.com/elek/spiridon/config"
	"github.com/elek/spiridon/db"
	"github.com/elek/spiridon/mud"
)

func Module(ball *mud.Ball) {
	mud.Provide[*Server](ball, func(db *db.Persistence, cfg config.Config) *Server {
		return NewServer(db, cfg.WebPort, cfg.CookieSecret, cfg.Domain)
	})
}
