package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mergetb/tech/stor"
	clientv3 "go.etcd.io/etcd/client/v3"
	proto "pulwar.isi.edu/sabres/orchestrator/discovery/protocol"

	ipkg "pulwar.isi.edu/sabres/orchestrator/inventory/pkg"
	"pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
	inventory "pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
	config "pulwar.isi.edu/sabres/orchestrator/pkg"
)

var (
	EtcdConfigPath string        = "/var/orchestrator/config.cfg"
	EndpointConfig string        = "/var/orchestrator/endpoints.cfg"
	ScanLatency    time.Duration = 60 * time.Second
)

func main() {

	var debug bool
	//var port int

	//flag.IntVar(&port, "port", pkg.DefaultScanDiscoveryPort, "set the Inventoryd control port")
	flag.BoolVar(&debug, "debug", false, "enable extra debug logging")

	flag.Parse()

	/*
		portStr := os.Getenv("SCANDISCOVERYPORT")
		if portStr != "" {
			portInt, err := strconv.Atoi(portStr)
			if err != nil {
				log.Warningf("Failed to convert DISCOVERYPORT to int, ignored: %v", portStr)
			} else {
				port = portInt
			}
		}
	*/

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

	cfg, err := config.LoadConfig(EtcdConfigPath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// read in environment variables for container
	err = config.ReadENVSettings(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}

	etcdCfg, err := config.SetEtcdSettings(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}

	stor.SetConfig(*etcdCfg)

	//log.Info(fmt.Sprintf("Discoveryd starting up on port %d", port))
	log.Infof("db settings: %+v\n", etcdCfg)

	log.Infof("Begining probing\n")

	// service loop
	for true {

		log.Debugf("Debug: Scanning\n")

		// for each host found in etcd
		eps, err := getEndpoints()
		if err != nil {
			log.Errorf("Error in get Endpoints: %v\n", err)
			time.Sleep(ScanLatency)
			continue
		}
		for _, ep := range eps {
			log.Debugf("Endpoint found: %#v\n", ep)
			err = scanEndpoint(ep)
			if err != nil {
				log.Errorf("%v\n", err)
			}
		}
		time.Sleep(ScanLatency)
	}
}

func getEndpoints() ([]*proto.Endpoint, error) {

	prefix := proto.EndpointPrefix
	endpoints := make([]*proto.Endpoint, 0)

	err := stor.WithEtcd(func(c *clientv3.Client) error {

		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		resp, err := c.Get(ctx, prefix, clientv3.WithPrefix())

		cancel()
		if err != nil {
			return err
		}

		for _, kv := range resp.Kvs {
			tmp := make(map[string]interface{})
			err = json.Unmarshal(kv.Value, &tmp)
			if err != nil {
				return err
			}

			intVer, err := strconv.Atoi(string(tmp["version"].(string)))
			if err != nil {
				log.Errorf("Failure to parse version number: %v\n", err)
				continue
			}

			tmp["version"] = int64(intVer)
			jsonString, _ := json.Marshal(tmp)

			ep := &proto.Endpoint{}
			err = json.Unmarshal(jsonString, ep)
			if err != nil {
				return err
			}

			endpoints = append(endpoints, ep)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return endpoints, nil
}

func scanEndpoint(ep *proto.Endpoint) error {
	// scan the endpoint for resources

	if ep.Uri == "" {
		log.Warnf("endpoint: %s missing URI", ep.Key())
		return fmt.Errorf("Missing URI endpoint for %s", ep.Key())
	}

	resp, err := http.Get(fmt.Sprintf("http://%s/resources", ep.Uri))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Debugf("Value: %#v\n", string(body))

	// this will be a hack, but we need a format, so assume its in
	// resource format

	recList := make([]inventory.ResourceItem, 0)
	err = json.Unmarshal(body, &recList)
	if err != nil {
		log.Warnf("Couldnt create resource: %v\n", err)
		return err
	}

	// if new resources, add back to inventory
	// TODO: make customizable endpoints for inventory service
	addr := fmt.Sprintf("%s:%d", "localhost", ipkg.DefaultInventoryPort)
	err = ipkg.WithInventory(addr, func(c protocol.InventoryClient) error {
		// check if inventory item exists- this is a bad approach: TODO
		resp, err := c.ListInventoryItems(context.TODO(),
			&protocol.ListInventoryItemsRequest{})
		if err != nil {
			return nil
		}

		resourceList := make([]inventory.ResourceItem, 0)

		// for inventory items, check each resource item (bad)
		for _, rec := range recList {

			recFound := false

			log.Debugf("resource: %#v\n", rec)

			for _, io := range resp.Items {
				if io.Resource == nil {
					continue
				}

				/*
					// TODO: right now we are going to assume that uuids
					// are assigned and managed out of band

					// convert to map
					//ir := structs.Map(io.Resource)
					//or := structs.Map(rec)

					delete(ir, "Uuid")
					delete(ir, "Parent")
					delete(or, "Uuid")
					delete(or, "Parent")
				*/

				log.Debugf("ir: %#v\n", io.Resource)
				log.Debugf("or: %#v\n", rec)

				if io.Resource.Uuid == rec.Uuid {
					log.Debugf("Resource already in inventory\n")
					recFound = true
					break
				}
			}

			if !recFound {
				log.Infof("Resource not found in inventory, requesting add\n")
				resourceList = append(resourceList, rec)
			}
		}

		for _, rec := range resourceList {

			log.Debugf("Sending create request for: %v\n", rec)
			// add new inventory item
			req := &inventory.CreateInventoryItemRequest{
				Request: &inventory.InventoryItem{
					Uuid:     rec.Uuid, // TODO: assumes resource has a uuid
					Resource: &rec,
					Entity: &inventory.Entity{
						Idtype: inventory.Entity_IP,
					},
					Notes: fmt.Sprintf("scanned by: %v", ep),
				},
			}
			//fmt.Printf("sent request to inventory: %v\n", req)

			cii_resp, err := c.CreateInventoryItem(context.TODO(), req)
			if err != nil {
				log.Errorf("Failed to add inventory item: %s\n", rec.Uuid)
				continue
			}

			fmt.Printf("response from inventory: %v\n", cii_resp)
		}

		return nil

	})

	return err

	// if removed resources, remove from inventory
	// TODO
}
