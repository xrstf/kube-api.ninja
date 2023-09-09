// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package render

import (
	htmltpl "html/template"
	"io"
	texttpl "text/template"
)

type Renderable interface {
	Name() string
	Execute(wr io.Writer, data any) error
}

func LoadHTMLTemplates() ([]Renderable, error) {
	tpl, err := htmltpl.New("kubernetes-apis").Funcs(templateFuncs).ParseGlob("templates/*")
	if err != nil {
		return nil, err
	}

	result := []Renderable{}
	for _, t := range tpl.Templates() {
		result = append(result, t.Lookup(t.Name()))
	}

	return result, nil
}

func LoadTextTemplates() ([]Renderable, error) {
	tpl, err := texttpl.New("kubernetes-apis").Funcs(templateFuncs).ParseGlob("templates/*")
	if err != nil {
		return nil, err
	}

	result := []Renderable{}
	for _, t := range tpl.Templates() {
		result = append(result, t.Lookup(t.Name()))
	}

	return result, nil
}
