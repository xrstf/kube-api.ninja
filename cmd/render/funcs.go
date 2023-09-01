// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"go.xrstf.de/kubernetes-apis/pkg/types"
)

var (
	templateFuncs = template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"getAPIGroupReleaseClass":      getAPIGroupReleaseClass,
		"getAPIVersionReleaseClass":    getAPIVersionReleaseClass,
		"getAPIVersionReleaseContent":  getAPIVersionReleaseContent,
		"getAPIResourceReleaseClass":   getAPIResourceReleaseClass,
		"getAPIResourceReleaseContent": getAPIResourceReleaseContent,
	}
)

func jumpMinorRelease(s string, minorSteps int) string {
	parts := strings.Split(s, ".")

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s.%d", parts[0], minor+minorSteps)
}

func getAPIGroupReleaseClass(apiOverview *types.APIOverview, apiGroup *types.GroupOverview, release string) string {
	classes := []string{"release"}

	preferred := apiGroup.PreferredVersion(release)
	if preferred == "" {
		classes = append(classes, "a10y-missing")
	} else {
		classes = append(classes, "a10y-exists")

		// there is a preferred version for this API group in this Kubernetes release;
		// figure out if this is the first or last release to offer this API group at all
		// (this is maybe slow, we could also precalculate the availability first and then
		// just iterate over all groups and keep track, instead of this lookbehind/ahead)

		edge := false

		if prevRelease := jumpMinorRelease(release, -1); apiOverview.HasRelease(prevRelease) {
			if apiGroup.PreferredVersion(prevRelease) == "" {
				classes = append(classes, "a10y-begin")
				edge = true
			}
		}

		if nextRelease := jumpMinorRelease(release, +1); apiOverview.HasRelease(nextRelease) {
			if apiGroup.PreferredVersion(nextRelease) == "" {
				classes = append(classes, "a10y-end")
				edge = true
			}
		}

		if !edge {
			// makes CSS easier
			classes = append(classes, "a10y-middle")
		}
	}

	return strings.Join(classes, " ")
}

func getAPIVersionReleaseClass(apiOverview *types.APIOverview, apiGroup *types.GroupOverview, apiVersion *types.VersionOverview, release string) string {
	classes := []string{"release"}

	if !apiVersion.HasRelease(release) {
		classes = append(classes, "a10y-missing")
	} else {
		classes = append(classes, "a10y-exists")

		// is this version the preferred version in this release?

		if apiGroup.PreferredVersions[release] == apiVersion.Version {
			classes = append(classes, "a10y-preferred")
		}

		// is this the first or last release this API version is available in?

		edge := false

		if prevRelease := jumpMinorRelease(release, -1); apiOverview.HasRelease(prevRelease) {
			if !apiVersion.HasRelease(prevRelease) {
				classes = append(classes, "a10y-begin")
				edge = true
			}
		}

		if nextRelease := jumpMinorRelease(release, +1); apiOverview.HasRelease(nextRelease) {
			if !apiVersion.HasRelease(nextRelease) {
				classes = append(classes, "a10y-end")
				edge = true
			}
		}

		if !edge {
			// makes CSS easier
			classes = append(classes, "a10y-middle")
		}
	}

	return strings.Join(classes, " ")
}

func getAPIVersionReleaseContent(apiOverview *types.APIOverview, apiGroup *types.GroupOverview, apiVersion *types.VersionOverview, release string) string {
	if !apiVersion.HasRelease(release) {
		return "–"
	}

	if apiGroup.PreferredVersions[release] == apiVersion.Version {
		return "preferred"
	}

	return "avalable"
}

func getAPIResourceReleaseClass(apiOverview *types.APIOverview, apiGroup *types.GroupOverview, apiVersion *types.VersionOverview, apiResource *types.ResourceOverview, release string) string {
	classes := []string{"release", "scope-" + strings.ToLower(apiResource.Scopes[release])}

	if !apiResource.HasRelease(release) {
		classes = append(classes, "a10y-missing")
	} else {
		classes = append(classes, "a10y-exists")

		// is this version the preferred version in this release?

		if apiGroup.PreferredVersions[release] == apiVersion.Version {
			classes = append(classes, "a10y-preferred")
		}

		// is this the first or last release this API version is available in?

		edge := false

		if prevRelease := jumpMinorRelease(release, -1); apiOverview.HasRelease(prevRelease) {
			if !apiResource.HasRelease(prevRelease) {
				classes = append(classes, "a10y-begin")
				edge = true
			}
		}

		if nextRelease := jumpMinorRelease(release, +1); apiOverview.HasRelease(nextRelease) {
			if !apiResource.HasRelease(nextRelease) {
				classes = append(classes, "a10y-end")
				edge = true
			}
		}

		if !edge {
			// makes CSS easier
			classes = append(classes, "a10y-middle")
		}
	}

	return strings.Join(classes, " ")
}

func getAPIResourceReleaseContent(apiOverview *types.APIOverview, apiGroup *types.GroupOverview, apiVersion *types.VersionOverview, apiResource *types.ResourceOverview, release string) string {
	if !apiResource.HasRelease(release) {
		return "–"
	}

	return apiResource.Scopes[release]
}
