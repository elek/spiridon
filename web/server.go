package web

import (
	"context"
	"embed"
	_ "embed"
	"github.com/elek/spiridon/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"github.com/spacemonkeygo/monkit/v3"
	"github.com/zeebo/errs"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"math/rand"
	"net/http"
	"os"
	"storj.io/common/storj"
	"strconv"
	"time"
)

var mon = monkit.Package()

//go:embed template
var res embed.FS

//go:embed dist/*
var dist embed.FS

type Server struct {
	db           *db.Persistence
	port         int
	cookieSecret string
	domain       string
}

func NewServer(nodes *db.Persistence, port int, cookieSecret string, domain string) *Server {
	return &Server{
		db:           nodes,
		port:         port,
		cookieSecret: cookieSecret,
		domain:       domain,
	}
}

func (s *Server) Run(ctx context.Context) error {
	e := echo.New()
	e.Debug = true
	e.Use(otelecho.Middleware("spiridon"))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			defer func() {
				mon.DurationVal("web_request",
					monkit.NewSeriesTag("path", c.Path()),
					monkit.NewSeriesTag("method", c.Request().Method),
				).Observe(time.Since(start))
			}()
			return next(c)
		}
	})
	if _, err := os.Stat("web/template/index.html"); err == nil {
		e.Renderer = NewDevRender()
	} else {
		e.Renderer = NewProdRender()
	}

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", nil)
	})
	var contentHandler = echo.WrapHandler(http.FileServer(http.FS(dist)))
	//var contentRewrite = middleware.Rewrite(map[string]string{"/*": "/static/$1"})
	e.GET("/dist/*", contentHandler)

	e.GET("/nodes", func(c echo.Context) error {

		all, err := s.db.ListNodes()
		if err != nil {
			return err
		}

		return c.Render(http.StatusOK, "nodes", map[string]interface{}{
			"nodes": all,
		})
	})
	e.GET("/wallet/:wallet", func(c echo.Context) error {

		param := c.Param("wallet")
		if param == "" {
			return c.String(http.StatusNotFound, "not found")
		}
		if param != getCurrentWallet(c) {
			return c.String(http.StatusForbidden, "access denied")
		}
		wallet, all, err := s.db.GetWalletWithNodes(common.HexToAddress(param))
		if err != nil {
			return err
		}

		return c.Render(http.StatusOK, "wallet", map[string]interface{}{
			"nodes":  all,
			"wallet": wallet,
		})
	})
	e.POST("/wallet/:wallet/ntfy-generate", func(c echo.Context) error {

		param := c.Param("wallet")
		if param == "" {
			return c.String(http.StatusNotFound, "not found")
		}
		if param != getCurrentWallet(c) {
			return c.String(http.StatusForbidden, "access denied")
		}
		uuid, err := uuid.NewV4()
		if err != nil {
			return errors.WithStack(err)
		}
		err = s.db.SaveWallet(db.Wallet{
			NtfyChannel: uuid.String(),
			Address:     getCurrentWallet(c),
		})
		if err != nil {
			return errors.WithStack(err)
		}
		c.Response().Header().Set("Location", "/wallet/"+getCurrentWallet(c))
		return c.Redirect(http.StatusSeeOther, "/wallet/"+getCurrentWallet(c))
	})
	e.POST("/wallet/:wallet/ntfy-reset", func(c echo.Context) error {

		param := c.Param("wallet")
		if param == "" {
			return c.String(http.StatusNotFound, "not found")
		}
		if param != getCurrentWallet(c) {
			return c.String(http.StatusForbidden, "access denied")
		}

		err := s.db.SaveWallet(db.Wallet{
			NtfyChannel: "",
			Address:     getCurrentWallet(c),
		})
		if err != nil {
			return errors.WithStack(err)
		}

		c.Response().Header().Set("Location", "/wallet/"+getCurrentWallet(c))
		return c.Redirect(http.StatusSeeOther, "/wallet/"+getCurrentWallet(c))
	})
	e.GET("/db.json", func(c echo.Context) error {

		all, err := s.db.ListNodes()
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, all)
	})
	e.GET("/satellites", func(c echo.Context) error {

		all, err := s.db.SatelliteList()
		if err != nil {
			return err
		}

		return c.Render(http.StatusOK, "satellites", map[string]interface{}{
			"satellites": all,
		})
	})
	e.GET("/api", func(c echo.Context) error {
		return c.Render(http.StatusOK, "api", map[string]interface{}{})
	})
	e.GET("/satellites.json", func(c echo.Context) error {

		all, err := s.db.SatelliteList()
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, all)
	})
	LoginEndpoints(e, s.cookieSecret, s.domain)
	e.GET("/node/:id", func(c echo.Context) error {
		nodeID, err := storj.NodeIDFromString(c.Param("id"))
		if err != nil {
			return err
		}
		id := db.NodeID{
			NodeID: nodeID,
		}

		node, err := s.db.Get(id)
		if err != nil {
			return err
		}

		status, err := s.db.GetStatus(id)
		if err != nil {
			return err
		}

		owned := getCurrentWallet(c) == node.OperatorWallet

		satellites, err := s.db.GetUsedSatellites(id)
		if err != nil {
			return err
		}

		values := map[string]interface{}{
			"node":       node,
			"status":     status,
			"satellites": satellites,
			"owned":      owned,
		}

		if owned {
			collection, err := s.db.LatestStat(c.Request().Context(), id)
			if err != nil {
				return err
			}
			collection.Get("used_space,scope=storj.io/storj/storagenode/monitor", "recent")
			values["stat"] = map[string]float64{
				"usedSpace": collection.Get("used_space,scope=storj.io/storj/storagenode/monitor", "recent").Value,
			}
		}

		return c.Render(http.StatusOK, "node", values)
	})

	e.GET("/node/:id/stat", func(c echo.Context) error {
		nodeID, err := storj.NodeIDFromString(c.Param("id"))
		if err != nil {
			return err
		}
		id := db.NodeID{
			NodeID: nodeID,
		}

		node, err := s.db.Get(id)
		if err != nil {
			return err
		}

		if node.OperatorWallet != getCurrentWallet(c) {
			return c.String(http.StatusForbidden, "access denied")
		}
		collection, err := s.db.LatestStat(c.Request().Context(), id)
		if err != nil {
			return err
		}

		return c.Render(http.StatusOK, "node_stat", map[string]any{
			"stats": collection,
		})
	})

	e.GET("/node/:id/charts/ud", func(c echo.Context) error {
		nodeID, err := storj.NodeIDFromString(c.Param("id"))
		if err != nil {
			return err
		}
		id := db.NodeID{
			NodeID: nodeID,
		}

		node, err := s.db.Get(id)
		if err != nil {
			return err
		}

		if node.OperatorWallet != getCurrentWallet(c) {
			return c.String(http.StatusForbidden, "access denied")
		}

		line := charts.NewLine()

		line.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
			charts.WithTitleOpts(opts.Title{
				Title:    "Upload and download requests",
				Subtitle: "Number of upload / download requests per minutes",
			}))
		uStat, err := s.db.GetStat(c.Request().Context(), id, "upload_success_size_bytes,scope=storj.io/storj/storagenode/piecestore")
		if err != nil {
			return errs.Wrap(err)
		}

		dStat, err := s.db.GetStat(c.Request().Context(), id, "download_cancel_duration_ns,action=GET,scope=storj.io/storj/storagenode/piecestore")
		if err != nil {
			return errs.Wrap(err)
		}

		line.SetXAxis(statToLabels(dStat)).
			AddSeries("Upload", rateStat(uStat)).
			AddSeries("Download", rateStat(dStat)).
			SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
		return line.Render(c.Response())
	})

	go func() {
		<-ctx.Done()
		e.Listener.Close()
	}()
	return e.Start("0.0.0.0:" + strconv.Itoa(s.port))
}

func rateStat(stat db.Stat) []opts.LineData {
	items := make([]opts.LineData, 0)
	last := 0
	for i, s := range stat.Values {
		if i == 0 {
			continue
		}
		if stat.Values[i-1].Value > 0 && s.Value > 0 {
			items = append(items, opts.LineData{Value: s.Value - stat.Values[i-1].Value})
			last = s.Value - stat.Values[i-1].Value
		} else {
			items = append(items, opts.LineData{Value: last})
		}
	}
	return items
}

func statToLabels(stat db.Stat) []string {
	res := []string{}
	for i, s := range stat.Values {
		if i == 0 {
			continue
		}
		res = append(res, s.Received.Format("15:04"))
	}
	return res
}

// generate random data for line chart
func generateLineItems() []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.LineData{Value: rand.Intn(300)})
	}
	return items
}
