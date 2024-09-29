// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package timeline

import (
	"fmt"
	"sort"
	"time"

	"go.xrstf.de/kube-api.ninja/pkg/database"
	"go.xrstf.de/kube-api.ninja/pkg/types"
	"go.xrstf.de/kube-api.ninja/pkg/version"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// The number of most recent releases we consider to show by default,
	// all older releases are "archived"; this is 11 because we want to
	// show e.g. 1.19..1.29, just because I think it looks nice.
	numRecentReleases = 11
)

func CreateTimeline(releases []*database.KubernetesRelease, now time.Time) (*Timeline, error) {
	timeline := &Timeline{
		Releases: []ReleaseMetadata{},
	}

	// sort releases to keep things consistent
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Semver().LessThan(releases[j].Semver())
	})

	// merge all releases together
	for _, release := range releases {
		// data is copied into the overview, so it's okay to have the loop re-use the same variable
		if err := mergeReleaseIntoOverview(timeline, release, now); err != nil {
			return nil, fmt.Errorf("failed to process release %s: %w", release.Version(), err)
		}
	}

	// mark old releases as archived
	if err := calculateArchivalStatus(timeline); err != nil {
		return nil, fmt.Errorf("failed to calculate archival status: %w", err)
	}

	// calculate "releases of interest":
	//   a) an API resource was added
	//   b) an API resource disappears
	//   c) a more mature version of an API group becomes available
	if err := calculateReleasesOfInterest(timeline); err != nil {
		return nil, fmt.Errorf("failed to calculate ROIs: %w", err)
	}

	// calculate latest available docs for each resource
	if err := calculateDocumentationReleases(timeline); err != nil {
		return nil, fmt.Errorf("failed to calculate documentation releases: %w", err)
	}

	// sort API groups alphabetically
	sort.Slice(timeline.APIGroups, func(i, j int) bool {
		return timeline.APIGroups[i].Name < timeline.APIGroups[j].Name
	})

	// sort versions for each API group in descending order (latest first)
	for idx, apiGroup := range timeline.APIGroups {
		sort.Slice(apiGroup.APIVersions, func(i, j int) bool {
			return version.CompareAPIVersions(apiGroup.APIVersions[j].Version, apiGroup.APIVersions[i].Version)
		})

		apiGroup.ReleasesOfInterest = version.SortReleases(apiGroup.ReleasesOfInterest)

		timeline.APIGroups[idx] = apiGroup
	}

	return timeline, nil
}

func mergeReleaseIntoOverview(timeline *Timeline, release *database.KubernetesRelease, now time.Time) error {
	api, err := release.API()
	if err != nil {
		return fmt.Errorf("failed to load API: %w", err)
	}

	metadata, err := createReleaseMetadata(release, now)
	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	timeline.Releases = append(timeline.Releases, metadata)

	// a cluster without any APIs
	if len(api.APIGroups) == 0 {
		return nil
	}

	if timeline.APIGroups == nil {
		timeline.APIGroups = []APIGroup{}
	}

	for _, apiGroup := range api.APIGroups {
		apiGroupName := apiGroup.Name
		if apiGroupName == "" {
			apiGroupName = "core"
		}

		// find a possibly pre-existing group info from a previous release
		var existingGroup *APIGroup
		for j, g := range timeline.APIGroups {
			if apiGroupName == g.Name {
				existingGroup = &timeline.APIGroups[j]
				break
			}
		}

		// create a new entry and set the pointer to it
		if existingGroup == nil {
			timeline.APIGroups = append(timeline.APIGroups, APIGroup{})
			existingGroup = &timeline.APIGroups[len(timeline.APIGroups)-1]
		}

		if err := mergeAPIGroupOverviews(existingGroup, &apiGroup, apiGroupName, release.Version()); err != nil {
			return fmt.Errorf("failed to process API group %s: %w", apiGroupName, err)
		}
	}

	return nil
}

func mergeAPIGroupOverviews(dest *APIGroup, groupinfo *types.APIGroup, groupName string, release string) error {
	// copy the name
	dest.Name = groupName

	// remember the preferred version of this group for this release
	if dest.PreferredVersions == nil {
		dest.PreferredVersions = map[string]string{}
	}
	dest.PreferredVersions[release] = groupinfo.PreferredVersion

	// a group without any versions
	if len(groupinfo.APIVersions) == 0 {
		return nil
	}

	if dest.APIVersions == nil {
		dest.APIVersions = []APIVersion{}
	}

	for _, apiVersion := range groupinfo.APIVersions {
		// find a possibly pre-existing version info from a previous release
		var existingVersion *APIVersion
		for j, v := range dest.APIVersions {
			if apiVersion.Version == v.Version {
				existingVersion = &dest.APIVersions[j]
				break
			}
		}

		// create a new entry and set the pointer to it
		if existingVersion == nil {
			dest.APIVersions = append(dest.APIVersions, APIVersion{})
			existingVersion = &dest.APIVersions[len(dest.APIVersions)-1]
		}

		if err := mergeAPIVersionOverviews(existingVersion, &apiVersion, release); err != nil {
			return fmt.Errorf("failed to process API version %s: %w", apiVersion.Version, err)
		}
	}

	// version-sort the .Releases on each created APIResource
	for i, apiVersion := range dest.APIVersions {
		apiVersion.Releases = version.SortReleases(apiVersion.Releases)
		apiVersion.ReleasesOfInterest = version.SortReleases(apiVersion.ReleasesOfInterest)

		dest.APIVersions[i] = apiVersion

		for j, apiResource := range apiVersion.Resources {
			apiResource.Releases = version.SortReleases(apiResource.Releases)
			apiResource.ReleasesOfInterest = version.SortReleases(apiResource.ReleasesOfInterest)

			dest.APIVersions[i].Resources[j] = apiResource
		}
	}

	return nil
}

func mergeAPIVersionOverviews(dest *APIVersion, versioninfo *types.APIVersion, release string) error {
	// copy the version
	dest.Version = versioninfo.Version
	dest.Releases = append(dest.Releases, release)

	// a version without any resources
	if len(versioninfo.Resources) == 0 {
		return nil
	}

	if dest.Resources == nil {
		dest.Resources = []APIResource{}
	}

	for _, resource := range versioninfo.Resources {
		// find a possibly pre-existing resource info from a previous release
		var existingResource *APIResource
		for j, r := range dest.Resources {
			if resource.Kind == r.Kind {
				existingResource = &dest.Resources[j]
				break
			}
		}

		// create a new entry and set the pointer to it
		if existingResource == nil {
			dest.Resources = append(dest.Resources, APIResource{})
			existingResource = &dest.Resources[len(dest.Resources)-1]
		}

		if err := mergeAPIResourceOverviews(existingResource, &resource, release); err != nil {
			return fmt.Errorf("failed to process API resource %s: %w", resource.Kind, err)
		}
	}

	return nil
}

func mergeAPIResourceOverviews(dest *APIResource, resourceinfo *types.Resource, release string) error {
	// copy the version
	dest.Kind = resourceinfo.Kind
	dest.Plural = resourceinfo.Plural
	dest.Singular = resourceinfo.Singular
	dest.Description = resourceinfo.Description
	dest.Releases = sets.New(dest.Releases...).Insert(release).UnsortedList()

	// remember the scope, which _could_ technically change between versions and/or releases
	if dest.Scopes == nil {
		dest.Scopes = map[string]string{}
	}

	if resourceinfo.Namespaced {
		dest.Scopes[release] = "Namespaced"
	} else {
		dest.Scopes[release] = "Cluster"
	}

	return nil
}

func createReleaseMetadata(release *database.KubernetesRelease, now time.Time) (ReleaseMetadata, error) {
	endOfLife, err := release.EndOfLifeDate()
	if err != nil {
		return ReleaseMetadata{}, fmt.Errorf("failed to read EOL date: %w", err)
	}

	releaseDate, err := release.ReleaseDate()
	if err != nil {
		return ReleaseMetadata{}, fmt.Errorf("failed to read release date: %w", err)
	}

	latestVersion, err := release.LatestVersion()
	if err != nil {
		return ReleaseMetadata{}, err
	}

	eol := endOfLife != nil && now.After(*endOfLife)

	// "!before" is not the same as "after"; on the release
	// date itself, it should be marked as supported
	released := !now.Before(releaseDate)
	supported := released && !eol

	return ReleaseMetadata{
		Version:       release.Version(),
		Released:      released,
		Supported:     supported,
		ReleaseDate:   releaseDate,
		EndOfLifeDate: endOfLife,
		LatestVersion: latestVersion,
		HasDocs:       release.HasDocumentation(),
	}, nil
}

func calculateReleasesOfInterest(tl *Timeline) error {
	for i, apiGroup := range tl.APIGroups {
		groupSuperset := sets.Set[string]{}

		for j, apiVersion := range apiGroup.APIVersions {
			versionSuperset := sets.Set[string]{}

			for k, apiResource := range apiVersion.Resources {
				notableReleases := getReleasesWithNotableChangesForResource(apiResource, tl.Releases)
				if len(notableReleases) > 0 {
					tl.APIGroups[i].APIVersions[j].Resources[k].ReleasesOfInterest = notableReleases
					versionSuperset.Insert(notableReleases...)
					// fmt.Printf("%s.%s.%s changes in %v\n", apiGroup.Name, apiVersion.Version, apiResource.Kind, notableReleases)
				}
			}

			if versionSuperset.Len() > 0 {
				tl.APIGroups[i].APIVersions[j].ReleasesOfInterest = sets.List(versionSuperset)
				groupSuperset = groupSuperset.Union(versionSuperset)
				// fmt.Printf("%s.%s changes in %v\n", apiGroup.Name, apiVersion.Version, sets.List(versionSuperset))
			}
		}

		if groupSuperset.Len() > 0 {
			tl.APIGroups[i].ReleasesOfInterest = sets.List(groupSuperset)
			// fmt.Printf("%s changes in %v\n", apiGroup.Name, sets.List(groupSuperset))
		}
	}

	return nil
}

func getReleasesWithNotableChangesForResource(res APIResource, sortedReleases []ReleaseMetadata) []string {
	availableInReleases := sets.New(res.Releases...)
	result := []string{}

	var wasAvailable bool
	for i, release := range sortedReleases {
		// for the first known release, we cannot determine if
		// there are breaking changes; this makes the loop quite neat
		if i > 0 {
			isAvailable := availableInReleases.Has(release.Version)

			if wasAvailable != isAvailable {
				result = append(result, release.Version)
			}
		}

		wasAvailable = availableInReleases.Has(release.Version)
	}

	return result
}

func calculateDocumentationReleases(tl *Timeline) error {
	hasDocsInRelease := func(version string) bool {
		for _, release := range tl.Releases {
			if release.Version == version {
				return release.HasDocs
			}
		}

		return false
	}

	for i, apiGroup := range tl.APIGroups {
		for j, apiVersion := range apiGroup.APIVersions {
			for k, apiResource := range apiVersion.Resources {
				availableIn := apiResource.Releases
				if len(availableIn) == 0 {
					continue // should never happen
				}

				latestReleaseWithDocs := ""

				// walk backwards to prefer later releases
				for i := len(availableIn) - 1; i >= 0; i-- {
					if version := availableIn[i]; hasDocsInRelease(version) {
						latestReleaseWithDocs = version
						break
					}
				}

				if latestReleaseWithDocs != "" {
					tl.APIGroups[i].APIVersions[j].Resources[k].DocRelease = latestReleaseWithDocs
				}
			}
		}
	}

	return nil
}

func calculateArchivalStatus(tl *Timeline) error {
	totalReleases := len(tl.Releases)
	archiveThresold := totalReleases - numRecentReleases

	// mark releases as archived
	archivedRelases := sets.Set[string]{}
	for i, rel := range tl.Releases {
		if i < archiveThresold {
			tl.Releases[i].Archived = true
			archivedRelases.Insert(rel.Version)
		}
	}

	// based on the list of archived releases, mark resources/versions/groups
	// as archived if they only show up in archived releases
	for i, apiGroup := range tl.APIGroups {
		groupArchived := true

		for j, apiVersion := range apiGroup.APIVersions {
			versionArchived := true

			for k, apiResource := range apiVersion.Resources {
				// check if the release only appears in archived releases
				present := sets.New(apiResource.Releases...)

				if present.Difference(archivedRelases).Len() == 0 {
					tl.APIGroups[i].APIVersions[j].Resources[k].Archived = true
				} else {
					versionArchived = false
				}
			}

			tl.APIGroups[i].APIVersions[j].Archived = versionArchived

			if !versionArchived {
				groupArchived = false
			}
		}

		tl.APIGroups[i].Archived = groupArchived
	}

	return nil
}
