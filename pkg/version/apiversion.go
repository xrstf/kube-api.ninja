package version

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var apiVersionRegex = regexp.MustCompile(`^(v[0-9]+)([a-z]+[0-9]+)?$`)

type APIVersion struct {
	version string
	suffix  string
}

func ParseAPIVersion(s string) (*APIVersion, error) {
	match := apiVersionRegex.FindStringSubmatch(s)
	if match == nil {
		return nil, errors.New("not a valid API version")
	}

	return &APIVersion{
		version: match[1],
		suffix:  match[2],
	}, nil
}

func (s *APIVersion) String() string {
	return fmt.Sprintf("%s%s", s.version, s.suffix)
}

func (s *APIVersion) Stable() bool {
	return s.suffix == ""
}

func (s *APIVersion) Maturity() string {
	if s.suffix == "" {
		return ""
	}

	match := suffixRegex.FindStringSubmatch(s.suffix)
	if match == nil {
		panic(fmt.Sprintf("%q is not a valid APIVersion suffix", s.suffix))
	}

	return match[1]
}

func (s *APIVersion) Prerelease() bool {
	return !s.Stable()
}

func (s *APIVersion) LessThan(other *APIVersion) bool {
	// comparing v1beta1 vs v2 => ignore suffix
	if s.version != other.version {
		return comparePrefixedNumbers(s.version, other.version)
	}

	if s.suffix != other.suffix {
		if other.suffix == "" { // comparing v1beta1 vs v1
			return true
		}

		if s.suffix == "" { // comparing v1 vs v1beta1
			return false
		}

		return comparePrefixedNumbers(s.suffix, other.suffix)
	}

	return false // versions are identical
}

var suffixRegex = regexp.MustCompile(`^([a-z]+)([0-9]+)$`)

// comparePrefixedNumbers returns true if i < j (e.g. "beta3" < "beta10" or "v2" < "v4")
func comparePrefixedNumbers(i, j string) bool {
	imatch := suffixRegex.FindStringSubmatch(i)
	if imatch == nil {
		panic(fmt.Sprintf("%q is not a valid prefixed number", i))
	}

	jmatch := suffixRegex.FindStringSubmatch(j)
	if jmatch == nil {
		panic(fmt.Sprintf("%q is not a valid prefixed number", j))
	}

	// same prefix level (could be "v2" vs. "v3")
	if imatch[1] == jmatch[1] {
		irelease, _ := strconv.Atoi(imatch[2])
		jrelease, _ := strconv.Atoi(jmatch[2])

		return irelease < jrelease
	}

	// compare beta vs. alpha
	return imatch[1] < jmatch[1]
}

// CompareAPIVersions is a helper to make slice.Sort() functions easier.
func CompareAPIVersions(i, j string) bool {
	iVersion, err := ParseAPIVersion(i)
	if err != nil {
		panic(fmt.Sprintf("%q is not a valid API version: %v", i, err))
	}

	jVersion, err := ParseAPIVersion(j)
	if err != nil {
		panic(fmt.Sprintf("%q is not a valid API version: %v", j, err))
	}

	return iVersion.LessThan(jVersion)
}

// PreferredAPIVersion tries to mimic Kubernetes' own mechanism to determine the
// preferred API version of an API.
func PreferredAPIVersion(versions []string) (*APIVersion, error) {
	var preferred *APIVersion

	for _, ver := range versions {
		apiVersion, err := ParseAPIVersion(ver)
		if err != nil {
			return nil, fmt.Errorf("%q is not a valid API version: %w", ver, err)
		}

		if preferred == nil {
			preferred = apiVersion
			continue
		}

		// same major version (e.g. "v1beta1" vs. "v1alpha3")
		if apiVersion.version == preferred.version {
			if apiVersion.Stable() { // e.g. "v1beta1" vs "v1"
				preferred = apiVersion
			} else if preferred.Stable() { // e.g. "v1" vs "v1beta1"
				// NOP
			} else if comparePrefixedNumbers(preferred.suffix, apiVersion.suffix) { // e.g. "v1alpha7" vs "v1beta1"
				preferred = apiVersion
			}

			continue
		}

		// different major versions

		// ... but same suffix (e.g. "v3" vs. "v6" or "v3beta1" vs. "v6beta1")
		if apiVersion.suffix == preferred.suffix {
			// highest major version wins
			if comparePrefixedNumbers(preferred.version, apiVersion.version) {
				preferred = apiVersion
			}

			continue
		}

		// everything's different, e.g. "v2beta2" vs. "v5alpha7"

		// In this somewhat unrealistic case, the highest maturity
		// wins, regardless of major version (e.g. "v1beta1" beats "v2alpha7").
		if apiVersion.Stable() {
			preferred = apiVersion
			continue
		}

		if preferred.Stable() {
			// NOP
			continue
		}

		// last resort, compare maturities (this relies on "alpha" being lexic. smaller than "beta")
		if apiVersion.Maturity() > preferred.Maturity() {
			preferred = apiVersion
			continue
		}

		// NOP
	}

	return preferred, nil
}
