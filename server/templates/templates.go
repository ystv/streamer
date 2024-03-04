package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"time"
)

//go:embed *.tmpl
var tmpls embed.FS

type Templater struct {
	Version string
}

type Template string

const (
	FacebookHelpTemplate     Template = "facebookHelp.tmpl"
	ListTemplate             Template = "list.tmpl"
	MainTemplate             Template = "main.tmpl"
	RecallTemplate           Template = "recall.tmpl"
	ResumeTemplate           Template = "resume.tmpl"
	SaveTemplate             Template = "save.tmpl"
	YouTubeHelpTemplate      Template = "youtubeHelp.tmpl"
	ErrorTemplate            Template = "error.tmpl"
	NotFound404Template      Template = "404NotFound.tmpl"
	baseTemplate             Template = "_base.tmpl"
	navbarTemplate           Template = "_navbar.tmpl"
	statusComponentsTemplate Template = "_statusComponents.tmpl"
	stopComponentsTemplate   Template = "_stopComponents.tmpl"
)

func (t Template) String() string {
	return string(t)
}

func NewTemplate(version string) *Templater {
	return &Templater{
		Version: version,
	}
}

func (t *Templater) RenderTemplate(w io.Writer, data interface{}, mainTmpl Template) error {
	var err error

	t1 := template.New("_base.tmpl")

	t1.Funcs(t.getFuncMaps())

	t1, err = t1.ParseFS(tmpls, baseTemplate.String(), navbarTemplate.String(), statusComponentsTemplate.String(), stopComponentsTemplate.String(), mainTmpl.String())

	if err != nil {
		log.Printf("failed to get templates for template(RenderTemplate): %+v", err)
		return fmt.Errorf("failed to get templates for template(RenderTemplate): %+v", err)
	}

	return t1.Execute(w, data)
}

func (t *Templater) getFuncMaps() template.FuncMap {
	return template.FuncMap{
		"thisYear": func() int {
			return time.Now().Year()
		},
		"getVersion": func() string {
			return t.Version
		},
	}
}

// This section is for go template linter
var (
	AllTemplates = [][]string{
		{"facebookHelp.tmpl", "_base.tmpl", "_navbar.tmpl", "_statusComponents.tmpl", "_stopComponents.tmpl"},
		{"list.tmpl", "_base.tmpl", "_navbar.tmpl", "_statusComponents.tmpl", "_stopComponents.tmpl"},
		{"main.tmpl", "_base.tmpl", "_navbar.tmpl", "_statusComponents.tmpl", "_stopComponents.tmpl"},
		{"recall.tmpl", "_base.tmpl", "_navbar.tmpl", "_statusComponents.tmpl", "_stopComponents.tmpl"},
		{"resume.tmpl", "_base.tmpl", "_navbar.tmpl", "_statusComponents.tmpl", "_stopComponents.tmpl"},
		{"save.tmpl", "_base.tmpl", "_navbar.tmpl", "_statusComponents.tmpl", "_stopComponents.tmpl"},
		{"youtubeHelp.tmpl", "_base.tmpl", "_navbar.tmpl", "_statusComponents.tmpl", "_stopComponents.tmpl"},
		{"error.tmpl", "_base.tmpl", "_navbar.tmpl", "_statusComponents.tmpl", "_stopComponents.tmpl"},
		{"404NotFound.tmpl", "_base.tmpl", "_navbar.tmpl", "_statusComponents.tmpl", "_stopComponents.tmpl"},
	}

	_ = AllTemplates
)
