package main

import (
	"flag"
	"fmt"
	"github.com/n3wscott/knap/pkg/config"
	"github.com/n3wscott/knap/pkg/graph"
	"github.com/n3wscott/knap/pkg/knative"
	"k8s.io/client-go/dynamic"
	"log"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// To use:
//   go run cmd/dot/graph.go cmd/dot/flags.go | dot -Tpng  > output.png &&  open output.png
// or
//   go run cmd/dot/graph.go cmd/dot/flags.go | dot -Tsvg  > output.svg &&  open output.svg

func main() {
	flag.Parse()

	cfg, err := config.BuildClientConfig(kubeconfig, cluster)
	if err != nil {
		log.Fatalf("Error building kubeconfig", err)
	}

	dynamicClient := dynamic.NewForConfigOrDie(cfg)

	ns := "default"

	c := knative.New(dynamicClient)

	g := graph.New(ns)

	// load the brokers
	for _, broker := range c.Brokers(ns) {
		g.AddBroker(broker)
	}

	// load the sources
	for _, source := range c.Sources(ns) {
		g.AddSource(source)
	}

	// load the triggers
	for _, trigger := range c.Triggers(ns) {
		g.AddTrigger(trigger)
	}

	// load the services
	for _, service := range c.KnServices(ns) {
		g.AddKnService(service)
	}

	fmt.Print(g.String())
}
