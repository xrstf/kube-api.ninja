// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package types

import "sort"

type KubernetesAPI struct {
	Version string `json:"version"`
	Release string `json:"release"`

	APIGroups []APIGroup `json:"apiGroups"`
}

func (r *KubernetesAPI) Sort() {
	for k, group := range r.APIGroups {
		group.Sort()
		r.APIGroups[k] = group
	}

	sort.Slice(r.APIGroups, func(i, j int) bool {
		return r.APIGroups[i].Name < r.APIGroups[j].Name
	})
}

type APIGroup struct {
	Name             string       `json:"name"` // e.g. "apps"
	PreferredVersion string       `json:"preferredVersion"`
	APIVersions      []APIVersion `json:"apiVersions"`
}

func (g *APIGroup) Sort() {
	for k, version := range g.APIVersions {
		version.Sort()
		g.APIVersions[k] = version
	}

	sort.Slice(g.APIVersions, func(i, j int) bool {
		return g.APIVersions[i].Version < g.APIVersions[j].Version
	})
}

type APIVersion struct {
	Version   string     `json:"version"` // e.g. "v1beta1"
	Resources []Resource `json:"resources"`
}

func (v *APIVersion) Sort() {
	sort.Slice(v.Resources, func(i, j int) bool {
		return v.Resources[i].Kind < v.Resources[j].Kind
	})
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
