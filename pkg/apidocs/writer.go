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
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.xrstf.de/kube-api.ninja/pkg/apidocs/api"
	"go.xrstf.de/kube-api.ninja/pkg/apidocs/types"
)

type htmlWriter struct {
	config *api.Config
	toc    tableOfContents

	buildDir string
	tmpFiles map[string]template.HTML

	// currentTOCItem is used to remember the current item between
	// calls to e.g. WriteResourceCategory() followed by WriteResource().
	currentTOCItem *tocItem
}

func newHTMLWriter(opts types.Options, config *api.Config, title string) (*htmlWriter, error) {
	writer := htmlWriter{
		config:   config,
		buildDir: opts.BuildDirectory,
		tmpFiles: map[string]template.HTML{},
		toc: tableOfContents{
			Title:    title,
			Sections: []*tocItem{},
		},
	}

	if err := os.MkdirAll(writer.buildDir, os.FileMode(0755)); err != nil {
		return nil, err
	}

	return &writer, nil
}

func (w *htmlWriter) WriteOverview() error {
	filename := "_overview.html"
	if err := w.writeStaticFile(filename, w.sectionHeading("API Overview")); err != nil {
		return err
	}

	item := tocItem{
		Level: 1,
		Title: "Overview",
		Link:  "api-overview",
		File:  filename,
	}
	w.toc.Sections = append(w.toc.Sections, &item)
	w.currentTOCItem = &item

	return nil
}

func (w *htmlWriter) WriteAPIGroupVersions(gvs api.GroupVersions) error {
	groups := api.ApiGroups{}
	for group := range gvs {
		groups = append(groups, api.ApiGroup(group))
	}
	sort.Sort(groups)

	tplGroups := []map[string]any{}

	for _, group := range groups {
		versionList := gvs[group.String()]
		sort.Sort(versionList)
		var versions []string
		for _, v := range versionList {
			versions = append(versions, v.String())
		}

		tplGroups = append(tplGroups, map[string]any{
			"group":    group,
			"versions": versions,
		})
	}

	filename := "_api_groups.html"
	if err := w.writeTemplate(filename, "api-groups.html", map[string]any{
		"groups": tplGroups,
	}); err != nil {
		return err
	}

	item := tocItem{
		Level: 1,
		Title: "API Groups",
		Link:  "api-groups",
		File:  filename,
	}
	w.toc.Sections = append(w.toc.Sections, &item)
	w.currentTOCItem = &item

	return nil
}

func (w *htmlWriter) WriteResourceCategory(name, file string) error {
	if err := w.writeStaticFile("_"+file+".html", w.resourceCategoryHeading(name)); err != nil {
		return err
	}

	link := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	item := tocItem{
		Level: 1,
		Title: template.HTML(name),
		Link:  link,
		File:  "_" + file + ".html",
	}
	w.toc.Sections = append(w.toc.Sections, &item)
	w.currentTOCItem = &item

	return nil
}

func (w *htmlWriter) resourceCategoryHeading(title string) template.HTML {
	rendered, err := renderTemplate("resource-category-heading.html", map[string]any{
		"title":     title,
		"sectionID": strings.ToLower(strings.ReplaceAll(title, " ", "-")),
	})
	if err != nil {
		panic(err)
	}

	return rendered
}

func (w *htmlWriter) sectionHeading(title string) template.HTML {
	rendered, err := renderTemplate("section-heading.html", title)
	if err != nil {
		panic(err)
	}

	return rendered
}

func (w *htmlWriter) WriteDefinitionsOverview() error {
	if err := w.writeStaticFile("_definitions.html", w.sectionHeading("Definitions")); err != nil {
		return err
	}

	item := tocItem{
		Level: 1,
		Title: "DEFINITIONS",
		Link:  "definitions",
		File:  "_definitions.html",
	}
	w.toc.Sections = append(w.toc.Sections, &item)
	w.currentTOCItem = &item

	return nil
}

func (w *htmlWriter) WriteOrphanedOperationsOverview() error {
	if err := w.writeStaticFile("_operations.html", w.sectionHeading("Operations")); err != nil {
		return err
	}

	item := tocItem{
		Level: 1,
		Title: "OPERATIONS",
		Link:  "operations",
		File:  "_operations.html",
	}
	w.toc.Sections = append(w.toc.Sections, &item)
	w.currentTOCItem = &item

	return nil
}

func (w *htmlWriter) WriteDefinition(d *api.Definition) error {
	filename := definitionFileName(d)
	nvg := fmt.Sprintf("%s %s %s", d.Name, d.Version, d.GroupDisplayName())
	linkID := getLink(nvg)
	title := gvkMarkup(d.GroupDisplayName(), d.Version, d.Name)

	// Definitions are added to the TOC to enable the generator to later collect
	// all the individual definition files, but definitions will not show up
	// in the nav treet because it would take up too much screen estate.
	item := tocItem{
		Level: 2,
		Title: title,
		Link:  linkID,
		File:  filename,
	}
	w.currentTOCItem.SubSections = append(w.currentTOCItem.SubSections, &item)

	return w.writeTemplate(filename, "definition.html", map[string]any{
		"nvg":        title,
		"linkID":     linkID,
		"definition": d,
	})
}

func (w *htmlWriter) WriteOperation(o *api.Operation) error {
	filename := operationFileName(o)
	nvg := o.ID
	linkID := getLink(nvg)

	oGroup, oVersion, oKind, _ := o.GetGroupVersionKindSub()
	oApiVersion := api.ApiVersion(oVersion)

	title := template.HTML(nvg)
	if len(oGroup) > 0 {
		title = gvkMarkup(oGroup, oApiVersion, oKind)
	}

	sort.Slice(o.HttpResponses, func(i, j int) bool {
		return strings.Compare(o.HttpResponses[i].Name, o.HttpResponses[j].Name) < 0
	})

	item := tocItem{
		Level: 2,
		Title: title,
		Link:  linkID,
		File:  filename,
	}
	w.currentTOCItem.SubSections = append(w.currentTOCItem.SubSections, &item)

	return w.writeTemplate(filename, "operation.html", map[string]any{
		"linkID":    linkID,
		"nvg":       nvg,
		"operation": o,
	})
}

func (w *htmlWriter) WriteResource(r *api.Resource) error {
	filename := conceptFileName(r.Definition)
	dvg := fmt.Sprintf("%s %s %s", r.Name, r.Definition.Version, r.Definition.GroupDisplayName())
	linkID := getLink(dvg)

	resourceItem := tocItem{
		Level: 2,
		Title: gvkMarkup(r.Definition.GroupDisplayName(), r.Definition.Version, r.Name),
		Link:  linkID,
		File:  filename,
	}
	w.currentTOCItem.SubSections = append(w.currentTOCItem.SubSections, &resourceItem)

	for _, oc := range r.Definition.OperationCategories {
		if len(oc.Operations) == 0 {
			continue
		}

		ocItem := tocItem{
			Level: 3,
			Title: template.HTML(oc.Name),
			Link:  oc.TocID(r.Definition),
		}
		resourceItem.SubSections = append(resourceItem.SubSections, &ocItem)

		for _, o := range oc.Operations {
			ocItem.SubSections = append(ocItem.SubSections, &tocItem{
				Level: 4,
				Title: template.HTML(o.Type.Name),
				Link:  o.TocID(r.Definition),
			})
		}
	}

	return w.writeTemplate(filename, "resource.html", map[string]any{
		"resource": r,
		"dvg":      resourceItem.Title,
		"linkID":   linkID,
	})
}

func (w *htmlWriter) WriteOldVersionsOverview() error {
	if err := w.writeStaticFile("_oldversions.html", w.sectionHeading("Old API Versions")); err != nil {
		return err
	}

	item := tocItem{
		Level: 1,
		Title: "OLD API VERSIONS",
		Link:  "old-api-versions",
		File:  "_oldversions.html",
	}
	w.toc.Sections = append(w.toc.Sections, &item)
	w.currentTOCItem = &item

	return nil
}

func (w *htmlWriter) WriteIndex() error {
	html, err := os.Create(filepath.Join(w.buildDir, "index.html"))
	if err != nil {
		return err
	}
	defer html.Close()

	// collect content from all the individual files we just created
	var content strings.Builder

	collect := func(filename string) {
		if fileContent, exists := w.tmpFiles[filename]; exists {
			content.WriteString(string(fileContent))
		} else {
			log.Printf("Collecting %s… \033[31mNot found\033[0m", filename)
		}
	}

	// TODO: Make this a recursive function.
	for _, sec := range w.toc.Sections {
		collect(sec.File)

		for _, sub := range sec.SubSections {
			if len(sub.File) > 0 {
				collect(sub.File)
			}

			for _, subsub := range sub.SubSections {
				if len(subsub.File) > 0 {
					collect(subsub.File)
				}
			}
		}
	}

	pos := strings.LastIndex(w.config.SpecVersion, ".")
	release := fmt.Sprintf("release-%s", w.config.SpecVersion[1:pos])
	specLink := "https://github.com/kubernetes/kubernetes/blob/" + release + "/api/openapi-spec/swagger.json"

	return renderTemplateTo(html, "index.html", map[string]any{
		"toc":      w.toc,
		"config":   w.config,
		"specLink": specLink,
		"content":  template.HTML(content.String()),
	})
}

func gvkMarkup(group string, version api.ApiVersion, kind string) template.HTML {
	return template.HTML(fmt.Sprintf(`<span class="gvk"><span class="k">%s</span> <span class="v">%s</span> <span class="g">%s</span></span>`, kind, version, group))
}

func generatedFileName(name string) string {
	return fmt.Sprintf("_generated_%s.html", strings.ToLower(strings.ReplaceAll(name, ".", "_")))
}

func definitionFileName(d *api.Definition) string {
	return generatedFileName(fmt.Sprintf("%s_%s_%s_definition", d.Name, d.Version, d.Group))
}

func operationFileName(o *api.Operation) string {
	return generatedFileName(fmt.Sprintf("%s_operation", o.ID))
}

func conceptFileName(d *api.Definition) string {
	return generatedFileName(fmt.Sprintf("%s_%s_%s_concept", d.Name, d.Version, d.Group))
}

func getLink(s string) string {
	tmp := strings.ReplaceAll(s, ".", "-")
	return strings.ToLower(strings.ReplaceAll(tmp, " ", "-"))
}

func (w *htmlWriter) writeTemplate(filename string, tplName string, data any) error {
	rendered, err := renderTemplate(tplName, data)
	if err != nil {
		return err
	}

	w.tmpFiles[filename] = rendered

	return nil
}

func (w *htmlWriter) writeStaticFile(filename string, defaultContent template.HTML) error {
	src := filepath.Join(w.config.SectionsDirectory, filename)

	// prefer a hand-crafted file if available
	if _, err := os.Stat(src); err == nil {
		content, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		defaultContent = template.HTML(content)
	}

	w.tmpFiles[filename] = defaultContent

	return nil
}
