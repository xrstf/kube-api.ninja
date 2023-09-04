// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"go.xrstf.de/kube-api.ninja/pkg/swaggerdumper"
	"go.xrstf.de/kube-api.ninja/pkg/version"
)

type appOptions struct {
	swaggerFile       string
	kubernetesVersion string
}

func (opts *appOptions) AddFlags(fs *flag.FlagSet) {
	flag.StringVar(&opts.swaggerFile, "swagger-file", "", "The Swagger file to read.")
	flag.StringVar(&opts.kubernetesVersion, "kubernetes-version", "", "The Kubernetes version the Swagger file belongs to.")
}

func (opts *appOptions) Validate() error {
	if opts.swaggerFile == "" {
		return errors.New("no -swagger-file specified")
	}

	if opts.kubernetesVersion == "" {
		return errors.New("no -kubernetes-version specified")
	}

	if _, err := version.ParseSemver(opts.kubernetesVersion); err != nil {
		return fmt.Errorf("invalid Kubernetes version: %w", err)
	}

	return nil
}

func main() {
	opts := appOptions{}

	opts.AddFlags(flag.CommandLine)
	flag.Parse()

	if err := opts.Validate(); err != nil {
		log.Fatalf("Invalid command line: %v", err)
	}

	releaseData, err := swaggerdumper.DumpSwaggerSpec(opts.swaggerFile, opts.kubernetesVersion)
	if err != nil {
		log.Fatalf("Failed to dump Swagger spec: %v", err)
	}

	releaseData.Sort()

	if false || true {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		if err := encoder.Encode(releaseData); err != nil {
			log.Fatalf("Failed to JSON encode result: %v", err)
		}
	}
}
