package satellite

import (
	"context"
	_ "embed"
	"fmt"
	bot "github.com/elek/spiridon/bot"
	"github.com/elek/spiridon/check"
	"github.com/elek/spiridon/db"
	"github.com/elek/spiridon/endpoint"
	"github.com/elek/spiridon/telemetry"
	"github.com/elek/spiridon/web"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/present"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"os"
	"os/signal"
	"storj.io/common/identity"
	"storj.io/common/pb"
	"storj.io/common/storj"
	"strings"
)

//go:embed satellites.txt
var knownSatellites []byte

func Run(config Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	registry := prometheus.NewRegistry()
	promExporter, err := otelprom.New(
		otelprom.WithRegisterer(registry))
	if err != nil {
		fmt.Println(err)
	}
	metricProvider := metric.NewMeterProvider(
		metric.WithReader(promExporter),
	)

	defer promExporter.Shutdown(ctx)
	cfg := identity.Config{
		CertPath: "identity.cert",
		KeyPath:  "identity.key",
	}
	ident, err := cfg.Load()
	if err != nil {
		return errors.WithStack(err)
	}

	orm, err := gorm.Open(postgres.Open(config.Db), &gorm.Config{})
	if err != nil {
		return err
	}

	err = orm.AutoMigrate(&db.Node{}, &db.Status{}, &db.Subscription{}, &db.Satellite{}, &db.SatelliteUsage{}, &db.Wallet{}, &db.Telemetry{})
	if err != nil {
		return err
	}

	persistence := db.NewPersistence(orm)
	err = InitSatellites(orm)
	if err != nil {
		return err
	}
	sub := db.NewSubscriptions(orm)

	rpc, err := NewRPCServer(ident, config.DrpcPort)
	if err != nil {
		return err
	}

	err = pb.DRPCRegisterHeldAmount(rpc.Mux, endpoint.HeldAmountEndpoint{})
	if err != nil {
		return errors.WithStack(err)
	}

	err = pb.DRPCRegisterNode(rpc.Mux, &endpoint.NodeEndpoint{
		Db: persistence,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	err = pb.DRPCRegisterNodeStats(rpc.Mux, &endpoint.NodeStatEndpoint{})
	if err != nil {
		return errors.WithStack(err)
	}

	err = pb.DRPCRegisterOrders(rpc.Mux, &endpoint.OrdersEndpoint{})
	if err != nil {
		return errors.WithStack(err)
	}

	if config.CookieSecret == "" {
		panic("Cookie secret is not set")
	}
	if config.Domain == "" {
		panic("Domain is not set")
	}
	webServer := web.NewServer(persistence, config.WebPort, config.CookieSecret, config.Domain, metricProvider)

	robot := bot.NewRobot(persistence, sub)
	tg, err := bot.NewTelegram(config.TelegramToken, robot)
	if err != nil {
		return err
	}

	not := bot.NewNotification(tg, sub, persistence)

	validator := check.NewValidator(persistence, not, ident)

	go func() {
		http.Handle(
			"/metrics", promhttp.HandlerFor(
				registry,
				promhttp.HandlerOpts{
					EnableOpenMetrics: true,
				}),
		)
		_ = http.ListenAndServe(":4444", nil)
	}()

	t, err := telemetry.NewTelemetry(persistence)
	if err != nil {
		return err
	}
	go t.Run(ctx)
	go http.ListenAndServe("0.0.0.0:9000", present.HTTP(monkit.Default))
	go validator.Loop(ctx)
	go tg.Run()
	go webServer.Run(ctx)

	return rpc.Run(ctx)
}

func InitSatellites(orm *gorm.DB) error {
	for _, sat := range strings.Split(string(knownSatellites), "\n") {
		if sat == "" {
			continue
		}
		parts := strings.SplitN(sat, " ", 2)
		url, err := storj.ParseNodeURL(parts[0])
		if err != nil {
			return errors.WithStack(err)
		}
		res := orm.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"address", "description"}),
		}).Create(&db.Satellite{
			ID: db.NodeID{
				NodeID: url.ID,
			},
			Address:     &url.Address,
			Description: &parts[1],
		})
		if res.Error != nil {
			return errors.WithStack(res.Error)
		}
	}
	return nil
}
