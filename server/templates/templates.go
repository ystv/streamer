package templates

import (
	"embed"
	"html/template"
	"io"
	"time"
)

//go:embed *.tmpl
var tpls embed.FS

type (
	Templater struct {
		dashboard *template.Template
	}
	BaseParams struct {
		SystemTime time.Time
	}
	DashboardParams struct {
		Base BaseParams
	}
)

var funcs = template.FuncMap{
	"cleantime": cleanTime,
}

func cleanTime(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

func parse(file string) *template.Template {
	return template.Must(
		template.New("base.tmpl").Funcs(funcs).ParseFS(tpls, "base.tmpl", file))
}

func NewMain() *Templater {
	return &Templater{
		dashboard: parse("main.tmpl"),
	}
}

func NewList() *Templater {
	return &Templater{
		dashboard: parse("list.tmpl"),
	}
}

func NewRecall() *Templater {
	return &Templater{
		dashboard: parse("recall.tmpl"),
	}
}

func NewResume() *Templater {
	return &Templater{
		dashboard: parse("resume.tmpl"),
	}
}

func NewSave() *Templater {
	return &Templater{
		dashboard: parse("save.tmpl"),
	}
}

func NewFacebookHelp() *Templater {
	return &Templater{
		dashboard: parse("facebookHelp.tmpl"),
	}
}

func NewYouTubeHelp() *Templater {
	return &Templater{
		dashboard: parse("youtubeHelp.tmpl"),
	}
}

func (t *Templater) Dashboard(w io.Writer, p DashboardParams) error {
	return t.dashboard.Execute(w, p)
}
