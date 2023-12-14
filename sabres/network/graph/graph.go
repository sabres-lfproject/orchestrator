package graph

import (
	"bytes"
	"encoding/json"
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

func FromJson(bstring []byte) (*Graph, error) {

	G := &Graph{}
	err := json.Unmarshal(bstring, G)
	if err != nil {
		return nil, err
	}

	return G, nil
}

func (g *Graph) ToJson() (string, error) {
	jsonG, err := json.Marshal(g)
	return string(jsonG), err
}

func (g *Graph) PrintGraph() {
	fmt.Printf("Graph: %s\n", g.Name)
	if len(g.Vertices) <= 0 {
		fmt.Printf("is empty\n")
		return
	} else {
		fmt.Printf("Vertices:\n")
		for _, v := range g.Vertices {
			fmt.Printf("\t%s: [%#v]\n", v.Name, v.Properties)
		}
	}

	if len(g.Edges) > 0 {
		fmt.Printf("Edges:\n")
		for _, e := range g.Edges {
			if len(e.Vertices) <= 1 {
				continue
			}

			fmt.Printf("\t%s->%s: [%#v]\n", e.Vertices[0].Name, e.Vertices[1].Name, e.Properties)
		}
	} else {
		fmt.Printf("No Edges.\n")
	}
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
					label := fmt.Sprintf("network: %s | bandwidth: %s", netuid, bw)
					val, ok := labelMap[e.Name]
					if !ok {
						labelMap[e.Name] = label
					} else {
						label = fmt.Sprintf("%s || network: %s | bandwidth: %s", val, netuid, bw)
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

// DeepCopy function
func (g *Graph) DeepCopy() (*Graph, error) {
	gJson, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}

	ng := Graph{}
	err = json.Unmarshal(gJson, &ng)
	if err != nil {
		return nil, err
	}

	return &ng, nil
}

// DeleteEdge function only to be used for cbs which will use
// a property value to differentiate edges
func (g *Graph) DeleteEdge(e *Edge) error {
	if g == nil {
		return fmt.Errorf("delete edge called on nil graph")
	}

	if g.Edges == nil {
		return fmt.Errorf("delete edge called on graph with nil edges")
	}

	if len(g.Edges) == 0 {
		return fmt.Errorf("delete edge called on graph without edges")
	}

	eList := make([]*Edge, 0)
	for _, ge := range g.Edges {
		//log.Debugf("edge: %v\n", ge)
		if ge.Name != e.Name {
			eList = append(eList, ge)
		} else {
			eprops := e.Properties
			if eprops == nil {
				return fmt.Errorf("edge needs properties for DeleteEdge")
			}
			geprops := ge.Properties
			if geprops == nil {
				log.Debugf("Missing props. Deleted edge: %v from graph\n", ge)
				continue
			}

			if eprops["selector"] == geprops["selector"] {
				log.Infof("Found Edge. Deleted edge: %v from graph\n", ge)
				continue
			} else {
				eList = append(eList, ge)
			}
		}
	}

	g.Edges = eList

	return nil
}

// TODO: this is a function only for pruning edges based on a selector
func PruneGraph(g *Graph, sel string) (*Graph, error) {
	if sel == "" {
		log.Warnf("Prune called without selector value")
		return g, nil
	}

	if g == nil {
		return nil, fmt.Errorf("Prune called with nil graph")
	}

	if len(g.Edges) == 0 {
		log.Warnf("Prune called with graph with no edges")
		return g, nil
	}

	log.Infof("prune called with selector: %s\n", sel)

	newGraph, err := g.DeepCopy()
	if err != nil {
		return nil, err
	}

	eList := make([]*Edge, len(newGraph.Edges))
	copy(eList, newGraph.Edges)

	for _, e := range eList {
		if e.Properties != nil {
			//log.Infof("Prune, edge: %s [%s] ? %s", e.Name, e.Properties["selector"], sel)
			val, ok := e.Properties["selector"]
			//log.Infof("val: %s ok: [%v]", val, ok)
			if ok {
				if val != sel {
					//log.Infof("Deleting Edge- not selected: %v\n", e)

					//log.Infof("before delete edge")
					//newGraph.PrintGraph()

					err = newGraph.DeleteEdge(e)
					if err != nil {
						return nil, err
					}

					//log.Infof("adter delete edge")
					//newGraph.PrintGraph()

				} else {
					log.Infof("Edge selected: %v\n", e)
				}
			} else {
				log.Infof("Deleting Edge- no selector: %v\n", e)
				err = newGraph.DeleteEdge(e)
				if err != nil {
					return nil, err
				}
			}
		} else {
			log.Infof("Deleting Edge- no properties to select: %v\n", e)
			err = newGraph.DeleteEdge(e)
			if err != nil {
				return nil, err
			}
		}
	}

	return newGraph, nil
}
