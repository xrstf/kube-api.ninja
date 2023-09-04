// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package dumper

import (
	"fmt"
	"strings"

	"go.xrstf.de/kube-api.ninja/pkg/types"
	"k8s.io/client-go/discovery"
)

func DumpClusterData(client *discovery.DiscoveryClient) (*types.KubernetesAPI, error) {
	result := &types.KubernetesAPI{}

	server, err := client.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to discover cluster version: %w", err)
	}

	result.Version = strings.TrimPrefix(server.String(), "v")
	result.Release = fmt.Sprintf("%s.%s", server.Major, server.Minor)

	groups, resourceLists, err := client.ServerGroupsAndResources()
	if err != nil {
		return nil, fmt.Errorf("failed to discover API: %w", err)
	}

	result.APIGroups = make([]types.APIGroup, len(groups))

	for i, apiGroup := range groups {
		g := types.APIGroup{
			Name:             apiGroup.Name,
			PreferredVersion: apiGroup.PreferredVersion.Version,
			APIVersions:      make([]types.APIVersion, len(apiGroup.Versions)),
		}

		for j, apiGroupVersion := range apiGroup.Versions {
			av := types.APIVersion{
				Version: apiGroupVersion.Version,
			}

			// find the resources for this API group version
			for _, resourceList := range resourceLists {
				if resourceList.GroupVersion == apiGroupVersion.GroupVersion {
					av.Resources = make([]types.Resource, 0) // we're not keeping all resources

					for _, resource := range resourceList.APIResources {
						// ignore subresources
						if strings.Contains(resource.Name, "/") {
							continue
						}

						singular := resource.SingularName
						if singular == "" {
							singular = strings.ToLower(resource.Kind)
						}

						av.Resources = append(av.Resources, types.Resource{
							Kind:       resource.Kind,
							Namespaced: resource.Namespaced,
							Singular:   singular,
							Plural:     resource.Name,
						})
					}
					break
				}
			}

			g.APIVersions[j] = av
		}

		result.APIGroups[i] = g
	}

	return result, nil
}
