/*
Copyright 2018 The Kubernetes Authors.

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

package apidocs

import (
	"fmt"
	"log"
	"sort"

	"go.xrstf.de/kube-api.ninja/pkg/apidocs/api"
	"go.xrstf.de/kube-api.ninja/pkg/apidocs/types"
)

func Generate(opts types.Options) error {
	// Load the apidocs config for the requested release
	config, err := api.NewConfig(opts)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	PrintInfo(opts, config)

	var title string
	if opts.BuildOperations {
		title = "Kubernetes API Reference Docs"
	} else {
		title = "Kubernetes Resource Reference Docs"
	}

	writer, err := newHTMLWriter(opts, config, title)
	if err != nil {
		return err
	}

	if err := writer.WriteOverview(); err != nil {
		return err
	}

	// Write API groups
	if err := writer.WriteAPIGroupVersions(config.Definitions.GroupVersions); err != nil {
		return err
	}

	// Write resource definitions
	for _, c := range config.ResourceCategories {
		if err := writer.WriteResourceCategory(c.Name, c.Include); err != nil {
			return err
		}

		for _, r := range c.Resources {
			if r.Definition == nil {
				log.Printf("Warning: Missing definition for item in TOC %s", r.Name)
				continue
			}
			if err := writer.WriteResource(r); err != nil {
				return err
			}
		}
	}

	// Write orphaned operation endpoints
	orphanedIDs := make([]string, 0)
	for id, o := range config.Operations {
		if o.Definition == nil && !config.OpExcluded(o.ID) {
			orphanedIDs = append(orphanedIDs, id)
		}
	}

	if len(orphanedIDs) > 0 {
		if err := writer.WriteOrphanedOperationsOverview(); err != nil {
			return err
		}

		sort.Strings(orphanedIDs)

		for _, opKey := range orphanedIDs {
			if err := writer.WriteOperation(config.Operations[opKey]); err != nil {
				return err
			}
		}
	}

	if err := writer.WriteDefinitionsOverview(); err != nil {
		return err
	}

	// Add other definition imports
	definitions := api.SortDefinitionsByName{}
	for _, d := range config.Definitions.All {
		// Don't add definitions for top level resources in the toc or inlined resources
		if d.InToc || d.IsInlined || d.IsOldVersion {
			continue
		}
		definitions = append(definitions, d)
	}
	sort.Sort(definitions)
	for _, d := range definitions {
		if err := writer.WriteDefinition(d); err != nil {
			return err
		}
	}

	if err := writer.WriteOldVersionsOverview(); err != nil {
		return err
	}

	oldversions := api.SortDefinitionsByName{}
	for _, d := range config.Definitions.All {
		// Don't add definitions for top level resources in the toc or inlined resources
		if d.IsOldVersion {
			oldversions = append(oldversions, d)
		}
	}
	sort.Sort(oldversions)
	for _, d := range oldversions {
		// Skip Inlined definitions
		if d.IsInlined {
			continue
		}
		r := &api.Resource{Definition: d, Name: d.Name}
		if err := writer.WriteResource(r); err != nil {
			return err
		}
	}

	if err := writer.WriteIndex(); err != nil {
		return err
	}

	return nil
}
