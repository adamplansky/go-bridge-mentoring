package crawler

import (
	"encoding/json"
	"net/url"
	"sync"
)

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges Edges  `json:"links"`
	mu    sync.RWMutex
}

func (g *Graph) EdgeExists(src url.URL, dst url.URL) bool {
	dsts := g.Edges[Node{URL: src}]
	target := Node{URL: dst}
	for _, dst := range dsts {
		if target == dst {
			return true
		}
	}
	return false
}
func (g *Graph) AddNode(URL url.URL) {
	for _, node := range g.Nodes {
		if node.URL == URL {
			return
		}
	}
	g.Nodes = append(g.Nodes, Node{URL})
}
func (g *Graph) AddEdge(src url.URL, dst url.URL) {
	srcNode := Node{src}
	dstNode := Node{dst}
	g.mu.Lock()
	defer g.mu.Unlock()

	g.AddNode(src)
	g.AddNode(dst)

	if g.Edges == nil {
		g.Edges = make(map[Node][]Node)
	}

	if g.EdgeExists(src, dst) {
		return
	}

	g.Edges[srcNode] = append(g.Edges[srcNode], dstNode)
}

var _ json.Marshaler = &Node{}

type Node struct {
	URL url.URL `json:"id"`
}

func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID string `json:"id"`
	}{
		ID: n.URL.String(),
	})
}

var _ json.Marshaler = &Edges{}

type Edges map[Node][]Node

func (e *Edges) MarshalJSON() ([]byte, error) {
	type jsonEdge struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Type   string `json:"type"`
	}

	s := make([]jsonEdge, 0)
	for src, dsts := range *e {
		for _, dst := range dsts {
			s = append(s, jsonEdge{
				Source: src.URL.String(),
				Target: dst.URL.String(),
				Type:   "link",
			})
		}
	}
	return json.Marshal(s)
}
