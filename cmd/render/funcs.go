// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"go.xrstf.de/kubernetes-apis/pkg/timeline"
	"go.xrstf.de/kubernetes-apis/pkg/version"
)

var (
	templateFuncs = template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"getReleaseHeaderClass":        getReleaseHeaderClass,
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

func getReleaseHeaderClassNames(tl *timeline.Timeline, release *timeline.ReleaseMetadata) []string {
	classes := []string{}

	if release.Supported {
		classes = append(classes, "release-supported")

		// is this the oldest supported release?
		isOldest := false

		for _, metadata := range tl.Releases {
			if metadata.Supported {
				isOldest = metadata.Version == release.Version
				break
			}
		}

		if isOldest {
			classes = append(classes, "oldest-release-supported")
		}

	} else {
		classes = append(classes, "release-unsupported")
	}

	return classes
}

func getReleaseHeaderClass(tl *timeline.Timeline, release *timeline.ReleaseMetadata) string {
	classes := append(getReleaseHeaderClassNames(tl, release), "release")

	return strings.Join(classes, " ")
}

func getAPIGroupReleaseClass(tl *timeline.Timeline, apiGroup *timeline.APIGroup, release *timeline.ReleaseMetadata) string {
	classes := append(getReleaseHeaderClassNames(tl, release), "release")

	if release.Supported {
		classes = append(classes, "supported")
	} else {
		classes = append(classes, "unsupported")
	}

	preferred := apiGroup.PreferredVersion(release.Version)
	if preferred == "" {
		classes = append(classes, "a10y-missing")
	} else {
		classes = append(classes, "a10y-exists")

		v, _ := version.ParseAPIVersion(preferred)
		if v.Prerelease() {
			classes = append(classes, "maturity-prerelease")
		} else {
			classes = append(classes, "maturity-stable")
		}

		// there is a preferred version for this API group in this Kubernetes release;
		// figure out if this is the first or last release to offer this API group at all
		// (this is maybe slow, we could also precalculate the availability first and then
		// just iterate over all groups and keep track, instead of this lookbehind/ahead)

		edge := false

		if prevRelease := jumpMinorRelease(release.Version, -1); tl.HasRelease(prevRelease) {
			if apiGroup.PreferredVersion(prevRelease) == "" {
				classes = append(classes, "a10y-begin")
				edge = true
			}
		}

		if nextRelease := jumpMinorRelease(release.Version, +1); tl.HasRelease(nextRelease) {
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

func getAPIVersionReleaseClass(tl *timeline.Timeline, apiGroup *timeline.APIGroup, apiVersion *timeline.APIVersion, release *timeline.ReleaseMetadata) string {
	classes := append(getReleaseHeaderClassNames(tl, release), "release")

	if release.Supported {
		classes = append(classes, "supported")
	} else {
		classes = append(classes, "unsupported")
	}

	if !apiVersion.HasRelease(release.Version) {
		classes = append(classes, "a10y-missing")
	} else {
		classes = append(classes, "a10y-exists")

		// is this version the preferred version in this release?

		if apiGroup.PreferredVersions[release.Version] == apiVersion.Version {
			classes = append(classes, "a10y-preferred")
		}

		// is this the first or last release this API version is available in?

		edge := false

		if prevRelease := jumpMinorRelease(release.Version, -1); tl.HasRelease(prevRelease) {
			if !apiVersion.HasRelease(prevRelease) {
				classes = append(classes, "a10y-begin")
				edge = true
			}
		}

		if nextRelease := jumpMinorRelease(release.Version, +1); tl.HasRelease(nextRelease) {
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

func getAPIVersionReleaseContent(tl *timeline.Timeline, apiGroup *timeline.APIGroup, apiVersion *timeline.APIVersion, release *timeline.ReleaseMetadata) template.HTML {
	if !apiVersion.HasRelease(release.Version) {
		return template.HTML("&nbsp;")
	}

	if apiGroup.PreferredVersions[release.Version] == apiVersion.Version {
		return "✪"
	}

	return template.HTML("&nbsp;")
}

func getAPIResourceReleaseClass(tl *timeline.Timeline, apiGroup *timeline.APIGroup, apiVersion *timeline.APIVersion, apiResource *timeline.APIResource, release *timeline.ReleaseMetadata) string {
	classes := append(getReleaseHeaderClassNames(tl, release), "release")

	if release.Supported {
		classes = append(classes, "supported")
	} else {
		classes = append(classes, "unsupported")
	}

	if !apiResource.HasRelease(release.Version) {
		classes = append(classes, "a10y-missing")
	} else {
		classes = append(classes, "a10y-exists", "scope-"+strings.ToLower(apiResource.Scopes[release.Version]))

		// is this version the preferred version in this release?

		if apiGroup.PreferredVersions[release.Version] == apiVersion.Version {
			classes = append(classes, "a10y-preferred")
		}

		// is this the first or last release this API version is available in?

		edge := false

		if prevRelease := jumpMinorRelease(release.Version, -1); tl.HasRelease(prevRelease) {
			if !apiResource.HasRelease(prevRelease) {
				classes = append(classes, "a10y-begin")
				edge = true
			}
		}

		if nextRelease := jumpMinorRelease(release.Version, +1); tl.HasRelease(nextRelease) {
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

func getAPIResourceReleaseContent(tl *timeline.Timeline, apiGroup *timeline.APIGroup, apiVersion *timeline.APIVersion, apiResource *timeline.APIResource, release *timeline.ReleaseMetadata) template.HTML {
	if !apiResource.HasRelease(release.Version) {
		return template.HTML("&nbsp;")
	}

	// TODO: Only show this per-cell if the scope of the resource actually changed during the lifetime of a single APIVersion
	// (which is extremely unlikely for upstream API groups).

	switch apiResource.Scopes[release.Version] {
	case "Namespaced":
		return "namespaced"
	case "Cluster":
		return "cluster-wide"
	default:
		return template.HTML("&nbsp;")
	}
}
