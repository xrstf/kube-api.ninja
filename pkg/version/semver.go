package version

import (
	"fmt"

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
