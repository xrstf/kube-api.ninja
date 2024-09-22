/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apidocs

import (
	"html/template"
)

type tocItem struct {
	Level       int
	Title       template.HTML
	Link        string
	File        string
	SubSections []*tocItem
}

func (ti *tocItem) ToHTML() template.HTML {
	rendered, err := renderTemplate("toc-item.html", ti)
	if err != nil {
		panic(err)
	}

	return rendered
}

type tableOfContents struct {
	Title    string
	Sections []*tocItem
}
