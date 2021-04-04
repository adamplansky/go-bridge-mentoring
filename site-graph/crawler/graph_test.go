package crawler

import (
	"encoding/json"
	"net/url"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestGraph_AddEdge(t *testing.T) {
	type edges struct {
		src url.URL
		dst url.URL
	}
	tests := []struct {
		name      string
		edges     []edges
		wantEdges Edges
	}{
		{
			name: "test",
			edges: []edges{
				{
					src: mustParse("https://www.adamplansky.cz/"),
					dst: mustParse("https://www.google.com/"),
				},
				{
					src: mustParse("https://www.google.com/"),
					dst: mustParse("https://www.facebook.com/"),
				},
				{
					src: mustParse("https://www.google.com/"),
					dst: mustParse("https://www.facebook.com/"),
				},
			},
			wantEdges: Edges{
				Node{mustParse("https://www.adamplansky.cz/")}: []Node{
					Node{mustParse("https://www.google.com/")},
				},
				Node{mustParse("https://www.google.com/")}: []Node{
					Node{mustParse("https://www.facebook.com/")},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Graph{
				Nodes: make([]Node, 0),
				Edges: make(map[Node][]Node, 0),
				mu:    sync.RWMutex{},
			}
			for _, edge := range tt.edges {
				g.AddEdge(edge.src, edge.dst)
			}

			require.Empty(t, cmp.Diff(g.Edges, tt.wantEdges))

		})
	}
}

func TestGraph_Marshaller(t *testing.T) {
	type edges struct {
		src url.URL
		dst url.URL
	}
	tests := []struct {
		name       string
		edges      []edges
		jsonResult string
	}{
		{
			name: "test",
			edges: []edges{
				{
					src: mustParse("https://www.adamplansky.cz/"),
					dst: mustParse("https://www.google.com/"),
				},
				{
					src: mustParse("https://www.google.com/"),
					dst: mustParse("https://www.facebook.com/"),
				},
				{
					src: mustParse("https://www.google.com/"),
					dst: mustParse("https://www.facebook.com/"),
				},
			},
			jsonResult: `
{ 
  "nodes": [
    { "id": "https://www.adamplansky.cz/" },
    { "id": "https://www.google.com/" },
    { "id": "https://www.facebook.com/" }
  ], 
  "links": [
    { "source": "https://www.adamplansky.cz/", "target": "https://www.google.com/", "type": "link" },
    { "source": "https://www.google.com/", "target": "https://www.facebook.com/", "type": "link" }
  ]
}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Graph{
				Nodes: make([]Node, 0),
				Edges: make(map[Node][]Node, 0),
			}
			for _, edge := range tt.edges {
				g.AddEdge(edge.src, edge.dst)
			}
			gotJson, err := json.Marshal(g)
			require.NoError(t, err)
			require.JSONEq(t, tt.jsonResult, string(gotJson))

		})
	}
}

func mustParse(URL string) url.URL {
	u, err := url.Parse(URL)
	if err != nil {
		panic(err)
	}
	return *u
}
