package graph

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGraphAddEdge(t *testing.T) {

	name := "test"
	G := &Graph{Name: name}

	assert.Equal(t, G.Name, name, "simple name check")

	// create 3 vertex in a line
	v1 := &Vertex{Name: "1"}
	v2 := &Vertex{Name: "2"}
	v3 := &Vertex{Name: "3"}

	// and 2 edges
	e1 := &Edge{Name: "e1", Vertices: []*Vertex{v1, v2}}
	e2 := &Edge{Name: "e2", Vertices: []*Vertex{v3, v2}}

	err := G.AddEdgeObj(e1)
	if err != nil {
		t.Errorf("Failed to add edge 1: %v\n", err)
	}

	found, _ := G.FindVertex(v1)
	assert.True(t, found, "vertex 1 should have been added")

	found, _ = G.FindVertex(v2)
	assert.True(t, found, "vertex 2 should have been added")

	found, _ = G.FindVertex(v3)
	assert.False(t, found, "vertex 3 has not been added")

	_, found = G.FindEdge(e1)
	assert.True(t, found, "edge 1 should have been added")

	_, found = G.FindEdge(e2)
	assert.False(t, found, "edge 2 has not been added")

	err = G.AddEdgeObj(e2)
	if err != nil {
		t.Errorf("Failed to add edge 2: %v\n", err)
	}

	found, _ = G.FindVertex(v3)
	assert.True(t, found, "vertex 3 should have been added")

	_, found = G.FindEdge(e2)
	assert.True(t, found, "edge 2 has not been added")

	out, err := G.DotViz()
	if err != nil {
		t.Errorf("Failed to create dotviz: %v\n", err)
	}

	log.Infof("straight line graph:\n\n%s\n", out)
}

func TestGraphAddVertices(t *testing.T) {

	name := "test-triangle"
	G := &Graph{Name: name}

	assert.Equal(t, G.Name, name, "simple name check")

	// create 3 vertex in a line
	v1 := &Vertex{Name: "1"}
	v2 := &Vertex{Name: "2"}
	v3 := &Vertex{Name: "3"}

	err := G.AddVertexObj(v1)
	if err != nil {
		t.Errorf("Failed to add vertex 1: %v\n", err)
	}

	found, _ := G.FindVertex(v1)
	assert.True(t, found, "vertex 1 should have been added")

	err = G.AddVertexObj(v2)
	if err != nil {
		t.Errorf("Failed to add vertex 2: %v\n", err)
	}

	found, _ = G.FindVertex(v2)
	assert.True(t, found, "vertex 2 should have been added")

	err = G.AddVertexObj(v3)
	if err != nil {
		t.Errorf("Failed to add vertex 3: %v\n", err)
	}

	found, _ = G.FindVertex(v3)
	assert.True(t, found, "vertex 3 should have been added")

	_, err = G.AddEdge(v1, v2, nil)
	if err != nil {
		t.Errorf("Failed to add edge 1: %v\n", err)
	}

	etemp1, err := NewEdge(v1, v2, nil)
	if err != nil {
		t.Errorf("Failed to create edge 1: %v\n", err)
	}

	_, found = G.FindEdge(etemp1)
	assert.True(t, found, "edge 1 should have been added")

	etemp2, err := NewEdge(v2, v3, nil)
	if err != nil {
		t.Errorf("Failed to create edge 1: %v\n", err)
	}

	_, found = G.FindEdge(etemp2)
	assert.False(t, found, "edge 2 has not been added yet")

	_, err = G.AddEdge(v2, v3, nil)
	if err != nil {
		t.Errorf("Failed to add edge 2: %v\n", err)
	}

	_, found = G.FindEdge(etemp2)
	assert.True(t, found, "edge 2 should have been added")

	_, err = G.AddEdge(v3, v1, nil)
	if err != nil {
		t.Errorf("Failed to add edge 3: %v\n", err)
	}

	etemp3, err := NewEdge(v3, v1, nil)
	if err != nil {
		t.Errorf("Failed to create edge 3: %v\n", err)
	}
	_, found = G.FindEdge(etemp3)
	assert.True(t, found, "edge 3 should have been added")

	out, err := G.DotViz()
	if err != nil {
		t.Errorf("Failed to create dotviz: %v\n", err)
	}

	log.Infof("straight line graph:\n\n%s\n", out)
}

func TestGraphDeepCopy(t *testing.T) {

	name := "test-copy"
	G := &Graph{Name: name}

	// create 3 vertex in a line
	v1 := &Vertex{Name: "1", Properties: map[string]string{"cpu": "1"}}
	v2 := &Vertex{Name: "2", Properties: map[string]string{"cpu": "2"}}
	v3 := &Vertex{Name: "3", Properties: map[string]string{"cpu": "3"}}

	err := G.AddVertexObj(v1)
	if err != nil {
		t.Errorf("Failed to add vertex 1: %v\n", err)
	}

	err = G.AddVertexObj(v2)
	if err != nil {
		t.Errorf("Failed to add vertex 2: %v\n", err)
	}

	err = G.AddVertexObj(v3)
	if err != nil {
		t.Errorf("Failed to add vertex 3: %v\n", err)
	}

	_, err = G.AddEdge(v1, v2, map[string]string{"bw": "1"})
	if err != nil {
		t.Errorf("Failed to add edge 1: %v\n", err)
	}

	_, err = G.AddEdge(v2, v3, map[string]string{"bw": "2"})
	if err != nil {
		t.Errorf("Failed to add edge 2: %v\n", err)
	}

	_, err = G.AddEdge(v3, v1, map[string]string{"bw": "3"})
	if err != nil {
		t.Errorf("Failed to add edge 3: %v\n", err)
	}

	GG, err := G.DeepCopy()
	if err != nil {
		t.Errorf("Deep copy failed\n")
	}

	b := reflect.DeepEqual(G, GG)
	if !b {
		t.Errorf("Reflect of copies not equal\n")
	}

}

func TestGraphPrune(t *testing.T) {

	name := "test-prune"
	G := &Graph{Name: name}

	// create 3 vertex in a line
	v1 := &Vertex{Name: "1", Properties: map[string]string{"cpu": "1"}}
	v2 := &Vertex{Name: "2", Properties: map[string]string{"cpu": "2"}}
	v3 := &Vertex{Name: "3", Properties: map[string]string{"cpu": "3"}}

	err := G.AddVertexObj(v1)
	if err != nil {
		t.Errorf("Failed to add vertex 1: %v\n", err)
	}

	err = G.AddVertexObj(v2)
	if err != nil {
		t.Errorf("Failed to add vertex 2: %v\n", err)
	}

	err = G.AddVertexObj(v3)
	if err != nil {
		t.Errorf("Failed to add vertex 3: %v\n", err)
	}

	_, err = G.AddEdge(v1, v2, map[string]string{"bw": "1", "selector": "a"})
	if err != nil {
		t.Errorf("Failed to add edge 1: %v\n", err)
	}

	_, err = G.AddEdge(v2, v3, map[string]string{"bw": "2", "selector": "a"})
	if err != nil {
		t.Errorf("Failed to add edge 2: %v\n", err)
	}

	GP, err := G.DeepCopy()
	if err != nil {
		t.Errorf("Failed to create deep copy: %v\n", err)
	}

	_, err = G.AddEdge(v3, v1, map[string]string{"bw": "3"})
	if err != nil {
		t.Errorf("Failed to add edge 3: %v\n", err)
	}

	gp, err := PruneGraph(G, "a")
	if err != nil {
		t.Errorf("Failed to prune graph: %v\n", err)

	}

	b := reflect.DeepEqual(gp, GP)
	if !b {
		t.Errorf("Reflect of copies not equal\n")
	}

}
