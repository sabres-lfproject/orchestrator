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

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gitlab.com/mergetb/tech/stor"
	"google.golang.org/grpc"
	cbspkg "pulwar.isi.edu/sabres/cbs/cbs/service/pkg"
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

		return nil
	})
	if err != nil {
		return nil, err
	}

	var cbsOut *cbspkg.JsonCBSOut

	err = netpkg.WithNetwork(netaddr, func(c protocol.NetworkClient) error {
		resp, err := c.RequestSolution(context.TODO(), &protocol.SolveRequest{
			Constraints: constraints,
		})
		if err != nil {
			return err
		}

		strCBSOut := resp.Response
		err = json.Unmarshal([]byte(strCBSOut), cbsOut)
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

	// go through inventory slices

	// print out slices as json

	return &proto.ShowSliceResponse{}, nil
}

func (s *ManagerServer) ConfigureSlice(ctx context.Context, req *proto.ConfigureSliceRequest) (*proto.ConfigureSliceResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("ConfigureSlice: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// go to inventory

	// correlate uuids with ip addresses

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
