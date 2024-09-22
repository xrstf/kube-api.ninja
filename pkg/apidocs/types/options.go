package types

import "flag"

type Options struct {
	BuildOperations   bool
	AllowErrors       bool
	BuildDirectory    string
	ConfigDirectory   string
	KubernetesRelease string
}

func DefaultOptions() Options {
	return Options{
		BuildOperations: true,
		BuildDirectory:  "_build",
		ConfigDirectory: "data/releases",
	}
}

func (o *Options) AddFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.BuildOperations, "build-operations", o.BuildOperations, "If true build operations in the docs.")
	fs.BoolVar(&o.AllowErrors, "allow-errors", o.AllowErrors, "If true, don't fail on errors.")
	fs.StringVar(&o.BuildDirectory, "build-dir", o.BuildDirectory, "Directory to write generated files to.")
	fs.StringVar(&o.ConfigDirectory, "config-dir", o.ConfigDirectory, "Directory where the apidocs.yaml lives.")
	fs.StringVar(&o.KubernetesRelease, "kubernetes-release", o.KubernetesRelease, "Kubernetes release version.")
}
