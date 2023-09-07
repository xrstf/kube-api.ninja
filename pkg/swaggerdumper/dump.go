// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package swaggerdumper

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"go.xrstf.de/kube-api.ninja/pkg/types"
	"go.xrstf.de/kube-api.ninja/pkg/version"

	"log/slog"
)

// define just enough of swagger's spec to parse what we need :)

type swaggerSpec struct {
	Definitions map[string]swaggerDefinitionSpec `json:"definitions"`
	Paths       map[string]swaggerPathSpec       `json:"paths"`
}

type swaggerDefinitionSpec struct {
	Description string `json:"description"`
}

// We deduce if a resource is namespaced based on the path (the key in this map).
type swaggerPathSpec struct {
	Get  *swaggerPathMethodSpec `json:"get"`
	Post *swaggerPathMethodSpec `json:"post"`
	// swaggerPathMethodSpec
}

type swaggerPathMethodSpec struct {
	KubernetesGVK swaggerGVK `json:"x-kubernetes-group-version-kind"`
	Responses     struct {
		OK struct {
			Schema struct {
				Ref string `json:"$ref"`
			} `json:"schema"`
		} `json:"200"`
	} `json:"responses"`
}

type swaggerGVK struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

var (
	coreApiVersionPath = regexp.MustCompile(`/api/(v.+)/$`)

	apisGroupPath    = regexp.MustCompile(`/apis/([^/]+)/$`)
	apisVersionPath  = regexp.MustCompile(`/apis/[^/]+/([^/]+)/$`)
	apisResourcePath = regexp.MustCompile(`/apis/[^/]+/[^/]+/([^/]+)$`)
)

func DumpSwaggerSpec(filename string, kubernetesVersion string) (*types.KubernetesAPI, error) {
	kubeVersion, err := version.ParseSemver(kubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid Kubernetes version: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Swagger spec: %w", err)
	}
	defer f.Close()

	result := &types.KubernetesAPI{
		Version:   kubeVersion.String(),
		Release:   kubeVersion.MajorMinor(),
		APIGroups: []types.APIGroup{},
	}

	spec := swaggerSpec{}
	if err := json.NewDecoder(f).Decode(&spec); err != nil {
		return nil, fmt.Errorf("failed to parse Swagger spec: %w", err)
	}

	// parse step by step and deduce data based on the request paths
	// step 1, find core versions
	result.APIGroups = append(result.APIGroups, dumpCoreAPIGroup(logger.With("group", ""), &spec))

	// step 2, find all other APIs
	for path := range spec.Paths {
		// there is exactly 1 path for the API group itself (plus many paths inside the API group)
		match := apisGroupPath.FindStringSubmatch(path)
		if match == nil {
			continue
		}

		result.APIGroups = append(result.APIGroups, dumpAPIGroup(logger.With("group", match[1]), &spec, match[1]))
	}

	// compute preferred version for each API group

	for i, group := range result.APIGroups {
		apiVersions := []string{}

		for _, v := range group.APIVersions {
			apiVersions = append(apiVersions, v.Version)
		}

		preferred, err := version.PreferredAPIVersion(apiVersions)
		if err != nil {
			panic(err)
		}

		result.APIGroups[i].PreferredVersion = preferred.String()
	}

	return result, nil
}

func dumpCoreAPIGroup(logger *slog.Logger, spec *swaggerSpec) types.APIGroup {
	logger.Info("Scanning…")

	g := types.APIGroup{
		Name:        "",
		APIVersions: []types.APIVersion{},
	}

	// find all versions within this API group
	for path := range spec.Paths {
		match := coreApiVersionPath.FindStringSubmatch(path)
		if match == nil {
			continue
		}

		g.APIVersions = append(g.APIVersions, dumpCoreAPIVersion(logger.With("version", match[1]), spec, match[1]))
	}

	return g
}

func dumpAPIGroup(logger *slog.Logger, spec *swaggerSpec, apiGroup string) types.APIGroup {
	logger.Info("Scanning…")

	g := types.APIGroup{
		Name:        apiGroup,
		APIVersions: []types.APIVersion{},
	}

	prefix := fmt.Sprintf("/apis/%s/", apiGroup)

	// find all versions within this API group
	for path := range spec.Paths {
		if !strings.HasPrefix(path, prefix) {
			continue
		}

		match := apisVersionPath.FindStringSubmatch(path)
		if match == nil {
			continue
		}

		g.APIVersions = append(g.APIVersions, dumpAPIVersion(logger.With("version", match[1]), spec, apiGroup, match[1]))
	}

	return g
}

func dumpCoreAPIVersion(logger *slog.Logger, spec *swaggerSpec, apiVersion string) types.APIVersion {
	logger.Info("Scanning…")

	g := types.APIVersion{
		Version:   apiVersion,
		Resources: []types.Resource{},
	}

	prefix := fmt.Sprintf("/api/%s/", apiVersion)
	coreApiResourcePath := regexp.MustCompile(fmt.Sprintf(`/api/%s/([^/]+)$`, apiVersion))

	// find all resources within this API group/version
	for path, methodSpecs := range spec.Paths {
		if !strings.HasPrefix(path, prefix) {
			continue
		}

		match := coreApiResourcePath.FindStringSubmatch(path)
		if match == nil {
			continue
		}

		methodSpec := methodSpecs.Post
		if methodSpec == nil {
			methodSpec = methodSpecs.Get
			if methodSpec == nil {
				panic(fmt.Sprintf("found not method for path %s", path))
			}
		}

		// check if there is a namespaced path for this resource
		// (e.g. "/apis/rbac.authorization.k8s.io/v1alpha1/namespaces/{namespace}/rolebindings")
		pluralName := match[1]
		pattern := fmt.Sprintf("/namespaces/{namespace}/%s", pluralName)

		namespaced := false
		for path := range spec.Paths {
			if !strings.HasPrefix(path, prefix) {
				continue
			}

			if strings.HasSuffix(path, pattern) {
				namespaced = true
				break
			}
		}

		reslogger := logger.With("resource", methodSpec.KubernetesGVK.Kind, "namespaced", namespaced)
		reslogger.Info("Found resource.")

		res := types.Resource{
			Kind:        methodSpec.KubernetesGVK.Kind,
			Namespaced:  namespaced,
			Plural:      pluralName,
			Singular:    strings.ToLower(methodSpec.KubernetesGVK.Kind),
			Description: getResourceDescription(spec, methodSpec.Responses.OK.Schema.Ref),
		}

		g.Resources = append(g.Resources, res)
	}

	return g
}

func dumpAPIVersion(logger *slog.Logger, spec *swaggerSpec, apiGroup string, apiVersion string) types.APIVersion {
	logger.Info("Scanning…")

	g := types.APIVersion{
		Version:   apiVersion,
		Resources: []types.Resource{},
	}

	prefix := fmt.Sprintf("/apis/%s/%s/", apiGroup, apiVersion)

	// find all resources within this API group/version
	for path, methodSpecs := range spec.Paths {
		if !strings.HasPrefix(path, prefix) {
			continue
		}

		match := apisResourcePath.FindStringSubmatch(path)
		if match == nil {
			continue
		}

		methodSpec := methodSpecs.Post
		if methodSpec == nil {
			methodSpec = methodSpecs.Get
			if methodSpec == nil {
				panic(fmt.Sprintf("found not method for path %s", path))
			}
		}

		// check if there is a namespaced path for this resource
		// (e.g. "/apis/rbac.authorization.k8s.io/v1alpha1/namespaces/{namespace}/rolebindings")
		pluralName := match[1]
		pattern := fmt.Sprintf("/namespaces/{namespace}/%s", pluralName)

		namespaced := false
		for path := range spec.Paths {
			if !strings.HasPrefix(path, prefix) {
				continue
			}

			if strings.HasSuffix(path, pattern) {
				namespaced = true
				break
			}
		}

		reslogger := logger.With("resource", methodSpec.KubernetesGVK.Kind, "namespaced", namespaced)
		reslogger.Info("Found resource.")

		res := types.Resource{
			Kind:        methodSpec.KubernetesGVK.Kind,
			Namespaced:  namespaced,
			Plural:      pluralName,
			Singular:    strings.ToLower(methodSpec.KubernetesGVK.Kind),
			Description: getResourceDescription(spec, methodSpec.Responses.OK.Schema.Ref),
		}

		g.Resources = append(g.Resources, res)
	}

	return g
}

func getResourceDescription(spec *swaggerSpec, ref string) string {
	// a ref looks like "#/definitions/io.k8s.api.core.v1.ReplicationControllerList"
	key := path.Base(ref)

	// during scanning we work with List requests, but want the description for each singular resource
	key = strings.TrimSuffix(key, "List")

	if definition, exists := spec.Definitions[key]; exists {
		return definition.Description
	}

	return ""
}
