package main

import (
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/n3wscott/knap/pkg/config"
	"github.com/n3wscott/knap/pkg/graph"
	"github.com/n3wscott/knap/pkg/knative"
	"k8s.io/client-go/dynamic"
	"log"
	"os"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type envConfig struct {
	// Name of this pod.
	Name string `envconfig:"POD_NAME" required:"true"`

	// Namespace this pod exists in.
	Namespace string `envconfig:"POD_NAMESPACE" required:"true"`
}

// To use:
//   go run cmd/dot/graph.go cmd/dot/flags.go | dot -Tpng  > output.png &&  open output.png
// or
//   go run cmd/dot/graph.go cmd/dot/flags.go | dot -Tsvg  > output.svg &&  open output.svg

func main() {
	flag.Parse()

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	cfg, err := config.BuildClientConfig(kubeconfig, cluster)
	if err != nil {
		log.Fatalf("Error building kubeconfig", err)
	}

	dynamicClient := dynamic.NewForConfigOrDie(cfg)

	c := knative.New(dynamicClient)

	g := graph.New(env.Namespace)

	// load the brokers
	for _, broker := range c.Brokers(env.Namespace) {
		g.AddBroker(broker)
	}

	// load the sources
	for _, source := range c.Sources(env.Namespace) {
		g.AddSource(source)
	}

	// load the triggers
	for _, trigger := range c.Triggers(env.Namespace) {
		g.AddTrigger(trigger)
	}

	// load the services
	for _, service := range c.KnServices(env.Namespace) {
		g.AddKnService(service)
	}

	fmt.Print(g.String())
}
