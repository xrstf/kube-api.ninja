// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package timeline

import "time"

type Timeline struct {
	APIGroups []APIGroup
	Releases  []ReleaseMetadata
}

type ReleaseMetadata struct {
	Version       string
	Supported     bool
	ReleaseDate   time.Time
	EndOfLifeDate *time.Time
	LatestVersion string
}

func (o *Timeline) ReleaseMetadata(release string) ReleaseMetadata {
	for _, r := range o.Releases {
		if r.Version == release {
			return r
		}
	}

	return ReleaseMetadata{}
}

func (o *Timeline) HasRelease(release string) bool {
	return o.ReleaseMetadata(release).Version != ""
}

type APIGroup struct {
	Name              string
	PreferredVersions map[string]string // lists the prefered version per release
	APIVersions       []APIVersion
}

// helper functions for templating :grin:

func (o *APIGroup) PreferredVersion(release string) string {
	return o.PreferredVersions[release]
}

type APIVersion struct {
	Version   string // e.g. "v1beta1"
	Releases  []string
	Resources []APIResource
}

func (o *APIVersion) HasRelease(release string) bool {
	for _, r := range o.Releases {
		if r == release {
			return true
		}
	}

	return false
}

type APIResource struct {
	Kind     string
	Singular string
	Plural   string
	Scopes   map[string]string
	Releases []string
}

func (o *APIResource) HasRelease(release string) bool {
	for _, r := range o.Releases {
		if r == release {
			return true
		}
	}

	return false
}
