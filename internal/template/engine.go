// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package template // import "miniflux.app/v2/internal/template"

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"miniflux.app/v2/internal/locale"

	"github.com/gorilla/mux"
)

//go:embed templates/common/*.html
var commonTemplateFiles embed.FS

//go:embed templates/views/*.html
var viewTemplateFiles embed.FS

//go:embed templates/standalone/*.html
var standaloneTemplateFiles embed.FS

// Engine handles the templating system.
type Engine struct {
	templates map[string]*template.Template
	funcMap   *funcMap
}

// NewEngine returns a new template engine.
func NewEngine(router *mux.Router) *Engine {
	return &Engine{
		templates: make(map[string]*template.Template),
		funcMap:   &funcMap{router},
	}
}

// ParseTemplates parses template files embed into the application.
func (e *Engine) ParseTemplates() error {
	var commonTemplateContents strings.Builder

	dirEntries, err := commonTemplateFiles.ReadDir("templates/common")
	if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		fileData, err := commonTemplateFiles.ReadFile("templates/common/" + dirEntry.Name())
		if err != nil {
			return err
		}
		commonTemplateContents.Write(fileData)
	}
	ct := template.Must(template.New("common").Funcs(e.funcMap.Map()).Parse(commonTemplateContents.String()))

	dirEntries, err = viewTemplateFiles.ReadDir("templates/views")
	if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		templateName := dirEntry.Name()
		slog.Debug("Parsing template",
			slog.String("template_name", templateName),
		)

		e.templates[templateName] = template.Must(ct.ParseFS(viewTemplateFiles, "template/views/"+templateName))
	}

	dirEntries, err = standaloneTemplateFiles.ReadDir("templates/standalone")
	if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		templateName := dirEntry.Name()
		slog.Debug("Parsing template", slog.String("template_name", templateName))
		templatePath := "templates/standalone/" + templateName
		e.templates[templateName] = template.Must(template.New("").Funcs(e.funcMap.Map()).ParseFS(standaloneTemplateFiles, templatePath))
	}

	return nil
}

// Render process a template.
func (e *Engine) Render(name string, data map[string]interface{}) []byte {
	tpl, ok := e.templates[name]
	if !ok {
		panic("This template does not exists: " + name)
	}

	printer := locale.NewPrinter(data["language"].(string))

	// Functions that need to be declared at runtime.
	tpl.Funcs(template.FuncMap{
		"elapsed": func(timezone string, t time.Time) string {
			return elapsedTime(printer, timezone, t)
		},
		"t": func(key interface{}, args ...interface{}) string {
			switch k := key.(type) {
			case string:
				return printer.Printf(k, args...)
			case error:
				return k.Error()
			default:
				return ""
			}
		},
		"plural": func(key string, n int, args ...interface{}) string {
			return printer.Plural(key, n, args...)
		},
	})

	var b bytes.Buffer
	err := tpl.ExecuteTemplate(&b, "base", data)
	if err != nil {
		panic(err)
	}

	return b.Bytes()
}
