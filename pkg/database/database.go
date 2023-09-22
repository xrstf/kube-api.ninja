package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"k8s.io/apimachinery/pkg/util/version"
)

type ReleaseDatabase struct {
	baseDir string
	docsDir string
}

func NewReleaseDatabase(baseDir string, docsDir string) (*ReleaseDatabase, error) {
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare data directory: %w", err)
	}

	baseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to determine absolute path for data directory: %w", err)
	}

	return &ReleaseDatabase{
		baseDir: baseDir,
		docsDir: docsDir,
	}, nil
}

func (db *ReleaseDatabase) Releases() ([]string, error) {
	dirs, err := filepath.Glob(filepath.Join(db.baseDir, "releases", "*"))
	if err != nil {
		return nil, fmt.Errorf("failed to find release directories: %w", err)
	}

	releases := []string{}
	for _, dir := range dirs {
		releases = append(releases, filepath.Base(dir))
	}

	sort.Slice(releases, func(i, j int) bool {
		iRelease, err := version.ParseGeneric(releases[i])
		if err != nil {
			panic(err)
		}

		jRelease, err := version.ParseGeneric(releases[j])
		if err != nil {
			panic(err)
		}

		return iRelease.LessThan(jRelease)
	})

	return releases, nil
}

func (db *ReleaseDatabase) Release(version string) (*KubernetesRelease, error) {
	fullDir := filepath.Join(db.baseDir, "releases", version)

	if _, err := os.Stat(fullDir); err != nil {
		return nil, fmt.Errorf("failed to find release %q: %w", version, err)
	}

	docsDir := filepath.Join(db.docsDir, version)
	_, err := os.Stat(docsDir)
	hasDocs := err == nil

	return &KubernetesRelease{
		release: version,
		baseDir: fullDir,
		hasDocs: hasDocs,
	}, nil
}
