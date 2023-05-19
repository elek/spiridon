package web

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"html/template"
	"io"
	"time"
)

var FuncMap = map[string]interface{}{
	"size": func(size int64) string {
		mb := size / 1024 / 1024
		if mb > 1024 {
			return fmt.Sprintf("%0.2f Gbyte", float64(mb)/1024)
		}
		return fmt.Sprintf("%d Mbyte", size/1024/1024)
	},
	"ms": func(d time.Duration) string {

		return fmt.Sprintf("%d ms", d.Milliseconds())
	},
}

func NewProdRender() echo.Renderer {
	return &Template{
		templates: template.Must(template.New("").Funcs(FuncMap).ParseFS(res, "template/*.html")).Funcs(FuncMap),
	}
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	b := bytes.NewBuffer([]byte{})
	err := t.templates.ExecuteTemplate(b, name+".html", data)
	if err != nil {
		return err
	}

	k := map[string]interface{}{
		"content": template.HTML(b.String()),
		"path":    c.Request().URL.Path,
		"wallet":  getCurrentWallet(c),
	}
	return t.templates.ExecuteTemplate(w, "frame.html", k)
}

func NewDevRender() echo.Renderer {
	return &DynamicTemplate{}
}

type DynamicTemplate struct{}

func (t *DynamicTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	templates, err := template.New("").Funcs(FuncMap).ParseGlob("web/template/*.html")
	if err != nil {
		return errors.WithStack(err)
	}

	b := bytes.NewBuffer([]byte{})
	err = templates.ExecuteTemplate(b, name+".html", data)
	if err != nil {
		return err
	}

	k := map[string]interface{}{
		"content": template.HTML(b.String()),
		"path":    c.Request().URL.Path,
		"wallet":  getCurrentWallet(c),
	}
	return templates.ExecuteTemplate(w, "frame.html", k)
}
