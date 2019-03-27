package main

import (
	"flag"
	"fmt"
	"github.com/n3wscott/knap/pkg/config"
	"github.com/n3wscott/knap/pkg/eventing"
	"github.com/tmc/dot"
	"k8s.io/client-go/dynamic"
	"strings"

	"log"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// To use:
//   go run cmd/dot/graph.go cmd/dot/flags.go | dot -Tpng  > output.png &&  open output.png

func main() {
	flag.Parse()

	cfg, err := config.BuildClientConfig(kubeconfig, cluster)
	if err != nil {
		log.Fatalf("Error building kubeconfig", err)
	}

	dynamicClient := dynamic.NewForConfigOrDie(cfg)

	ns := "default"

	c := eventing.New(dynamicClient)

	g := dot.NewGraph("G")
	_ = g.Set("label", "Triggers in "+ns)

	g.Set("rankdir", "LR")

	nodes := make(map[string]*dot.Node)

	for _, trigger := range c.Triggers(ns) {

		broker := trigger.Spec.Broker

		var bn *dot.Node
		var ok bool
		if bn, ok = nodes[brokerKey(broker)]; !ok {
			bn = dot.NewNode("Broker " + broker)
			g.AddNode(bn)
			nodes[brokerKey(broker)] = bn
		}

		tn := dot.NewNode("Trigger " + trigger.Name)
		g.AddNode(tn)
		nodes[triggerKey(trigger.Name)] = tn

		e := dot.NewEdge(bn, tn)
		//_ = e.Set("dir", "none") // "forward" "back" "both" "none"
		g.AddEdge(e)

		label := ""
		if trigger.Spec.Filter != nil && trigger.Spec.Filter.SourceAndType != nil {
			label = fmt.Sprintf("Source:%s\nType:%s",
				trigger.Spec.Filter.SourceAndType.Source,
				trigger.Spec.Filter.SourceAndType.Type,
			)
		}

		if trigger.Spec.Subscriber != nil {
			key := ""
			subscriber := "?"

			if trigger.Spec.Subscriber.DNSName != nil {
				subscriber = *trigger.Spec.Subscriber.DNSName
				key = uriKey(*trigger.Spec.Subscriber.DNSName)
			} else if trigger.Spec.Subscriber.Ref != nil {
				subscriber = trigger.Spec.Subscriber.Ref.Kind + "/" + trigger.Spec.Subscriber.Ref.Name
				key = refKey(
					trigger.Spec.Subscriber.Ref.APIVersion,
					trigger.Spec.Subscriber.Ref.Kind,
					trigger.Spec.Subscriber.Ref.Name,
				)
			}
			var sub *dot.Node
			var ok bool
			if sub, ok = nodes[key]; !ok {
				sub = dot.NewNode("subscriber " + subscriber)
				nodes[key] = sub
				g.AddNode(sub)
			}

			dns := dot.NewNode("subscriber " + subscriber)
			g.AddNode(dns)
			e := dot.NewEdge(tn, dns)

			if err := e.Set("label", label); err != nil {
				log.Fatalf("failed to set label on edge: %s", err)
			}

			g.AddEdge(e)
		}
	}

	fmt.Print(g.String())
}

/*

   ref:
     apiVersion: serving.knative.dev/v1alpha1
     kind: Service
     name: slash-command-parser

*/

func key(group, version, kind, name string) string {
	return strings.ToLower(fmt.Sprintf("%s/%s/%s/%s", group, version, kind, name))
}

func uriKey(uri string) string {
	return strings.ToLower(fmt.Sprintf("uri/%s", uri))
}

func refKey(apiVersion, kind, name string) string {
	return strings.ToLower(fmt.Sprintf("%s/%s/%s", apiVersion, kind, name))
}

func eventingKey(kind, name string) string {
	return key("eventing.knative.dev", "v1alpha1", kind, name)
}

func triggerKey(name string) string {
	return eventingKey("trigger", name)
}

func brokerKey(name string) string {
	return eventingKey("broker", name)
}
