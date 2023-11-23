package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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
		ma["cpu"] = string(io.Phy.Cores)
		ma["mem"] = string(io.Phy.Memory)
		ma["disk"] = string(io.Phy.Storage)
	} else {
		if io.Virt != nil {
			ma["cpu"] = string(io.Virt.Cores)
			ma["mem"] = string(io.Virt.Memory)
			ma["disk"] = string(io.Virt.Storage)
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

			err = addVertex(G, io.Resource)
			if err != nil {
				continue
			}

			// add vertices
			net := io.Resource.Network
			if net != nil {
				for _, link := range net.Adjlist {
					src := link.SrcResource
					dst := link.DstResource

					// check if src, dst are in the graph
					found := G.FindVertex(&graph.Vertex{Name: src})
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
					found = G.FindVertex(&graph.Vertex{Name: dst})
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

					ma["bw"] = string(link.Bandwidth)
					ma["lat"] = string(link.Latency)
					ma["jit"] = string(link.Jitter)

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
	GlobalGraph = g
	defer mutex.Unlock()

	return &proto.CreateGraphResponse{}, nil
}

func (s *NetworkServer) DeleteGraph(ctx context.Context, req *proto.DeleteGraphRequest) (*proto.DeleteGraphResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("DeleteGraph: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	mutex.Lock()
	GlobalGraph = nil
	defer mutex.Unlock()

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

func (s *NetworkServer) RequestSolution(ctx context.Context, req *proto.SolveRequest) (*proto.SolveResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Solve: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	mutex.Lock()
	defer mutex.Unlock()

	if GlobalGraph == nil {
		return nil, fmt.Errorf("graph not found, must be created first.")
	}

	// reach out to cbs for it to its thing

	return &proto.SolveResponse{}, nil
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
