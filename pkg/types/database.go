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
	APIGroups []GroupOverview `json:"apiGroups"`
}

type GroupOverview struct {
	Name              string            `json:"name"`
	PreferredVersions map[string]string `json:"preferredVersions"` // lists the prefered version per release
	APIVersions       []VersionOverview `json:"apiVersions"`
}

type VersionOverview struct {
	Version   string             `json:"version"` // e.g. "v1beta1"
	Releases  []string           `json:"releases"`
	Resources []ResourceOverview `json:"resources"`
}

type ResourceOverview struct {
	Kind     string            `json:"kind"`
	Singular string            `json:"singular"`
	Plural   string            `json:"plural"`
	Scopes   map[string]string `json:"scopes"`
	Releases []string          `json:"releases"`
}
