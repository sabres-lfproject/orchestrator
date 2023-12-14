package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	cbs "pulwar.isi.edu/sabres/cbs/cbs/service/pkg"
	ipkg "pulwar.isi.edu/sabres/orchestrator/inventory/pkg"
	inventory "pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
	"pulwar.isi.edu/sabres/orchestrator/sabres/network/graph"
	"pulwar.isi.edu/sabres/orchestrator/sabres/network/pkg"
	proto "pulwar.isi.edu/sabres/orchestrator/sabres/network/protocol"
)

var (
	GlobalGraph *graph.Graph
	CBSHost     string
	mutex       sync.Mutex
)

// TODO: A node with physical and virtual resources overlap, so some
// work in future towards how phy resources are allocated as virt,
// and then how reservation affects both

// for now, assume either phy or virt
func addVertex(G *graph.Graph, io *inventory.ResourceItem) error {
	ma := make(map[string]string)
	if io.Phy != nil {
		ma["cpu"] = fmt.Sprintf("%d", io.Phy.Cores)
		ma["mem"] = fmt.Sprintf("%d", io.Phy.Memory)
		ma["disk"] = fmt.Sprintf("%d", io.Phy.Storage)
	} else {
		if io.Virt != nil {
			ma["cpu"] = fmt.Sprintf("%d", io.Virt.Cores)
			ma["mem"] = fmt.Sprintf("%d", io.Virt.Memory)
			ma["disk"] = fmt.Sprintf("%d", io.Virt.Storage)
		}
	}

	log.Infof("Adding vertex: %s\n", io.Uuid)

	_, err := G.AddVertex(io.Uuid, "", ma)
	if err != nil {
		if err == graph.ErrVertexAlreadyExists {
			return err
		}

		log.Errorf("Could not add %s: %#v to graph\n", io.Uuid, io.Phy)
		return err
	}

	return nil
}

func createInventoryGraph() (*graph.Graph, error) {
	G := &graph.Graph{}

	// read from inventory what the structure of the graph is
	// add each component to the graph
	addr := fmt.Sprintf("%s:%d", "localhost", ipkg.DefaultInventoryPort)
	err := ipkg.WithInventory(addr, func(c inventory.InventoryClient) error {
		resp, err := c.ListInventoryItems(context.TODO(),
			&inventory.ListInventoryItemsRequest{},
		)

		if err != nil {
			return err
		}

		log.Infof("items: %d\n", len(resp.Items))

		for _, io := range resp.Items {
			if io.Resource == nil {
				continue
			}

			// check what type of resource, is it network, node, both?

			// if it is not a network, add it a vertex in the graph
			if io.Resource.Network == nil {
				err = addVertex(G, io.Resource)
				if err != nil {
					continue
				}
			}

			// if it is a network, add the edges, and potentially the vertices of the edges.
			net := io.Resource.Network
			if net != nil {
				for _, link := range net.Adjlist {
					src := link.SrcResource
					dst := link.DstResource

					// check if src, dst are in the graph
					found, _ := G.FindVertex(&graph.Vertex{Name: src})
					if !found {
						// Look for inventory for node
						respSrc, err := c.GetResourceItem(context.TODO(),
							&inventory.GetResourceItemRequest{
								Uuid: src,
							},
						)

						if err != nil {
							log.Errorf("Add Edge: src vertex not found in inv: %s: %v\n", src, err)
							continue
						}

						rec := respSrc.Item
						if rec == nil {
							continue
						}

						err = addVertex(G, rec.Resource)
						if err != nil {
							log.Errorf("Unable to add missing vertex in edge: %s: %v\n", src, err)
						}
					}

					// check if src, dst are in the graph
					found, _ = G.FindVertex(&graph.Vertex{Name: dst})
					if !found {
						// Look for inventory for node
						respSrc, err := c.GetResourceItem(context.TODO(),
							&inventory.GetResourceItemRequest{
								Uuid: dst,
							},
						)

						if err != nil {
							log.Errorf("Add Edge: dst vertex not found in inv: %s: %v\n", dst, err)
							continue
						}

						rec := respSrc.Item
						if rec == nil {
							continue
						}

						err = addVertex(G, rec.Resource)
						if err != nil {
							log.Errorf("Unable to add missing vertex in edge: %s: %v\n", dst, err)
						}
					}

					// if not add them manually now

					ma := make(map[string]string)

					ma["bw"] = fmt.Sprintf("%d", link.Bandwidth)
					ma["lat"] = fmt.Sprintf("%d", link.Latency)
					ma["jit"] = fmt.Sprintf("%d", link.Jitter)
					ma["uuid"] = link.Uuid
					ma["name"] = net.Name
					ma["selector"] = io.Resource.Parent

					srcV := &graph.Vertex{Name: src}
					dstV := &graph.Vertex{Name: dst}

					_, err = G.AddEdge(srcV, dstV, ma)

					if err != nil {
						log.Errorf("Could not add edge %s:%s to graph\n", src, dst)
						continue
					}
				}
			}

		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return G, nil
}

type NetworkServer struct {
	proto.UnimplementedNetworkServer
}

func (s *NetworkServer) CreateGraph(ctx context.Context, req *proto.CreateGraphRequest) (*proto.CreateGraphResponse, error) {

	if req == nil {
		errMsg := fmt.Sprintf("CreateGraph: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	g, err := createInventoryGraph()
	if err != nil {
		return nil, err
	}

	mutex.Lock()
	defer mutex.Unlock()

	GlobalGraph = g

	return &proto.CreateGraphResponse{}, nil
}

func (s *NetworkServer) DeleteGraph(ctx context.Context, req *proto.DeleteGraphRequest) (*proto.DeleteGraphResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("DeleteGraph: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	mutex.Lock()
	defer mutex.Unlock()

	GlobalGraph = nil

	return &proto.DeleteGraphResponse{}, nil
}

func (s *NetworkServer) ShowGraph(ctx context.Context, req *proto.ShowGraphRequest) (*proto.ShowGraphResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("ShowGraph: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	mutex.Lock()
	defer mutex.Unlock()

	if GlobalGraph == nil {
		return &proto.ShowGraphResponse{Exists: false}, nil
	}

	dotviz, err := GlobalGraph.DotViz()
	if err != nil {
		return nil, err
	}

	return &proto.ShowGraphResponse{Exists: true, Dotviz: dotviz}, nil
}

func (s *NetworkServer) GetGraph(ctx context.Context, req *proto.GetGraphRequest) (*proto.GetGraphResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("GetGraph: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	mutex.Lock()
	defer mutex.Unlock()

	if GlobalGraph == nil {
		return nil, fmt.Errorf("graph not defined. run create first.")
	}

	jsonGraph, err := GlobalGraph.ToJson()
	if err != nil {
		return nil, err
	}

	return &proto.GetGraphResponse{Graph: jsonGraph}, nil
}

func (s *NetworkServer) RequestSolution(ctx context.Context, req *proto.SolveRequest) (*proto.SolveResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Solve: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	if req.Constraints == nil {
		return nil, fmt.Errorf("constraints not given")
	}

	if len(req.Constraints) == 0 {
		return nil, fmt.Errorf("constraints not found")
	}

	mutex.Lock()
	defer mutex.Unlock()

	if CBSHost == "" {
		return nil, fmt.Errorf("CBSHost has not been set yet, use SetCBSLocation first")
	}

	if GlobalGraph == nil {
		return nil, fmt.Errorf("graph not found, must be created first.")
	}

	gg, err := GlobalGraph.DeepCopy()
	if err != nil {
		return nil, err
	}

	c := &cbs.CBSRequest{}
	modConstraints := req.Constraints
	for _, cc := range modConstraints {
		if cc.Object == "cpu" {
			// for cbs, we need a mechanism to tell cbs that we specifically
			// want to include this node in our graph
			if len(cc.Vertices) < 1 {
				return nil, fmt.Errorf("constraint not formatted correctly")
			}
			v := cc.Vertices[0]

			ok, vertex := gg.FindVertex(&graph.Vertex{Name: v})
			if !ok {
				return nil, fmt.Errorf("couldnt find constraint vertex in graph")
			}

			vertex.Properties["endpoint"] = "yes"
			log.Infof("Updating vertex info: %v", vertex)
		}
	}

	c.Constraints = modConstraints

	// TODO: selector only to be used on edges
	// todo is to manage that for later on with more variables
	// that work on nodes,  And there can only be one selector
	selector := ""
	for _, c := range req.Constraints {
		if c.Selector != "" {
			if selector != "" && c.Selector != selector {
				return nil, fmt.Errorf("Too many edge selectors given: %s, %s. Can only use 1", selector, c.Selector)
			}
			selector = c.Selector
		}
	}

	log.Infof("selector for solving: %s\n", selector)

	log.Infof("global graph:\n")
	gg.PrintGraph()

	if selector != "" {
		g, err := graph.PruneGraph(gg, selector)
		if err != nil {
			return nil, err
		}

		k, err := g.DeepCopy()
		if err != nil {
			return nil, err
		}

		log.Infof("after prune graph:\n")
		k.PrintGraph()
		c.Graph = k

	} else {
		log.Infof("selector not set, using primary graph")
		var err error
		c.Graph, err = gg.DeepCopy()
		if err != nil {
			return nil, err
		}
	}

	// take constraints, create json
	cons, err := json.Marshal(c)
	if err != nil {
		log.Errorf("erro marshaling: %v\n", err)
		return nil, err
	}

	log.Infof("Request to send to CBS: %s\n", cons)

	//POST request to CBSHost
	endpoint := fmt.Sprintf("http://%s/cbs", CBSHost)
	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(cons))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Infof("resp from cbs: %s\n", string(body))

	// reach out to cbs for it to its thing

	// TODO: Reservation system

	return &proto.SolveResponse{Response: string(body)}, nil
}

func (s *NetworkServer) SetCBSLocation(ctx context.Context, req *proto.SetCBSRequest) (*proto.SetCBSResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Solve: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	if req.Host == "" || req.Port == "" {
		return nil, fmt.Errorf("Either Host or Port was not set: [%s:%s]", req.Host, req.Port)
	}

	mutex.Lock()
	defer mutex.Unlock()

	CBSHost = fmt.Sprintf("%s:%s", req.Host, req.Port)
	log.Infof("CBS Host set to: %s\n", CBSHost)

	return &proto.SetCBSResponse{}, nil
}

func main() {
	var debug bool
	var port int

	flag.IntVar(&port, "port", pkg.DefaultNetworkPort, "set the Networkd control port")
	flag.BoolVar(&debug, "debug", false, "enable extra debug logging")

	portStr := os.Getenv("NETWORKPORT")
	if portStr != "" {
		portInt, err := strconv.Atoi(portStr)
		if err != nil {
			log.Warningf("Failed to convert MOAPORT to int, ignored: %v", portStr)
		} else {
			port = portInt
		}
	}

	debugStr := os.Getenv("DEBUG")
	if debugStr != "" {
		debugInt, err := strconv.ParseBool(debugStr)
		if err != nil {
			log.Warningf("Failed to convert DEBUG to bool, ignored: %v", debugStr)
		} else {
			debug = debugInt
		}
	}

	// daemon mode
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Info(fmt.Sprintf("Networkd starting up on port %d", port))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterNetworkServer(grpcServer, &NetworkServer{})
	grpcServer.Serve(lis)

	// services, take in graph
	// services, take in constraints

	// send request to cbs

	// get response from cbs

	// send reservation request to allocator

}
