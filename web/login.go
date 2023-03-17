package web

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/spruceid/siwe-go"
	"net/http"
	"path/filepath"
	"strings"
)

func LoginEndpoints(e *echo.Echo, cookieSecret string, domain string) {
	e.Use(session.MiddlewareWithConfig(session.Config{
		Store: sessions.NewCookieStore([]byte(cookieSecret)),
	}))
	e.GET("/logout", func(c echo.Context) error {
		sess, _ := session.Get("auth", c)
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
		}
		sess.Values["address"] = ""
		err := sess.Save(c.Request(), c.Response())
		if err != nil {
			return err
		}
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	})
	e.GET("/login", func(c echo.Context) error {
		address := getCurrentAddress(c)
		var js = ""
		entries, err := dist.ReadDir("dist/assets")
		if err != nil {
			return err
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), "js") {
				js = filepath.Join("/dist/assets", e.Name())
				break
			}
		}
		return c.Render(http.StatusOK, "login", map[string]interface{}{
			"address": address,
			"js":      js,
		})
	})
	e.POST("/login", func(c echo.Context) error {
		rawMessage, _ := hex.DecodeString(c.FormValue("message"))
		signature := c.FormValue("signature")
		a := string(rawMessage)
		fmt.Println(a)
		fmt.Println(hex.EncodeToString([]byte(rawMessage)))
		message, err := siwe.ParseMessage(a)
		if err != nil {
			return err
		}

		publicKey, err := message.Verify(signature, &domain, nil, nil)
		if err != nil {
			return err
		}
		address := crypto.PubkeyToAddress(*publicKey).String()
		sess, _ := session.Get("auth", c)
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
		}
		sess.Values["address"] = address
		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, "login", map[string]interface{}{
			"loggedin": true,
			"address":  address,
		})
	})
}

func getCurrentAddress(c echo.Context) string {
	sess, _ := session.Get("auth", c)
	var address string
	if sess.Values["address"] != nil {
		address = sess.Values["address"].(string)
	}
	return address
}
