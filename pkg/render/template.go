// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package render

import (
	"html/template"
)

func LoadTemplates() (*template.Template, error) {
	return template.New("kubernetes-apis").Funcs(templateFuncs).ParseGlob("templates/*")
}
