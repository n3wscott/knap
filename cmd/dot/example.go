package main

import (
	"fmt"
	"github.com/tmc/dot"
)

func main() {
	fmt.Print(Graph())
}

func Graph() string {
	g := dot.NewGraph("G")
	g.Set("label", "Example graph")
	n1, n2 := dot.NewNode("Node1"), dot.NewNode("Node2")

	n1.Set("color", "sienna")

	g.AddNode(n1)
	g.AddNode(n2)

	e := dot.NewEdge(n1, n2)
	e.Set("dir", "both")
	g.AddEdge(e)

	return g.String()
}

// use like:
//   go run cmd/dot/example.go | dot -Tpng  > output.png
