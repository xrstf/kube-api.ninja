// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"go.xrstf.de/kubernetes-apis/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

func main() {
	releaseFiles, err := filepath.Glob("data/release-*.json")
	if err != nil {
		log.Fatalf("Failed to find release files: %v", err)
	}

	overview := types.APIOverview{}

	for _, releaseFile := range releaseFiles {
		releaseInfo, err := loadReleaseFile(releaseFile)
		if err != nil {
			log.Fatalf("Failed to load release file %q: %v", releaseFile, err)
		}

		mergeReleaseIntoOverview(&overview, releaseInfo)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(overview); err != nil {
		log.Fatalf("Failed to JSON encode overview: %v", err)
	}
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

func mergeReleaseIntoOverview(overview *types.APIOverview, release *types.KubernetesRelease) {
	// a cluster without any APIs
	if len(release.APIGroups) == 0 {
		return
	}

	if overview.APIGroups == nil {
		overview.APIGroups = []types.GroupOverview{}
	}

	for _, apiGroup := range release.APIGroups {
		// find a possibly pre-existing group info from a previous release
		var existingGroupOverview *types.GroupOverview
		for j, g := range overview.APIGroups {
			if apiGroup.Name == g.Name {
				existingGroupOverview = &overview.APIGroups[j]
				break
			}
		}

		// create a new entry and set the pointer to it
		if existingGroupOverview == nil {
			overview.APIGroups = append(overview.APIGroups, types.GroupOverview{})
			existingGroupOverview = &overview.APIGroups[len(overview.APIGroups)-1]
		}

		mergeAPIGroupOverviews(existingGroupOverview, &apiGroup, release.Release)
	}
}

func mergeAPIGroupOverviews(dest *types.GroupOverview, groupinfo *types.APIGroup, release string) {
	// copy the name
	dest.Name = groupinfo.Name

	// remember the preferred version of this group for this release
	if dest.PreferredVersions == nil {
		dest.PreferredVersions = map[string]string{}
	}
	dest.PreferredVersions[release] = groupinfo.PreferredVersion

	// a group without any versions
	if len(groupinfo.APIVersions) == 0 {
		return
	}

	if dest.APIVersions == nil {
		dest.APIVersions = []types.VersionOverview{}
	}

	for _, apiVersion := range groupinfo.APIVersions {
		// find a possibly pre-existing version info from a previous release
		var existingVersionOverview *types.VersionOverview
		for j, v := range dest.APIVersions {
			if apiVersion.Version == v.Version {
				existingVersionOverview = &dest.APIVersions[j]
				break
			}
		}

		// create a new entry and set the pointer to it
		if existingVersionOverview == nil {
			dest.APIVersions = append(dest.APIVersions, types.VersionOverview{})
			existingVersionOverview = &dest.APIVersions[len(dest.APIVersions)-1]
		}

		mergeAPIVersionOverviews(existingVersionOverview, &apiVersion, release)
	}
}

func mergeAPIVersionOverviews(dest *types.VersionOverview, versioninfo *types.APIVersion, release string) {
	// copy the version
	dest.Version = versioninfo.Version

	// a version without any resources
	if len(versioninfo.Resources) == 0 {
		return
	}

	if dest.Resources == nil {
		dest.Resources = []types.ResourceOverview{}
	}

	for _, resource := range versioninfo.Resources {
		// find a possibly pre-existing resource info from a previous release
		var existingResourceOverview *types.ResourceOverview
		for j, r := range dest.Resources {
			if resource.Kind == r.Kind {
				existingResourceOverview = &dest.Resources[j]
				break
			}
		}

		// create a new entry and set the pointer to it
		if existingResourceOverview == nil {
			dest.Resources = append(dest.Resources, types.ResourceOverview{})
			existingResourceOverview = &dest.Resources[len(dest.Resources)-1]
		}

		mergeAPIResourceOverviews(existingResourceOverview, &resource, release)
	}
}

func mergeAPIResourceOverviews(dest *types.ResourceOverview, resourceinfo *types.Resource, release string) {
	// copy the version
	dest.Kind = resourceinfo.Kind
	dest.Plural = resourceinfo.Plural
	dest.Singular = resourceinfo.Singular

	// remember the scope, which _could_ technically change between versions and/or releases
	if dest.Scopes == nil {
		dest.Scopes = map[string]string{}
	}

	if resourceinfo.Namespaced {
		dest.Scopes[release] = "Namespaced"
	} else {
		dest.Scopes[release] = "Cluster"
	}

	dest.Releases = sets.List(sets.New(dest.Releases...).Insert(release))
}
