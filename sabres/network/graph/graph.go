package graph

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	log "github.com/sirupsen/logrus"
)

// really disliked https://github.com/dominikbraun/graph/blob/main/graph.go
// graph implementation, so just rewriting the functions with simple
// implementation.  Original code under Apache 2 license.
// https://github.com/dominikbraun/graph/blob/main/LICENSE
var (
	ErrVertexNotFound      = errors.New("vertex not found")
	ErrVertexAlreadyExists = errors.New("vertex already exists")
	ErrEdgeNotFound        = errors.New("edge not found")
	ErrEdgeAlreadyExists   = errors.New("edge already exists")
	ErrEdgeVertsNotFound   = errors.New("edge vertices not found")
)

type Vertex struct {
	Name       string            `yaml:"name" json:"name" binding:"required"`
	Value      string            `yaml:"value" json:"value"`
	Properties map[string]string `yaml:"properties" json:"properties"`
	Weight     int               `yaml:"weight" json:"weight"`
}

type Edge struct {
	Name       string            `yaml:"name" json:"name" binding:"required"`
	Vertices   []*Vertex         `yaml:"vertices" json:"vertices" binding:"required"`
	Properties map[string]string `yaml:"properties" json:"properties"`
	Weight     int               `yaml:"weight" json:"weight"`
}

type Graph struct {
	Name     string    `yaml:"name" json:"name" binding:"required"`
	Vertices []*Vertex `yaml:"vertices" json:"vertices"`
	Edges    []*Edge   `yaml:"edges" json:"edges"`
}

func (g *Graph) FindEdge(e *Edge) ([]*Edge, bool) {
	eList := make([]*Edge, 0)
	for _, ge := range g.Edges {
		m1 := make(map[string]string)
		for _, v1 := range e.Vertices {
			m1[v1.Name] = ""
		}
		m2 := make(map[string]string)
		for _, v2 := range ge.Vertices {
			m2[v2.Name] = ""
		}

		// check if edge by name already exists
		if reflect.DeepEqual(m1, m2) {
			eList = append(eList, ge)
		}
	}

	if len(eList) > 0 {
		return eList, true
	}

	// edge not found
	return nil, false
}

func (g *Graph) AddEdge(v1, v2 *Vertex, prop map[string]string) (*Edge, error) {
	e, err := NewEdge(v1, v2, prop)
	if err != nil {
		return nil, err
	}

	found, _ := g.FindVertex(v1)
	if !found {
		log.Debugf("adding vertex 1: %v\n", !found)
		// adding vertex
		err = g.AddVertexObj(v1)
		// should not return error because we already checked it didnt exist
		if err != nil {
			return nil, err
		}
	}

	found, _ = g.FindVertex(v2)
	if !found {
		log.Debugf("adding vertex 2: %v\n", !found)
		// adding vertex
		err = g.AddVertexObj(v2)
		// should not return error because we already checked it didnt exist
		if err != nil {
			return nil, err
		}
	}

	if g.Edges != nil {
		eList, found := g.FindEdge(e)
		uuid, ok := prop["uuid"]
		if !ok && found {
			return nil, ErrEdgeAlreadyExists
		}
		if found {
			for _, ee := range eList {
				if ee.Properties != nil {
					propUuid, ok := ee.Properties["uuid"]
					if ok {
						if propUuid == uuid {
							return nil, ErrEdgeAlreadyExists
						} else {
							log.Debugf("edge with same vertices, found, different network. adding: %v\n", e)
						}
					}
				}
			}
		} else {
			log.Debugf("adding edge: %v\n", e)
		}
	} else {
		g.Edges = make([]*Edge, 0)
	}

	g.Edges = append(g.Edges, e)

	return e, nil
}

func (g *Graph) AddEdgeObj(edge *Edge) error {
	if edge.Vertices != nil {
		if len(edge.Vertices) == 2 {
			e, err := g.AddEdge(edge.Vertices[0], edge.Vertices[1], edge.Properties)
			log.Debugf("edge add: %v\n", e)
			log.Debugf("edges: %#v\n", g.Edges)
			return err
		} else {
			return errors.New("AddEdge requires only 2 vertices")
		}
	} else {
		return ErrEdgeVertsNotFound
	}
}

func (g *Graph) FindVertex(v *Vertex) (bool, *Vertex) {
	for _, gv := range g.Vertices {
		if gv.Name == v.Name {
			// vertex found
			return true, gv
		}
	}

	// vertex not found
	return false, nil
}

func (g *Graph) AddVertex(name, value string, prop map[string]string) (*Vertex, error) {
	v := &Vertex{
		Name:       name,
		Value:      value,
		Properties: prop,
	}

	if g.Vertices != nil {
		found, _ := g.FindVertex(v)
		if found {
			return nil, ErrVertexAlreadyExists
		}
	} else {
		g.Vertices = make([]*Vertex, 0)
	}

	g.Vertices = append(g.Vertices, v)

	return v, nil
}

func (g *Graph) AddVertexObj(v *Vertex) error {
	_, err := g.AddVertex(v.Name, v.Value, v.Properties)
	return err
}

func NewEdge(v1, v2 *Vertex, prop map[string]string) (*Edge, error) {
	if v1 == nil || v2 == nil {
		return nil, ErrVertexNotFound
	}

	if v1.Name == "" || v2.Name == "" {
		return nil, errors.New("vertex missing name field")
	}

	if v1.Name == v2.Name {
		return nil, errors.New("vertices share the same name")
	}

	return &Edge{
		Name:       fmt.Sprintf("%s-%s", v1.Name, v2.Name),
		Vertices:   []*Vertex{v1, v2},
		Properties: prop,
	}, nil
}

func (g *Graph) DotViz() (string, error) {
	gviz := graphviz.New()
	gvizObj, err := gviz.Graph()
	if err != nil {
		return "", err
	}
	defer func() error {
		if err := gvizObj.Close(); err != nil {
			return err
		}

		return gviz.Close()
	}()

	log.Debugf("vertices: %#v\n", g.Vertices)

	m := make(map[string]*cgraph.Node)
	for _, v := range g.Vertices {
		vtemp, err := gvizObj.CreateNode(v.Name)
		if err != nil {
			return "", err
		}
		vtemp.SetLabel(v.Name)
		m[v.Name] = vtemp
	}

	log.Debugf("vertices: %#v\n", m)

	labelMap := make(map[string]string)
	edgeMap := make(map[string]*cgraph.Edge)
	for _, e := range g.Edges {
		for i, v := range e.Vertices {
			log.Debugf("vert %d: %s", i, v.Name)
		}
		if e.Vertices != nil {
			if len(e.Vertices) == 2 {
				v1, ok := m[e.Vertices[0].Name]
				if !ok {
					return "", errors.New(fmt.Sprintf("Vertex not found in graph: %s", e.Vertices[0].Name))
				}
				v2, ok := m[e.Vertices[1].Name]
				if !ok {
					return "", errors.New(fmt.Sprintf("Vertex not found in graph: %s", e.Vertices[1].Name))
				}

				etemp, err := gvizObj.CreateEdge(e.Name, v1, v2)
				if err != nil {
					return "", err
				}
				edgeMap[e.Name] = etemp

				bw, ok := e.Properties["bw"]
				netuid, ok2 := e.Properties["uuid"]

				if ok && ok2 {
					label := fmt.Sprintf("network: %s | bandwidth: %d", netuid, bw)
					val, ok := labelMap[e.Name]
					if !ok {
						labelMap[e.Name] = label
					} else {
						label = fmt.Sprintf("%s || network: %s | bandwidth: %d", val, netuid, bw)
						labelMap[e.Name] = label
					}
					//etemp.SetLabel(label)
				} else {
					val, ok := labelMap[e.Name]
					if !ok {
						//etemp.SetLabel(e.Name)
						labelMap[e.Name] = e.Name
					} else {
						if val != e.Name {
							label := fmt.Sprintf("%s || %s ", val, e.Name)
							labelMap[e.Name] = label
						}
					}
				}

			}
		}
	}

	for name, edge := range edgeMap {
		val, ok := labelMap[name]
		if ok {
			edge.SetLabel(val)
		}
	}

	var buf bytes.Buffer
	if err := gviz.Render(gvizObj, "dot", &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
