// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	stamp := os.Getenv("ASSET_STAMP")
	if stamp == "" {
		stamp = now.Format("2006-01-02-15-04-05")
	} else if len(stamp) > 10 {
		stamp = stamp[:10]
	}

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

	data := &pageData{
		Timeline:   timelineObj,
		AssetStamp: stamp,
	}

	if err := renderFileType(outputDirectory, htmlTemplates, data, "html"); err != nil {
		log.Fatalf("Failed to render: %v", err)
	}

	if err := renderFileType(filepath.Join(outputDirectory, "static", "css"), textTemplates, data, "css"); err != nil {
		log.Fatalf("Failed to render: %v", err)
	}

	if err := renderFileType(filepath.Join(outputDirectory, "static", "js"), textTemplates, data, "js"); err != nil {
		log.Fatalf("Failed to render: %v", err)
	}

	log.Println("Done.")
}

type pageData struct {
	Timeline    *timeline.Timeline
	AssetStamp  string
	CurrentPage string
}

func renderFileType(targetDir string, tpls []render.Renderable, data *pageData, filetype string) error {
	extension := fmt.Sprintf(".%s", filetype)

	for _, t := range tpls {
		basename := t.Name()
		if !strings.HasSuffix(basename, extension) {
			continue
		}

		// ignore partials/helpers
		if strings.HasPrefix(basename, "_") {
			continue
		}

		log.Printf("Rendering %s…", basename)
		f, err := os.Create(filepath.Join(targetDir, basename))
		if err != nil {
			return err
		}

		data.CurrentPage = basename

		if err := t.Execute(f, data); err != nil {
			f.Close()
			return fmt.Errorf("failed to render %s: %w", basename, err)
		}

		f.Close()
	}

	return nil
}
