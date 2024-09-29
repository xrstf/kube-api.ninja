package version

import (
	"fmt"
	"slices"

	kversion "k8s.io/apimachinery/pkg/util/version"
)

type Semver struct {
	v *kversion.Version
}

func ParseSemver(s string) (*Semver, error) {
	parsed, err := kversion.ParseSemantic(s)
	if err != nil {
		return nil, err
	}

	return &Semver{
		v: parsed,
	}, nil
}

func (s *Semver) String() string {
	return s.v.String()
}

func (s *Semver) LessThan(other *Semver) bool {
	return s.v.LessThan(other.v)
}

func (s *Semver) MajorMinor() string {
	return fmt.Sprintf("%d.%d", s.v.Major(), s.v.Minor())
}

func SortReleases(releases []string) []string {
	releases = slices.Clone(releases)
	slices.SortFunc(releases, func(a, b string) int {
		compared, err := kversion.MustParseGeneric(a).Compare(b)
		if err != nil {
			panic(err)
		}

		return compared
	})

	return releases
}

func Sort(versions []string) []string {
	versions = slices.Clone(versions)
	slices.SortFunc(versions, func(a, b string) int {
		aVersion, err := ParseSemver(a)
		if err != nil {
			panic(err)
		}

		compared, err := aVersion.v.Compare(b)
		if err != nil {
			panic(err)
		}

		return compared
	})

	return versions
}
