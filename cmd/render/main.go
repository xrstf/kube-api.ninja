// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.xrstf.de/kubernetes-apis/pkg/database"
	"go.xrstf.de/kubernetes-apis/pkg/timeline"
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

	templates, err := template.New("kubernetes-apis").Funcs(templateFuncs).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		log.Fatalf("Failed to create %s directory: %v", outputDirectory, err)
	}

	log.Println("Rendering index.html…")
	if err := renderIndex(outputDirectory, templates, timelineObj); err != nil {
		log.Fatalf("Failed to render index.html: %v", err)
	}

	log.Println("Done.")
}

type indexData struct {
	Timeline *timeline.Timeline
}

func renderIndex(directory string, tpl *template.Template, timelineObj *timeline.Timeline) error {
	f, err := os.Create(filepath.Join(directory, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tpl.ExecuteTemplate(f, "index.html", indexData{
		Timeline: timelineObj,
	})
}
