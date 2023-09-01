// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package types

type KubernetesRelease struct {
	Version string `json:"version"`
	Release string `json:"release"`

	APIGroups []APIGroup `json:"apiGroups"`
}

type APIGroup struct {
	Name             string       `json:"name"` // e.g. "apps"
	PreferredVersion string       `json:"preferredVersion"`
	APIVersions      []APIVersion `json:"apiVersions"`
}

type APIVersion struct {
	Version   string     `json:"version"` // e.g. "v1beta1"
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Kind       string `json:"kind"`
	Namespaced bool   `json:"namespaced"`
	Singular   string `json:"singular"`
	Plural     string `json:"plural"`
}

type APIOverview struct {
	APIGroups []GroupOverview
	Releases  []string
}

func (o *APIOverview) HasRelease(release string) bool {
	for _, r := range o.Releases {
		if r == release {
			return true
		}
	}

	return false
}

type GroupOverview struct {
	Name              string
	PreferredVersions map[string]string // lists the prefered version per release
	APIVersions       []VersionOverview
}

// helper functions for templating :grin:

func (o *GroupOverview) PreferredVersion(release string) string {
	return o.PreferredVersions[release]
}

type VersionOverview struct {
	Version   string // e.g. "v1beta1"
	Releases  []string
	Resources []ResourceOverview
}

func (o *VersionOverview) HasRelease(release string) bool {
	for _, r := range o.Releases {
		if r == release {
			return true
		}
	}

	return false
}

type ResourceOverview struct {
	Kind     string
	Singular string
	Plural   string
	Scopes   map[string]string
	Releases []string
}

func (o *ResourceOverview) HasRelease(release string) bool {
	for _, r := range o.Releases {
		if r == release {
			return true
		}
	}

	return false
}
