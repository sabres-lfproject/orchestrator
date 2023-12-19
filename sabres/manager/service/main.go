package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gitlab.com/mergetb/tech/stor"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	cbspkg "pulwar.isi.edu/sabres/cbs/cbs/service/pkg"
	inv "pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
	config "pulwar.isi.edu/sabres/orchestrator/pkg"
	"pulwar.isi.edu/sabres/orchestrator/sabres/manager/pkg"
	proto "pulwar.isi.edu/sabres/orchestrator/sabres/manager/protocol"
	netpkg "pulwar.isi.edu/sabres/orchestrator/sabres/network/pkg"
	"pulwar.isi.edu/sabres/orchestrator/sabres/network/protocol"
)

var (
	EtcdConfigPath string = "/var/orchestrator/config.cfg"
)

type ManagerServer struct {
	proto.UnimplementedManagerServer
}

func (s *ManagerServer) CreateSlice(ctx context.Context, req *proto.CreateSliceRequest) (*proto.CreateSliceResponse, error) {

	if req == nil {
		errMsg := fmt.Sprintf("CreateSlice: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	cbsaddr := fmt.Sprintf("localhost:%d", cbspkg.DefaultCBSPort)
	netaddr := fmt.Sprintf("localhost:%d", netpkg.DefaultNetworkPort)
	//invaddr := fmt.Sprintf("localhost:%d", invpkg.DefaultInventoryPort)

	// set addresses based on those passed in
	if req.CbsAddr == "" {
		cbsaddr = req.CbsAddr
	}
	if req.NetAddr == "" {
		netaddr = req.NetAddr
	}
	/*
		if req.InvAddr == "" {
			invaddr = req.InvAddr
		}
	*/

	// constraint field
	constraints := req.Constraints

	log.Infof("constraints: %+v\n", constraints)

	// TODO: check validity

	var cbsOut cbspkg.JsonCBSOut

	err := netpkg.WithNetwork(netaddr, func(c protocol.NetworkClient) error {
		cbsAddrSplit := strings.Split(cbsaddr, ":")
		if len(cbsAddrSplit) != 2 {
			return fmt.Errorf("invalid network address: %s", cbsaddr)
		}
		_, err := c.SetCBSLocation(context.TODO(), &protocol.SetCBSRequest{
			Host: cbsAddrSplit[0],
			Port: cbsAddrSplit[1],
		})
		if err != nil {
			return err
		}

		resp, err := c.RequestSolution(context.TODO(), &protocol.SolveRequest{
			Constraints: constraints,
		})
		if err != nil {
			return err
		}

		strCBSOut := resp.Response

		log.Infof("solution response from cbs: %v\n", strCBSOut)

		err = json.Unmarshal([]byte(strCBSOut), &cbsOut)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sliceUuid := uuid.New().String()

	sliceObj := &pkg.Slice{
		Name:    sliceUuid,
		Uuid:    sliceUuid,
		Devices: cbsOut.Nodes,
		Edges:   cbsOut.Edges,
	}

	log.Infof("uuid for solution: %s\n", sliceUuid)

	// send to etcd
	objs := []stor.Object{sliceObj}

	err = stor.WriteObjects(objs, true)
	if err != nil {
		return nil, err
	}

	return &proto.CreateSliceResponse{Uuid: sliceUuid}, nil
}

func (s *ManagerServer) DeleteSlice(ctx context.Context, req *proto.DeleteSliceRequest) (*proto.DeleteSliceResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("DeleteSlice: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// inventory

	// remove uuid from slice

	return &proto.DeleteSliceResponse{}, nil
}

func (s *ManagerServer) ShowSlice(ctx context.Context, req *proto.ShowSliceRequest) (*proto.ShowSliceResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("ShowSlice: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("Show Slice Request\n")

	pg := make([]*pkg.Slice, 0)

	err := stor.WithEtcd(func(c *clientv3.Client) error {
		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		resp, err := c.Get(ctx, pkg.SlicePrefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
		cancel()

		if err != nil {
			return err
		}

		log.Debugf("connected to etcd\n")

		for _, kv := range resp.Kvs {
			log.Infof("found keys: %s\n", string(kv.Key))

			keyStr := strings.Split(string(kv.Key), "/")
			if len(keyStr) < 2 {
				log.Debugf("Invalid key length: %s\n", keyStr)
				continue
			}

			so := &pkg.Slice{Uuid: keyStr[2]}
			sso := stor.Object(so)
			err = stor.Read(sso)
			if err != nil {
				return err
			}

			log.Infof("object: %v\n", so)

			pg = append(pg, so)
		}

		return nil

	})
	if err != nil {
		return nil, err
	}

	log.Infof("Slices found: %v\n", pg)

	// print out slices as json
	jsonData, err := json.Marshal(pg)
	if err != nil {
		return nil, err
	}

	return &proto.ShowSliceResponse{JsonResponse: string(jsonData)}, nil
}

func (s *ManagerServer) ConfigureSlice(ctx context.Context, req *proto.ConfigureSliceRequest) (*proto.ConfigureSliceResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("ConfigureSlice: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// validate uuid from user
	ru := req.Uuid
	_, err := uuid.Parse(ru)
	if err != nil {
		return nil, err
	}

	// go to inventory and get slice
	so := &pkg.Slice{Uuid: ru}
	sso := stor.Object(so)
	err = stor.Read(sso)
	if err != nil {
		return nil, err
	}

	if len(so.Devices) <= 0 {
		errMsg := fmt.Errorf("There are no devices in this slice")
		log.Errorf("%v\n", errMsg)
		return nil, errMsg
	}

	if len(so.Edges) <= 0 {
		errMsg := fmt.Errorf("There are no edges in this slice")
		log.Errorf("%v\n", errMsg)
		return nil, errMsg
	}

	mgmt := map[string]string{}

	// for each uuid object, get management ip
	for _, edge := range so.Edges {
		for k, v := range edge {

			// if the edge property is either a src or dst node, check if we have an ip
			if k == "src" || k == "dst" {
				_, ok := mgmt[v]
				if !ok {
					// if the uuid is not in mgmt, go get the ip address
					// TODO: we hacked before that uuid is also the same for inventory
					// so we know we can use the uuid to get the inventory item
					io := &inv.InventoryItem{Uuid: v}
					so := stor.Object(io)
					err = stor.Read(so)
					if err != nil {
						return nil, err
					}

					if io.Entity == nil {
						errMsg := fmt.Errorf("Inventory object has nil entity")
						log.Errorf("%v\n", errMsg)
						return nil, errMsg
					}

					if io.Entity.Idtype == inv.Entity_IP {

						// save the ip now in the mgmt map
						mgmt[v] = io.Entity.Identification

					} else {
						errMsg := fmt.Errorf("Inventory object has unknown entity: %#v", io.Entity)
						log.Errorf("%v\n", errMsg)
						return nil, errMsg
					}
				}
			}
		}
	}

	log.Infof("all our edge management ips found: %v\n", mgmt)

	// create the route path
	path, err := pkg.CreatePath(so.Edges)
	if err != nil {
		log.Errorf("create path failure: %v\n", err)
		return nil, err
	}

	log.Infof("path through slice: %v\n", path)

	// assign data plane network ips

	// generate configuration script

	// run configuration script

	// save to inventory (?)

	return &proto.ConfigureSliceResponse{}, nil
}

func main() {
	var debug bool
	var port int

	flag.IntVar(&port, "port", pkg.DefaultManagementPort, "set the Managerd control port")
	flag.BoolVar(&debug, "debug", false, "enable extra debug logging")

	portStr := os.Getenv("MANAGERPORT")
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

	log.Info(fmt.Sprintf("Manager starting up on port %d", port))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterManagerServer(grpcServer, &ManagerServer{})
	grpcServer.Serve(lis)

}
