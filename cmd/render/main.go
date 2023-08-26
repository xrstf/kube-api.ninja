// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"html/template"
	"log"
	"os"
	"slices"

	"go.xrstf.de/kubernetes-apis/pkg/render/html"
	"go.xrstf.de/kubernetes-apis/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

func main() {
	db, err := loadDatabase("data/database.json")
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	tpl, err := template.New("kubernetes-apis").Funcs(template.FuncMap{
		"sliceContains": slices.Contains[[]string],
		"getPreferredVersion": func(apiGroup *types.GroupOverview, release string) string {
			return apiGroup.PreferredVersions[release]
		},
		"isPreferredVersion": func(apiGroup *types.GroupOverview, apiVersion *types.VersionOverview, release string) bool {
			return apiGroup.PreferredVersions[release] == apiVersion.Version
		},
	}).Parse(html.Index)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	allReleases := sets.List(sets.KeySet(db.APIGroups[0].PreferredVersions))

	tpl.Execute(os.Stdout, html.IndexData{
		Database: db,
		Releases: allReleases,
	})
}

func loadDatabase(filename string) (*types.APIOverview, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)

	rel := &types.APIOverview{}
	if err := decoder.Decode(rel); err != nil {
		return nil, err
	}

	return rel, nil
}
