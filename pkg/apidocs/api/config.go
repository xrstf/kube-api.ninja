/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"go.xrstf.de/kube-api.ninja/pkg/apidocs/types"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"gopkg.in/yaml.v3"
)

func NewConfig(opts types.Options) (*Config, error) {
	config, err := LoadConfigFromYAML(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to load config yaml: %w", err)
	}

	spec, err := config.loadOpenApiSpec()
	if err != nil {
		return nil, fmt.Errorf("failed to load openapi spec: %w", err)
	}

	// Parse spec version
	config.ParseSpecInfo(spec)

	// Set the spec version
	config.SpecVersion = fmt.Sprintf("v%s.%s", opts.KubernetesRelease, "0")

	// Initialize all of the operations
	defs, err := NewDefinitions(config, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to init definitions: %w", err)
	}
	config.Definitions = *defs

	// Initialization for ToC resources only
	if err := config.visitResourcesInToc(); err != nil {
		return nil, fmt.Errorf("failed to visit resources in TOC: %w", err)
	}

	if err := config.initOperations(spec); err != nil {
		return nil, fmt.Errorf("failed to init operations: %w", err)
	}

	config.CleanUp()

	return config, nil
}

func (c *Config) getExampleProviders() []ExampleProvider {
	if c.options.BuildOperations {
		return ExampleProviders
	} else {
		return EmptyExampleProviders
	}
}

// initOperations returns all Operations found in the Documents
func (c *Config) initOperations(spec *loads.Document) error {
	c.Operations = Operations{}
	VisitOperations(spec, func(operation Operation) {
		op := &operation
		op.exampleProviders = c.getExampleProviders()
		c.Operations[op.ID] = op
	})

	if err := c.mapOperationsToDefinitions(); err != nil {
		return err
	}

	VisitOperations(spec, func(target Operation) {
		if op, ok := c.Operations[target.ID]; !ok || op.Definition == nil {
			if !c.OpExcluded(op.ID) {
				fmt.Printf("\033[31mNo Definition found for %s [%s].\033[0m\n", op.ID, op.Path)
			} else {
				fmt.Printf("Op excluded: %s\n", op.ID)
			}
		}
	})

	if err := c.initOperationParameters(spec); err != nil {
		return err
	}

	// Clear the operations.  We still have to calculate the operations because that is how we determine
	// the API Group for each definition.
	if !c.options.BuildOperations {
		c.Operations = Operations{}
		c.OperationCategories = []OperationCategory{}
		for _, d := range c.Definitions.All {
			d.OperationCategories = []*OperationCategory{}
		}
	}

	return nil
}

func (c *Config) OpExcluded(op string) bool {
	for _, pattern := range c.ExcludedOperations {
		if strings.Contains(op, pattern) {
			return true
		}
	}
	return false
}

// CleanUp sorts and dedups fields
func (c *Config) CleanUp() {
	for _, d := range c.Definitions.All {
		sort.Sort(d.AppearsIn)
		sort.Sort(d.Fields)
		dedup := SortDefinitionsByName{}
		var last *Definition
		for _, i := range d.AppearsIn {
			if last != nil &&
				i.Name == last.Name &&
				i.Group.String() == last.Group.String() &&
				i.Version.String() == last.Version.String() {
				continue
			}
			last = i
			dedup = append(dedup, i)
		}
		d.AppearsIn = dedup
	}
}

// LoadConfigFromYAML reads the config yaml file into a struct
func LoadConfigFromYAML(opts types.Options) (*Config, error) {
	releaseDir := filepath.Join(opts.ConfigDirectory, opts.KubernetesRelease)

	config := &Config{
		options:           opts,
		SectionsDirectory: filepath.Join(releaseDir, "sections"),
		ReleaseDirectory:  releaseDir,
	}

	contents, err := os.ReadFile(filepath.Join(releaseDir, "apidocs.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read apidocs.yaml file: %w", err)
	}
	if err = yaml.Unmarshal(contents, config); err != nil {
		return nil, err
	}

	writeCategory := OperationCategory{
		Name: "Write Operations",
		OperationTypes: []OperationType{
			{
				Name:  "Create",
				Match: "create${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Create Eviction",
				Match: "create${group}${version}(Namespaced)?${resource}Eviction",
			},
			{
				Name:  "Patch",
				Match: "patch${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Replace",
				Match: "replace${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Delete",
				Match: "delete${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Delete Collection",
				Match: "delete${group}${version}Collection(Namespaced)?${resource}",
			},
		},
	}

	readCategory := OperationCategory{
		Name: "Read Operations",
		OperationTypes: []OperationType{
			{
				Name:  "Read",
				Match: "read${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "List",
				Match: "list${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "List All Namespaces",
				Match: "list${group}${version}(Namespaced)?${resource}ForAllNamespaces",
			},
			{
				Name:  "Watch",
				Match: "watch${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Watch List",
				Match: "watch${group}${version}(Namespaced)?${resource}List",
			},
			{
				Name:  "Watch List All Namespaces",
				Match: "watch${group}${version}(Namespaced)?${resource}ListForAllNamespaces",
			},
		},
	}

	statusCategory := OperationCategory{
		Name: "Status Operations",
		OperationTypes: []OperationType{
			{
				Name:  "Patch Status",
				Match: "patch${group}${version}(Namespaced)?${resource}Status",
			},
			{
				Name:  "Read Status",
				Match: "read${group}${version}(Namespaced)?${resource}Status",
			},
			{
				Name:  "Replace Status",
				Match: "replace${group}${version}(Namespaced)?${resource}Status",
			},
		},
	}

	ephemaralCategory := OperationCategory{
		Name: "EphemeralContainers Operations",
		OperationTypes: []OperationType{
			{
				Name:  "Patch EphemeralContainers",
				Match: "patch${group}${version}(Namespaced)?${resource}Ephemeralcontainers",
			},
			{
				Name:  "Read EphemeralContainers",
				Match: "read${group}${version}(Namespaced)?${resource}Ephemeralcontainers",
			},
			{
				Name:  "Replace EphemeralContainers",
				Match: "replace${group}${version}(Namespaced)?${resource}Ephemeralcontainers",
			},
		},
	}

	config.OperationCategories = append([]OperationCategory{writeCategory, readCategory, statusCategory, ephemaralCategory}, config.OperationCategories...)

	return config, nil
}

const (
	PATH  = "path"
	QUERY = "query"
	BODY  = "body"
)

func (c *Config) initOperationParameters(doc *loads.Document) error {
	s := c.Definitions
	for _, op := range c.Operations {
		pathItem := op.item

		var (
			location string
			param    spec.Parameter
			found    bool
		)

		// Path parameters
		for _, p := range pathItem.Parameters {
			if p.In == "" {
				paramID := strings.Split(p.Ref.String(), "/")[2]
				if param, found = doc.Spec().Parameters[paramID]; found {
					location = param.In
				}
			} else {
				location = p.In
				param = p
			}

			switch location {
			case PATH:
				op.PathParams = append(op.PathParams, s.parameterToField(param))
			case QUERY:
				op.QueryParams = append(op.QueryParams, s.parameterToField(param))
			case BODY:
				op.BodyParams = append(op.BodyParams, s.parameterToField(param))
			default:
				return fmt.Errorf("unknown location %q", location)
			}
		}

		// Query parameters
		location = ""
		for _, p := range op.op.Parameters {
			if p.In == "" {
				paramID := strings.Split(p.Ref.String(), "/")[2]
				if param, found = doc.Spec().Parameters[paramID]; found {
					location = param.In
				}
			} else {
				location = p.In
				param = p
			}

			switch location {
			case PATH:
				op.PathParams = append(op.PathParams, s.parameterToField(param))
			case QUERY:
				op.QueryParams = append(op.QueryParams, s.parameterToField(param))
			case BODY:
				op.BodyParams = append(op.BodyParams, s.parameterToField(param))
			default:
				return fmt.Errorf("unknown location %q", location)
			}
		}

		for code, response := range op.op.Responses.StatusCodeResponses {
			if response.Schema == nil {
				continue
			}
			r := &HttpResponse{
				Field: Field{
					Description: strings.ReplaceAll(response.Description, "\n", " "),
					Type:        GetTypeName(*response.Schema),
					Name:        fmt.Sprintf("%d", code),
				},
				Code: fmt.Sprintf("%d", code),
			}
			if IsComplex(*response.Schema) {
				r.Definition, _ = s.GetForSchema(*response.Schema)
				if r.Definition != nil {
					r.Definition.FoundInOperation = true
				}
			}
			op.HttpResponses = append(op.HttpResponses, r)
		}
	}

	return nil
}

func (c *Config) getOperationGroupName(group string) string {
	for k, v := range c.OperationGroupMap {
		if strings.ToLower(group) == k {
			return v
		}
	}
	return strings.Title(group)
}

func (c *Config) getOperationID(match string, group string, version ApiVersion, kind string) string {
	ver := []rune(string(version))
	ver[0] = unicode.ToUpper(ver[0])

	match = strings.ReplaceAll(match, "${group}", group)
	match = strings.ReplaceAll(match, "${version}", string(ver))
	match = strings.ReplaceAll(match, "${resource}", kind)

	return match
}

func (c *Config) setOperation(match, namespace string, ot *OperationType, oc *OperationCategory, d *Definition) error {
	key := strings.ReplaceAll(match, "(Namespaced)?", namespace)
	if o, ok := c.Operations[key]; ok {
		// Each operation should have exactly 1 definition
		if o.Definition != nil {
			return fmt.Errorf(
				"Found multiple matching definitions [%s/%s/%s, %s/%s/%s] for operation key: %s",
				d.Group, d.Version, d.Name, o.Definition.Group, o.Definition.Version, o.Definition.Name, key)
		}
		o.Type = *ot
		o.Definition = d
		if err := o.initExample(c); err != nil {
			return fmt.Errorf("failed to init example: %w", err)
		}
		oc.Operations = append(oc.Operations, o)
	}

	return nil
}

// mapOperationsToDefinitions adds operations to the definitions they operate
func (c *Config) mapOperationsToDefinitions() error {
	for _, d := range c.Definitions.All {
		if d.IsInlined {
			continue
		}

		// XXX: The TokenRequest definition has operation defined as "createCoreV1NamespacedServiceAccountToken"!
		if d.Name == "TokenRequest" && d.Group.String() == "authentication" && d.Version == "v1" {
			if o, ok := c.Operations["createCoreV1NamespacedServiceAccountToken"]; ok {
				o.Definition = d
				o.Definition.InToc = true
				if err := o.initExample(c); err != nil {
					return fmt.Errorf("failed to init example: %w", err)
				}
			}

			continue
		}

		for i := range c.OperationCategories {
			oc := c.OperationCategories[i]
			for j := range oc.OperationTypes {
				ot := oc.OperationTypes[j]
				groupName := c.getOperationGroupName(d.Group.String())
				operationId := c.getOperationID(ot.Match, groupName, d.Version, d.Name)
				if err := c.setOperation(operationId, "Namespaced", &ot, &oc, d); err != nil {
					return err
				}
				if err := c.setOperation(operationId, "", &ot, &oc, d); err != nil {
					return err
				}
			}

			if len(oc.Operations) > 0 {
				d.OperationCategories = append(d.OperationCategories, &oc)
			}
		}
	}

	return nil
}

// For each resource in the ToC, look up its definition and visit it.
func (c *Config) visitResourcesInToc() error {
	missing := false
	for _, cat := range c.ResourceCategories {
		for _, r := range cat.Resources {
			if d, ok := c.Definitions.GetByVersionKind(r.Group, r.Version, r.Name); ok {
				d.InToc = true // Mark as in Toc
				if err := d.initExample(c); err != nil {
					return fmt.Errorf("failed to init example: %w", err)
				}
				r.Definition = d
			} else {
				fmt.Printf("\033[31mCould not find definition for resource in TOC: %s %s %s.\033[0m\n", r.Group, r.Version, r.Name)
				missing = true
			}
		}
	}
	if missing {
		fmt.Printf("\033[36mAll known definitions: %v\033[0m\n", c.Definitions.All)
	}

	return nil
}
