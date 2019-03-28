package main

import (
	"flag"
	"github.com/n3wscott/knap/pkg/config"
	"github.com/n3wscott/knap/pkg/knative"
	"k8s.io/client-go/dynamic"

	"log"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	flag.Parse()

	cfg, err := config.BuildClientConfig(kubeconfig, cluster)
	if err != nil {
		log.Fatalf("Error building kubeconfig", err)
	}

	dynamicClient := dynamic.NewForConfigOrDie(cfg)

	ns := "default"

	c := knative.New(dynamicClient)

	for _, t := range c.Triggers(ns) {
		if len(t.ObjectMeta.OwnerReferences) > 0 {
			for _, o := range t.ObjectMeta.OwnerReferences {
				log.Printf("%s %s - owned by %s %s %s", t.Kind, t.Name, o.Name, o.Kind, o.APIVersion)
			}
		} else {
			log.Printf("%s %s", t.Kind, t.Name)
		}
	}

	for _, t := range c.Subscriptions(ns) {
		if len(t.ObjectMeta.OwnerReferences) > 0 {
			for _, o := range t.ObjectMeta.OwnerReferences {
				log.Printf("%s %s - owned by %s %s %s", t.Kind, t.Name, o.Name, o.Kind, o.APIVersion)
			}
		} else {
			log.Printf("%s %s", t.Kind, t.Name)
		}
	}

	for _, t := range c.Brokers(ns) {
		if len(t.ObjectMeta.OwnerReferences) > 0 {
			for _, o := range t.ObjectMeta.OwnerReferences {
				log.Printf("%s %s - owned by %s %s %s", t.Kind, t.Name, o.Name, o.Kind, o.APIVersion)
			}
		} else {
			log.Printf("%s %s", t.Kind, t.Name)
		}
	}

	for _, t := range c.Channels(ns) {
		if len(t.ObjectMeta.OwnerReferences) > 0 {
			for _, o := range t.ObjectMeta.OwnerReferences {
				log.Printf("%s %s - owned by %s %s %s", t.Kind, t.Name, o.Name, o.Kind, o.APIVersion)
			}
		} else {
			log.Printf("%s %s", t.Kind, t.Name)
		}
	}

	for _, t := range c.Sources("default") {
		if len(t.ObjectMeta.OwnerReferences) > 0 {
			for _, o := range t.ObjectMeta.OwnerReferences {
				log.Printf("%s %s - owned by %s %s %s", t.Kind, t.Name, o.Name, o.Kind, o.APIVersion)
			}
		} else {
			/*
				group: sources.eventing.knative.dev
				names:
				categories:
					- all
					- knative
					- eventing
					- sources
				kind: ContainerSource
				listKind: ContainerSourceList
				plural: containersources
				singular: containersource
			*/
			//log.Printf("Source : %s %s %s", t.APIVersion, t.Kind, t.Name)
			log.Printf("%s %s %s", t.Name, t.GroupVersionKind().String(), *t.Status.SinkURI)
		}
	}

	for _, t := range c.KnServices(ns) {
		if len(t.ObjectMeta.OwnerReferences) > 0 {
			for _, o := range t.ObjectMeta.OwnerReferences {
				log.Printf("%s %s - owned by %s %s %s", t.Kind, t.Name, o.Name, o.Kind, o.APIVersion)
			}
		} else {
			log.Printf("%s %s", t.Kind, t.Name)
		}
	}
}
