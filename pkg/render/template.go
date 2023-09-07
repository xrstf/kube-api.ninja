// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package render

import (
	htmltpl "html/template"
	texttpl "text/template"
)

func LoadHTMLTemplates() (*htmltpl.Template, error) {
	return htmltpl.New("kubernetes-apis").Funcs(templateFuncs).ParseGlob("templates/*")
}

func LoadTextTemplates() (*texttpl.Template, error) {
	return texttpl.New("kubernetes-apis").Funcs(templateFuncs).ParseGlob("templates/*")
}
