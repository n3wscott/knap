package main

import (
	"flag"
	"fmt"
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	duckv1alpha1 "github.com/n3wscott/knap/pkg/apis/duck/v1alpha1"
	"github.com/n3wscott/knap/pkg/config"
	"github.com/n3wscott/knap/pkg/eventing"
	"github.com/tmc/dot"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
	"strings"

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
	_ = g.Set("shape", "box")
	_ = g.Set("label", "Triggers in "+ns)
	_ = g.Set("rankdir", "LR")

	nodes := make(map[string]*dot.Node)

	subgraphs := make(map[string]*dot.SubGraph)

	dnsToKey := make(map[string]string) // maps domain name to node key

	// load the brokers
	for _, broker := range c.Brokers(ns) {
		key := brokerKey(broker.Name)
		dns := brokerDNS(broker)
		bn := dot.NewNode("Broker " + dns)
		bn.Set("shape", "oval")
		_ = bn.Set("label", "Ingress")

		//bn.Set("style", "invis")
		//g.AddNode(bn)

		nodes[key] = bn
		dnsToKey[dns] = key

		bg := dot.NewSubgraph(fmt.Sprintf("cluster_%d", len(subgraphs)))
		_ = bg.Set("label", fmt.Sprintf("Broker %s\n%s", broker.Name, dns))
		subgraphs[key] = bg
		bg.AddNode(bn)
		g.AddSubgraph(bg)
	}

	// load the sources
	for _, source := range c.Sources(ns) {
		key := gvkKey(source.GroupVersionKind(), source.Name)
		sn := dot.NewNode("Source " + source.Name)
		_ = sn.Set("shape", "box")
		g.AddNode(sn)
		nodes[key] = sn

		sink := sinkDNS(source)

		if sink != "" {
			var bn *dot.Node
			var bk string
			var ok bool
			if bk, ok = dnsToKey[sink]; !ok {
				// TODO: unknown sink.
				bn = dot.NewNode("UnknownSink " + sink)
				g.AddNode(bn)
			} else {
				if bn, ok = nodes[bk]; !ok {
					// TODO: unknown broker.
					bn = dot.NewNode("UnknownSink " + sink)
					g.AddNode(bn)
				}
			}

			e := dot.NewEdge(sn, bn)
			if sg, ok := subgraphs[bk]; ok {
				// This is not working.
				_ = e.Set("lhead", sg.Name())
			}
			g.AddEdge(e)
		}
	}

	// load the triggers
	for _, trigger := range c.Triggers(ns) {
		broker := trigger.Spec.Broker
		bk := brokerKey(broker)
		bn, ok := nodes[bk]
		if !ok {
			bn = dot.NewNode("UnknownBroker " + broker)
			g.AddNode(bn)
			nodes[bk] = bn
		}

		tn := dot.NewNode("Trigger " + trigger.Name)
		tn.Set("shape", "box")

		if sg, ok := subgraphs[bk]; ok {
			sg.AddNode(tn)
		} else {
			g.AddNode(tn)
		}
		nodes[triggerKey(trigger.Name)] = tn

		//e := dot.NewEdge(bn, tn)
		//_ = e.Set("dir", "none") // "forward" "back" "both" "none"
		//g.AddEdge(e)

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

func sinkDNS(source duckv1alpha1.SourceType) string {
	if source.Status.SinkURI != nil {
		return fmt.Sprintf("%s/", *(source.Status.SinkURI))
	}
	return ""
}

func brokerDNS(broker eventingv1alpha1.Broker) string {
	return fmt.Sprintf("http://%s/", broker.Status.Address.Hostname)
}

func brokerKey(name string) string {
	return eventingKey("broker", name)
}

func gvkKey(gvk schema.GroupVersionKind, name string) string {
	return strings.ToLower(fmt.Sprintf("%s/%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind, name))
}

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
