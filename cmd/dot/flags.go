package main

import (
	"flag"
	"os/user"
	"path"
)

var (
	cluster    string
	kubeconfig string
)

func init() {
	flag.StringVar(&cluster, "cluster", "",
		"Provide the cluster to test against. Defaults to the current cluster in kubeconfig.")

	var defaultKubeconfig string
	if usr, err := user.Current(); err == nil {
		defaultKubeconfig = path.Join(usr.HomeDir, ".kube/config")
	}

	flag.StringVar(&kubeconfig, "kubeconfig", defaultKubeconfig,
		"Provide the path to the `kubeconfig` file you'd like to use for these tests. The `current-context` will be used.")

}
