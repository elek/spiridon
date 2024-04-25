package satellite

import (
	"context"
	_ "embed"
	"fmt"
	bot "github.com/elek/spiridon/bot"
	"github.com/elek/spiridon/check"
	"github.com/elek/spiridon/db"
	"github.com/elek/spiridon/dns"
	"github.com/elek/spiridon/endpoint"
	"github.com/elek/spiridon/web"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/present"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net"
	"net/http"
	"os"
	"os/signal"
	"storj.io/common/identity"
	"storj.io/common/pb"
	"storj.io/common/storj"
	"strings"
	"time"
)

//go:embed satellites.txt
var knownSatellites []byte

func Run(config Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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

	err = persistence.Init()
	if err != nil {
		return err
	}

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
	webServer := web.NewServer(persistence, config.WebPort, config.CookieSecret, config.Domain)

	robot := bot.NewRobot(persistence, sub)
	tg, err := bot.NewTelegram(config.TelegramToken, robot)
	if err != nil {
		return err
	}

	not := bot.NewNotification(tg, sub, persistence)

	validator := check.NewValidator(persistence, not, ident)

	go func() {
		http.Handle(
			"/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data := make(map[string][]string)
				var components []string
				monkit.Default.Stats(func(key monkit.SeriesKey, field string, val float64) {
					components = components[:0]

					measurement := sanitize(key.Measurement)
					for tag, tagVal := range key.Tags.All() {
						components = append(components,
							fmt.Sprintf("%s=%q", sanitize(tag), sanitize(tagVal)))
					}
					components = append(components,
						fmt.Sprintf("field=%q", sanitize(field)))

					data[measurement] = append(data[measurement],
						fmt.Sprintf("{%s} %g", strings.Join(components, ","), val))
				})

				for measurement, samples := range data {
					_, _ = fmt.Fprintln(w, "# TYPE", measurement, "gauge")
					for _, sample := range samples {
						_, _ = fmt.Fprintf(w, "%s%s\n", measurement, sample)
					}
				}

			}))
		_ = http.ListenAndServe(":4444", nil)
	}()

	go http.ListenAndServe("0.0.0.0:9000", present.HTTP(monkit.Default))
	go validator.Loop(ctx)
	go tg.Run()
	go webServer.Run(ctx)
	go func() {
		time.Sleep(5 * time.Minute)
		for {
			persistence.RefreshViews(ctx)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Minute):
				continue
			}
		}
	}()
	go func() {

		server := &dns.Server{
			Db: persistence,
		}
		lis, err := net.Listen("tcp", "127.0.0.1:8053")
		log.Error().Err(err)
		grpcServer := grpc.NewServer()
		dns.RegisterDnsServiceServer(grpcServer, server)
		fmt.Println("listening")
		err = grpcServer.Serve(lis)
		log.Error().Err(err)
		fmt.Println("done")

	}()
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

// sanitize formats val to be suitable for prometheus.
func sanitize(val string) string {
	// https://prometheus.io/docs/concepts/data_model/
	// specifies all metric names must match [a-zA-Z_:][a-zA-Z0-9_:]*
	// Note: The colons are reserved for user defined recording rules.
	// They should not be used by exporters or direct instrumentation.
	if val == "" {
		return ""
	}
	if '0' <= val[0] && val[0] <= '9' {
		val = "_" + val
	}
	return strings.Map(func(r rune) rune {
		switch {
		case 'a' <= r && r <= 'z':
			return r
		case 'A' <= r && r <= 'Z':
			return r
		case '0' <= r && r <= '9':
			return r
		default:
			return '_'
		}
	}, val)
}
