// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	htmltpl "html/template"
	"log"
	"os"
	"path/filepath"
	"text/template"
	texttpl "text/template"
	"time"

	"go.xrstf.de/kube-api.ninja/pkg/database"
	"go.xrstf.de/kube-api.ninja/pkg/render"
	"go.xrstf.de/kube-api.ninja/pkg/timeline"
)

const (
	outputDirectory = "public"
)

func main() {
	now := time.Now().UTC()

	db, err := database.NewReleaseDatabase("data")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	releaseNames, err := db.Releases()
	if err != nil {
		log.Fatalf("Failed to list available releases: %v", err)
	}

	releases := []*database.KubernetesRelease{}
	for _, releaseName := range releaseNames {
		release, err := db.Release(releaseName)
		if err != nil {
			log.Fatalf("Failed to load release %q: %v", releaseName, err)
		}

		releases = append(releases, release)
	}

	timelineObj, err := timeline.CreateTimeline(releases, now)
	if err != nil {
		log.Fatalf("Failed to create timeline: %v", err)
	}

	htmlTemplates, err := render.LoadHTMLTemplates()
	if err != nil {
		log.Fatalf("Failed to parse HTML template: %v", err)
	}

	textTemplates, err := render.LoadTextTemplates()
	if err != nil {
		log.Fatalf("Failed to parse text template: %v", err)
	}

	for _, dir := range []string{
		filepath.Join(outputDirectory, "static", "css"),
		filepath.Join(outputDirectory, "static", "js"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create %s directory: %v", dir, err)
		}
	}

	log.Println("Rendering index.html…")
	if err := renderIndex(outputDirectory, htmlTemplates, timelineObj); err != nil {
		log.Fatalf("Failed to render index.html: %v", err)
	}

	log.Println("Rendering about.html…")
	if err := renderAbout(outputDirectory, htmlTemplates, timelineObj); err != nil {
		log.Fatalf("Failed to render about.html: %v", err)
	}

	log.Println("Rendering site.css…")
	if err := renderCSS(outputDirectory, textTemplates, timelineObj); err != nil {
		log.Fatalf("Failed to render site.css: %v", err)
	}

	log.Println("Rendering site.js")
	if err := renderJS(outputDirectory, textTemplates, timelineObj); err != nil {
		log.Fatalf("Failed to render site.js: %v", err)
	}

	log.Println("Done.")
}

type pageData struct {
	Timeline *timeline.Timeline
}

func renderIndex(directory string, tpl *htmltpl.Template, timelineObj *timeline.Timeline) error {
	f, err := os.Create(filepath.Join(directory, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tpl.ExecuteTemplate(f, "index.html", pageData{
		Timeline: timelineObj,
	})
}

func renderAbout(directory string, tpl *htmltpl.Template, timelineObj *timeline.Timeline) error {
	f, err := os.Create(filepath.Join(directory, "about.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tpl.ExecuteTemplate(f, "about.html", pageData{
		Timeline: timelineObj,
	})
}

func renderCSS(directory string, tpl *texttpl.Template, timelineObj *timeline.Timeline) error {
	f, err := os.Create(filepath.Join(directory, "static", "css", "site.css"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tpl.ExecuteTemplate(f, "site.css", pageData{
		Timeline: timelineObj,
	})
}

func renderJS(directory string, tpl *template.Template, timelineObj *timeline.Timeline) error {
	f, err := os.Create(filepath.Join(directory, "static", "js", "site.js"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tpl.ExecuteTemplate(f, "site.js", pageData{
		Timeline: timelineObj,
	})
}
