package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	ipkg "pulwar.isi.edu/sabres/orchestrator/inventory/pkg"
	inventory "pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
	"pulwar.isi.edu/sabres/orchestrator/sabres/network/graph"
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

func createInventoryGraph() {
	G := &graph.Graph{}

	// read from inventory what the structure of the graph is
	// add each component to the graph
	addr := fmt.Sprintf("%s:%d", "localhost", ipkg.DefaultInventoryPort)
	err := ipkg.WithInventory(addr, func(c inventory.InventoryClient) error {
		resp, err := c.ListInventoryItems(context.TODO(),
			&inventory.ListInventoryItemsRequest{},
		)

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
					err := G.Vertex(&graph.Vertex{Name: src})
					if err == graph.ErrVertexNotFound {
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
					err = G.Vertex(&graph.Vertex{Name: dst})
					if err == graph.ErrVertexNotFound {
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
		log.Fatalf("we done goofed: %v", err)
	}

}

func main() {

	// services, take in graph
	// services, take in constraints

	// send request to cbs

	// get response from cbs

	// send reservation request to allocator

}
