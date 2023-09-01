// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"

	"go.xrstf.de/kubernetes-apis/pkg/dumper"
	"k8s.io/client-go/discovery"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

type appOptions struct {
	kubeconfig string
}

func (opts *appOptions) AddFlags(fs *flag.FlagSet) {
	flag.StringVar(&opts.kubeconfig, "kubeconfig", "", "The kubeconfig to dump the API from (defaults to $KUBECONFIG).")
}

func (opts *appOptions) Validate() error {
	if opts.kubeconfig == "" {
		opts.kubeconfig = os.Getenv("KUBECONFIG")

		if opts.kubeconfig == "" {
			return errors.New("neither -kubeconfig nor $KUBECONFIG specified")
		}
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

	// load kubeconfig
	config, err := clientcmd.LoadFromFile(opts.kubeconfig)
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// turn config into REST config
	restConfig, err := clientcmd.NewDefaultClientConfig(*config, nil).ClientConfig()
	if err != nil {
		log.Fatalf("Failed to build REST config: %v", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to build discovery client: %v", err)
	}

	releaseData, err := dumper.DumpClusterData(discoveryClient)
	if err != nil {
		log.Fatalf("Failed to dump cluster info: %v", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(releaseData); err != nil {
		log.Fatalf("Failed to JSON encode result: %v", err)
	}
}
