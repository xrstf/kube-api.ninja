// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"slices"

	"go.xrstf.de/kubernetes-apis/pkg/merger"
	"go.xrstf.de/kubernetes-apis/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	outputDirectory = "public"
)

var (
	templateFuncs = template.FuncMap{
		"sliceContains": slices.Contains[[]string],
		"getPreferredVersion": func(apiGroup *types.GroupOverview, release string) string {
			return apiGroup.PreferredVersions[release]
		},
		"isPreferredVersion": func(apiGroup *types.GroupOverview, apiVersion *types.VersionOverview, release string) bool {
			return apiGroup.PreferredVersions[release] == apiVersion.Version
		},
	}
)

func main() {
	releaseFiles, err := filepath.Glob("data/release-*.json")
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
	Database *types.APIOverview
	Releases []string
}

func renderIndex(directory string, tpl *template.Template, apiOverview *types.APIOverview) error {
	f, err := os.Create(filepath.Join(directory, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	allReleases := sets.List(sets.KeySet(apiOverview.APIGroups[0].PreferredVersions))

	return tpl.ExecuteTemplate(f, "index.html", indexData{
		Database: apiOverview,
		Releases: allReleases,
	})
}
