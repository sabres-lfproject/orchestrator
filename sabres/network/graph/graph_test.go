package graph

import (
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

	found := G.FindVertex(v1)
	assert.True(t, found, "vertex 1 should have been added")

	found = G.FindVertex(v2)
	assert.True(t, found, "vertex 2 should have been added")

	found = G.FindVertex(v3)
	assert.False(t, found, "vertex 3 has not been added")

	found = G.FindEdge(e1)
	assert.True(t, found, "edge 1 should have been added")

	found = G.FindEdge(e2)
	assert.False(t, found, "edge 2 has not been added")

	err = G.AddEdgeObj(e2)
	if err != nil {
		t.Errorf("Failed to add edge 2: %v\n", err)
	}

	found = G.FindVertex(v3)
	assert.True(t, found, "vertex 3 should have been added")

	found = G.FindEdge(e2)
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

	found := G.FindVertex(v1)
	assert.True(t, found, "vertex 1 should have been added")

	err = G.AddVertexObj(v2)
	if err != nil {
		t.Errorf("Failed to add vertex 2: %v\n", err)
	}

	found = G.FindVertex(v2)
	assert.True(t, found, "vertex 2 should have been added")

	err = G.AddVertexObj(v3)
	if err != nil {
		t.Errorf("Failed to add vertex 3: %v\n", err)
	}

	found = G.FindVertex(v3)
	assert.True(t, found, "vertex 3 should have been added")

	_, err = G.AddEdge(v1, v2, nil)
	if err != nil {
		t.Errorf("Failed to add edge 1: %v\n", err)
	}

	etemp1, err := NewEdge(v1, v2, nil)
	if err != nil {
		t.Errorf("Failed to create edge 1: %v\n", err)
	}

	found = G.FindEdge(etemp1)
	assert.True(t, found, "edge 1 should have been added")

	etemp2, err := NewEdge(v2, v3, nil)
	if err != nil {
		t.Errorf("Failed to create edge 1: %v\n", err)
	}

	found = G.FindEdge(etemp2)
	assert.False(t, found, "edge 2 has not been added yet")

	_, err = G.AddEdge(v2, v3, nil)
	if err != nil {
		t.Errorf("Failed to add edge 2: %v\n", err)
	}

	found = G.FindEdge(etemp2)
	assert.True(t, found, "edge 2 should have been added")

	_, err = G.AddEdge(v3, v1, nil)
	if err != nil {
		t.Errorf("Failed to add edge 3: %v\n", err)
	}

	etemp3, err := NewEdge(v3, v1, nil)
	if err != nil {
		t.Errorf("Failed to create edge 3: %v\n", err)
	}
	found = G.FindEdge(etemp3)
	assert.True(t, found, "edge 3 should have been added")

	out, err := G.DotViz()
	if err != nil {
		t.Errorf("Failed to create dotviz: %v\n", err)
	}

	log.Infof("straight line graph:\n\n%s\n", out)
}
