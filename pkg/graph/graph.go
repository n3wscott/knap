package graph

import (
	"fmt"
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	duckv1alpha1 "github.com/n3wscott/knap/pkg/apis/duck/v1alpha1"
	"github.com/tmc/dot"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

type Graph struct {
	*dot.Graph
	nodes     map[string]*dot.Node
	subgraphs map[string]*dot.SubGraph
	dnsToKey  map[string]string // maps domain name to node key
}

func New(ns string) *Graph {
	g := dot.NewGraph("G")
	_ = g.Set("shape", "box")
	_ = g.Set("label", "Triggers in "+ns)
	_ = g.Set("rankdir", "LR")

	graph := &Graph{
		Graph:     g,
		nodes:     make(map[string]*dot.Node),
		subgraphs: make(map[string]*dot.SubGraph),
		dnsToKey:  make(map[string]string),
	}

	return graph
}

func (g *Graph) AddBroker(broker eventingv1alpha1.Broker) {
	key := brokerKey(broker.Name)
	dns := brokerDNS(broker)
	bn := dot.NewNode("Broker " + dns)
	_ = bn.Set("shape", "oval")
	_ = bn.Set("label", "Ingress")

	//bn.Set("style", "invis")
	//g.AddNode(bn)

	g.nodes[key] = bn
	g.dnsToKey[dns] = key

	bg := dot.NewSubgraph(fmt.Sprintf("cluster_%d", len(g.subgraphs)))
	_ = bg.Set("label", fmt.Sprintf("Broker %s\n%s", broker.Name, dns))
	g.subgraphs[key] = bg
	bg.AddNode(bn)
	g.AddSubgraph(bg)
}

func (g *Graph) AddSource(source duckv1alpha1.SourceType) {
	key := gvkKey(source.GroupVersionKind(), source.Name)
	sn := dot.NewNode(fmt.Sprintf("Source %s\nKind: %s\n%s", source.Name, source.Kind, source.APIVersion))
	_ = sn.Set("shape", "box")
	g.AddNode(sn)
	g.nodes[key] = sn

	sink := sinkDNS(source)

	if sink != "" {
		var bn *dot.Node
		var bk string
		var ok bool
		if bk, ok = g.dnsToKey[sink]; !ok {
			// TODO: unknown sink.
			bn = dot.NewNode("UnknownSink " + sink)
			g.AddNode(bn)
		} else {
			if bn, ok = g.nodes[bk]; !ok {
				// TODO: unknown broker.
				bn = dot.NewNode("UnknownSink " + sink)
				g.AddNode(bn)
			}
		}

		e := dot.NewEdge(sn, bn)
		if sg, ok := g.subgraphs[bk]; ok {
			// This is not working.
			_ = e.Set("lhead", sg.Name())
		}
		g.AddEdge(e)
	}
}

func (g *Graph) AddTrigger(trigger eventingv1alpha1.Trigger) {
	broker := trigger.Spec.Broker
	bk := brokerKey(broker)
	bn, ok := g.nodes[bk]
	if !ok {
		bn = dot.NewNode("UnknownBroker " + broker)
		g.AddNode(bn)
		g.nodes[bk] = bn
	}

	tn := dot.NewNode("Trigger " + trigger.Name)
	tn.Set("shape", "box")

	if sg, ok := g.subgraphs[bk]; ok {
		sg.AddNode(tn)
	} else {
		g.AddNode(tn)
	}
	g.nodes[triggerKey(trigger.Name)] = tn

	//e := dot.NewEdge(bn, tn)
	//_ = e.Set("dir", "none") // "forward" "back" "both" "none"
	//g.AddEdge(e)

	if trigger.Spec.Filter != nil && trigger.Spec.Filter.SourceAndType != nil {
		label := fmt.Sprintf("Source:%s\nType:%s",
			trigger.Spec.Filter.SourceAndType.Source,
			trigger.Spec.Filter.SourceAndType.Type,
		)
		_ = tn.Set("label", fmt.Sprintf("%s\n%s", tn.Name(), label))
	}

	if trigger.Spec.Subscriber != nil {
		key := ""
		subscriber := "?"

		if trigger.Spec.Subscriber.DNSName != nil {
			subscriber = *trigger.Spec.Subscriber.DNSName
			key = uriKey(*trigger.Spec.Subscriber.DNSName)
		} else if trigger.Spec.Subscriber.Ref != nil {
			subscriber = fmt.Sprintf("%s\nKind: %s\n%s",
				trigger.Spec.Subscriber.Ref.Name,
				trigger.Spec.Subscriber.Ref.Kind,
				trigger.Spec.Subscriber.Ref.APIVersion,
			)
			key = refKey(
				trigger.Spec.Subscriber.Ref.APIVersion,
				trigger.Spec.Subscriber.Ref.Kind,
				trigger.Spec.Subscriber.Ref.Name,
			)
		}
		var sub *dot.Node
		var ok bool
		if sub, ok = g.nodes[key]; !ok {
			sub = dot.NewNode("Subscriber " + subscriber)
			g.nodes[key] = sub
			g.AddNode(sub)
		}

		e := dot.NewEdge(tn, sub)

		g.AddEdge(e)
	}
}

func sinkDNS(source duckv1alpha1.SourceType) string {
	if source.Status.SinkURI != nil {
		uri := *(source.Status.SinkURI)
		if !strings.HasSuffix(uri, "/") {
			uri += "/"
		}
		return uri
	}
	return ""
}

func brokerDNS(broker eventingv1alpha1.Broker) string {
	uri := fmt.Sprintf("http://%s", broker.Status.Address.Hostname)
	if !strings.HasSuffix(uri, "/") {
		uri += "/"
	}
	return uri
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
