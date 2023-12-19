package pkg

import (
	"fmt"
	"net"
	"sort"

	log "github.com/sirupsen/logrus"
	graph "pulwar.isi.edu/sabres/orchestrator/sabres/network/graph"
)

// IPAssignment will do the heavy lifting for static ip routing
func IPAssignment(path []string, counter int) (map[string]net.IP, map[string][]net.IPNet, int, error) {

	// first and last nodes in path will only have a single ip address
	totalNetworks := len(path) - 1

	ipMap := make(map[string][]net.IP)
	ipRoute := make(map[string][]net.IPNet)

	for i, n := range path {
		// last node
		if i == len(path) {
		}

		// otherwise look at both nodes
		thisIP := fmt.Sprintf("10.10.%d.%d/24", counter)

	}
}

func WalkManager(source, target string, G *graph.Graph) ([]string, error) {

	if source == target {
		return []string{source}, nil
	}

	v1 := &graph.Vertex{Name: source}
	v2 := &graph.Vertex{Name: target}

	found, _ := G.FindVertex(v1)
	if !found {
		return nil, fmt.Errorf("source not in graph: %s\n", source)
	}

	found, _ = G.FindVertex(v2)
	if !found {
		return nil, fmt.Errorf("target not in graph: %s\n", source)
	}

	path := []string{source}

	only, neigh, err := NeighborWalk("", source, G)
	if err != nil {
		return nil, err
	}

	path = append(path, neigh)

	prev := source
	current := neigh

	for {

		only, neigh, err = NeighborWalk(prev, current, G)
		if err != nil {
			return nil, err
		}

		path = append(path, neigh)
		if only {
			return path, nil
		}

		temp := current
		current = neigh
		prev = temp

		// again this shouldnt happen, but because this is linked list it
		// should also hit this point
		if current == target {
			return path, nil
		}

		log.Infof("Path: %v\n", path)

		// because this is a linked list, we cant have more edges than nodes
		// so this will be the trigger
		if len(path) >= len(G.Vertices) {
			errMsg := fmt.Errorf("Going infinite: %s", path)
			log.Errorf("%v\n", errMsg)
			return nil, errMsg
		}

	}

	return nil, nil

}

// TODO: move to graph lib. assumes single neighbor
func NeighborWalk(prev, current string, G *graph.Graph) (bool, string, error) {

	only := true

	for _, edge := range G.Edges {
		if len(edge.Vertices) < 2 {
			return false, "", fmt.Errorf("edge had less than 2 vertices")
		}
		vX := edge.Vertices[0]
		vY := edge.Vertices[1]

		// if we dont have a previous, we are at an endpoint
		if prev == "" {
			if vX.Name == current {
				return only, vY.Name, nil
			}
			if vY.Name == current {
				return only, vX.Name, nil
			}
		} else {
			// otherwise we need to make sure this isnt the same edge
			// we passed in
			if vX.Name == current && vY.Name != prev {
				only = false
				return only, vY.Name, nil
			}
			if vY.Name == current && vX.Name != prev {
				only = false
				return only, vX.Name, nil
			}
		}
	}

	return only, "", nil
}

func CreatePath(edgeMap []map[string]string) ([]string, error) {

	counter := map[string]int{}

	G := &graph.Graph{Name: "path-select"}

	for _, edge := range edgeMap {
		v1 := &graph.Vertex{Name: edge["src"]}
		v2 := &graph.Vertex{Name: edge["dst"]}
		names := []string{v1.Name, v2.Name}
		sort.Strings(names)
		e1 := &graph.Edge{
			Name:     fmt.Sprintf("%s-%s", names[0], names[1]),
			Vertices: []*graph.Vertex{v1, v2},
		}

		err := G.AddEdgeObj(e1)
		if err != nil {
			log.Errorf("Failed to add edge 1: %v\n", err)
			return nil, err
		}

		_, ok := counter[v1.Name]
		if ok {
			counter[v1.Name]++
		} else {
			counter[v1.Name] = 1
		}

		_, ok2 := counter[v2.Name]
		if ok2 {
			counter[v2.Name]++
		} else {
			counter[v2.Name] = 1
		}

	}

	// now we should know a source sink from counter map
	var s, t string
	set := false
	for k, v := range counter {
		if v == 1 {
			if !set {
				s = k
				set = true
			} else {
				t = k
			}
		}
	}

	log.Infof("found: %s <--> %s\n", s, t)

	// then we walk the graph
	path, err := WalkManager(s, t, G)
	if err != nil {
		return nil, err
	}

	return path, nil
}
