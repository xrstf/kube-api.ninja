package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.xrstf.de/kube-api.ninja/pkg/types"
	"go.xrstf.de/kube-api.ninja/pkg/version"
)

type KubernetesRelease struct {
	release string
	baseDir string
	hasDocs bool
}

func (r *KubernetesRelease) Version() string {
	return r.release
}

func (r *KubernetesRelease) HasDocumentation() bool {
	return r.hasDocs
}

func (r *KubernetesRelease) Semver() *version.Semver {
	parsed, err := version.ParseSemver(fmt.Sprintf("v%s.0", r.release))
	if err != nil {
		panic(err)
	}

	return parsed
}

func (r *KubernetesRelease) API() (*types.KubernetesAPI, error) {
	f, err := os.Open(filepath.Join(r.baseDir, "api.json"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)

	rel := &types.KubernetesAPI{}
	if err := decoder.Decode(rel); err != nil {
		return nil, err
	}

	return rel, nil
}

func (r *KubernetesRelease) ReleaseDate() (time.Time, error) {
	return r.readTime("released.txt")
}

func (r *KubernetesRelease) EndOfLifeDate() (*time.Time, error) {
	// EOL files are optional (EOL dates are not known before a new release)
	data, err := r.readFile("eol.txt")
	if err != nil || len(data) == 0 {
		return nil, nil
	}

	t, err := r.readTime("eol.txt")
	if err != nil {
		return nil, err
	}

	return &t, err
}

func (r *KubernetesRelease) LatestVersion() (string, error) {
	return r.readFile("latest.txt")
}

func (r *KubernetesRelease) readFile(basename string) (string, error) {
	data, err := os.ReadFile(filepath.Join(r.baseDir, basename))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func (r *KubernetesRelease) readTime(basename string) (time.Time, error) {
	contents, err := r.readFile(basename)
	if err != nil {
		return time.Time{}, err
	}

	date, err := time.ParseInLocation("2006-01-02", contents, time.UTC)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}
