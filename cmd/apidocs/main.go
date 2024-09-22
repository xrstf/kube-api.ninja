// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"log"
	"strings"

	"go.xrstf.de/kube-api.ninja/pkg/apidocs"
	"go.xrstf.de/kube-api.ninja/pkg/apidocs/types"
)

func main() {
	opts := types.DefaultOptions()
	opts.AddFlags(flag.CommandLine)

	flag.Parse()

	opts.KubernetesRelease = strings.TrimPrefix(opts.KubernetesRelease, "v")

	if err := apidocs.Generate(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
