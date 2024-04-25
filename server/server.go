package satellite

import (
	"context"
	_ "embed"
	bot "github.com/elek/spiridon/bot"
	"github.com/elek/spiridon/check"
	"github.com/elek/spiridon/config"
	"github.com/elek/spiridon/db"
	"github.com/elek/spiridon/dns"
	"github.com/elek/spiridon/mud"
	"github.com/elek/spiridon/ops"
	"github.com/elek/spiridon/web"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"storj.io/common/identity"
)

func Run(cfg config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	icfg := identity.Config{
		CertPath: "identity.cert",
		KeyPath:  "identity.key",
	}
	ident, err := icfg.Load()
	if err != nil {
		return errors.WithStack(err)
	}

	if cfg.CookieSecret == "" {
		panic("Cookie secret is not set")
	}
	if cfg.Domain == "" {
		panic("Domain is not set")
	}

	ball := mud.NewBall()
	mud.Supply[config.Config](ball, cfg)
	mud.Supply[*identity.FullIdentity](ball, ident)
	ops.Module(ball)
	db.Module(ball)
	bot.Module(ball)
	check.Module(ball)
	dns.Module(ball)
	web.Module(ball)
	for _, c := range mud.FindSelectedWithDependencies(ball, mud.All) {
		err := c.Init(ctx)
		if err != nil {
			return err
		}
	}

	eg := &errgroup.Group{}
	for _, c := range mud.Find(ball, mud.All) {
		err := c.Run(ctx, eg)
		if err != nil {
			return err
		}
	}

	return eg.Wait()
}
