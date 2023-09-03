// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"go.xrstf.de/kubernetes-apis/pkg/merger"
	"go.xrstf.de/kubernetes-apis/pkg/types"
)

const (
	outputDirectory = "public"
)

func main() {
	releaseFiles, err := filepath.Glob("data/swagger/release-*.json")
	if err != nil {
		log.Fatalf("Failed to find release files: %v", err)
	}

	releases := []types.KubernetesRelease{}
	for _, releaseFile := range releaseFiles {
		releaseInfo, err := loadReleaseFile(releaseFile)
		if err != nil {
			log.Fatalf("Failed to load release file %q: %v", releaseFile, err)
		}

		releases = append(releases, *releaseInfo)
	}

	overview := merger.MergeKubernetesReleases(releases)

	templates, err := template.New("kubernetes-apis").Funcs(templateFuncs).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		log.Fatalf("Failed to create %s directory: %v", outputDirectory, err)
	}

	log.Println("Rendering index.html…")
	if err := renderIndex(outputDirectory, templates, &overview); err != nil {
		log.Fatalf("Failed to render index.html: %v", err)
	}

	log.Println("Done.")
}

func loadReleaseFile(filename string) (*types.KubernetesRelease, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)

	rel := &types.KubernetesRelease{}
	if err := decoder.Decode(rel); err != nil {
		return nil, err
	}

	return rel, nil
}

type indexData struct {
	Overview *types.APIOverview
}

func renderIndex(directory string, tpl *template.Template, apiOverview *types.APIOverview) error {
	f, err := os.Create(filepath.Join(directory, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tpl.ExecuteTemplate(f, "index.html", indexData{
		Overview: apiOverview,
	})
}
