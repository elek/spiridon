package web

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"github.com/elek/spiridon/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"net/http"
	"os"
	"storj.io/common/storj"
	"strconv"
)

//go:embed template
var res embed.FS

//go:embed dist/*
var dist embed.FS

type Server struct {
	db             *db.Persistence
	port           int
	cookieSecret   string
	domain         string
	metricProvider *metricsdk.MeterProvider
}

func NewServer(nodes *db.Persistence, port int, cookieSecret string, domain string, provider *metricsdk.MeterProvider) *Server {
	return &Server{
		db:             nodes,
		port:           port,
		cookieSecret:   cookieSecret,
		domain:         domain,
		metricProvider: provider,
	}
}

func (s *Server) Run(ctx context.Context) error {
	e := echo.New()
	e.Debug = true
	e.Use(otelecho.Middleware("spiridon"))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	meter := s.metricProvider.Meter("github.com/elek/spiridon/web_request")

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		counter, err := meter.Int64Counter("request")
		fmt.Println(err)
		return func(c echo.Context) error {
			attrs := semconv.HTTPServerMetricAttributesFromHTTPRequest("", c.Request())
			counter.Add(context.TODO(), 1, attrs...)
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
		if param != getCurrentAddress(c) {
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
		if param != getCurrentAddress(c) {
			return c.String(http.StatusForbidden, "access denied")
		}
		uuid, err := uuid.NewV4()
		if err != nil {
			return errors.WithStack(err)
		}
		err = s.db.SaveWallet(db.Wallet{
			NtfyChannel: uuid.String(),
			Address:     getCurrentAddress(c),
		})
		if err != nil {
			return errors.WithStack(err)
		}
		c.Response().Header().Set("Location", "/wallet/"+getCurrentAddress(c))
		return c.Redirect(http.StatusSeeOther, "/wallet/"+getCurrentAddress(c))
	})
	e.POST("/wallet/:wallet/ntfy-reset", func(c echo.Context) error {

		param := c.Param("wallet")
		if param == "" {
			return c.String(http.StatusNotFound, "not found")
		}
		if param != getCurrentAddress(c) {
			return c.String(http.StatusForbidden, "access denied")
		}

		err := s.db.SaveWallet(db.Wallet{
			NtfyChannel: "",
			Address:     getCurrentAddress(c),
		})
		if err != nil {
			return errors.WithStack(err)
		}

		c.Response().Header().Set("Location", "/wallet/"+getCurrentAddress(c))
		return c.Redirect(http.StatusSeeOther, "/wallet/"+getCurrentAddress(c))
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
		satellites, err := s.db.GetUsedSatellites(id)
		if err != nil {
			return err
		}

		return c.Render(http.StatusOK, "node", map[string]interface{}{
			"node":       node,
			"status":     status,
			"satellites": satellites,
		})
	})
	go func() {
		<-ctx.Done()
		e.Listener.Close()
	}()
	return e.Start("0.0.0.0:" + strconv.Itoa(s.port))
}
