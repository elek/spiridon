package web

import (
	"context"
	"embed"
	_ "embed"
	"github.com/elek/spiridon/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spacemonkeygo/monkit/v3"
	"net/http"
	"os"
	"storj.io/common/storj"
	"strconv"
	"strings"
)

var mon = monkit.Package()

//go:embed template
var res embed.FS

//go:embed dist/*
var dist embed.FS

type Server struct {
	nodes        *db.Nodes
	port         int
	cookieSecret string
	domain       string
}

func NewServer(nodes *db.Nodes, port int, cookieSecret string, domain string) *Server {
	return &Server{
		nodes:        nodes,
		port:         port,
		cookieSecret: cookieSecret,
		domain:       domain,
	}
}

func (s *Server) Run(ctx context.Context) error {
	e := echo.New()
	e.Debug = true
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			mon.Event("http_request", monkit.NewSeriesTag("path", strings.ReplaceAll(c.Request().URL.Path, "/", "_")))
			return next(c)
		}
	})
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
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

		all, err := s.nodes.ListALl()
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
		all, err := s.nodes.ListForWallet(param)
		if err != nil {
			return err
		}

		return c.Render(http.StatusOK, "nodes", map[string]interface{}{
			"nodes": all,
		})
	})
	e.GET("/nodes.json", func(c echo.Context) error {

		all, err := s.nodes.ListALl()
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, all)
	})
	e.GET("/satellites", func(c echo.Context) error {

		all, err := s.nodes.SatelliteList()
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

		all, err := s.nodes.SatelliteList()
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

		node, err := s.nodes.Get(id)
		if err != nil {
			return err
		}

		status, err := s.nodes.GetStatus(id)
		if err != nil {
			return err
		}
		satellites, err := s.nodes.GetUsedSatellites(id)
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
