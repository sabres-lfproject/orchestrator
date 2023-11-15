package main

import (
	// standard

	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	// merge deps
	"gitlab.com/mergetb/tech/stor"
	pkg "pulwar.isi.edu/sabres/orchestrator/discovery/pkg"
	proto "pulwar.isi.edu/sabres/orchestrator/discovery/protocol"
	config "pulwar.isi.edu/sabres/orchestrator/pkg"

	// deps
	uuid "github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

var (
	EtcdConfigPath string = "/var/orchestrator/config.cfg"
)

type DiscoveryServer struct {
	proto.UnimplementedDiscoveryServer
}

func checkRequest(ii *proto.Endpoint) error {

	if ii.Services == nil {
		errMsg := fmt.Sprintf("Check Request: Empty Endpoint")
		log.Errorf("%s", errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	if ii.Services.Name == "" {
		errMsg := fmt.Sprintf("Check Request: Endpoint with no name")
		log.Errorf("%s", errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

func checkUuid(u string) error {
	_, err := uuid.Parse(u)
	return err
}

func (s *DiscoveryServer) CreateDEP(ctx context.Context, req *proto.CreateDEPRequest) (*proto.CRUDDEPResponse, error) {

	if req == nil {
		errMsg := fmt.Sprintf("CreateDEP: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	if req.Endpoint == nil {
		errMsg := fmt.Sprintf("CreateDEP: Empty Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	err := checkRequest(req.Endpoint)
	if err != nil {
		return nil, err
	}

	ep := req.Endpoint
	ep.Services.Uuid = uuid.New().String()

	log.WithFields(log.Fields{
		"Name":     ep.Services.Name,
		"Endpoint": ep.Uri,
	}).Info("CreateDEP")

	objs := []stor.Object{ep}

	err = stor.WriteObjects(objs, true)
	if err != nil {
		return nil, err
	}

	return &proto.CRUDDEPResponse{Uuid: ep.Services.Uuid}, nil
}

func (s *DiscoveryServer) ModifyDEP(ctx context.Context, req *proto.ModifyDEPRequest) (*proto.CRUDDEPResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("ModifyDEP: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	if req.Endpoint == nil {
		errMsg := fmt.Sprintf("ModifyDEP: Empty Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	err := checkRequest(req.Endpoint)
	if err != nil {
		return nil, err
	}

	rid := req.Uuid

	log.WithFields(log.Fields{
		"Uuid":     rid,
		"Endpoint": req.Endpoint,
	}).Info("ModifyDEP")

	err = checkUuid(rid)
	if err != nil {
		return nil, err
	}

	ep := req.Endpoint
	err = checkRequest(ep)
	if err != nil {
		return nil, err
	}
	ep.Services.Uuid = rid

	objs := []stor.Object{ep}

	err = stor.WriteObjects(objs, true)
	if err != nil {
		return nil, err
	}

	return &proto.CRUDDEPResponse{Uuid: rid}, nil
}

func (s *DiscoveryServer) DeleteDEP(ctx context.Context, req *proto.DeleteDEPRequest) (*proto.CRUDDEPResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("protoDeleteprotoDEP: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.WithFields(log.Fields{
		"Name": req.Uuid,
	}).Info("DeleteprotoDEP")

	rid := req.Uuid
	err := checkUuid(rid)
	if err != nil {
		return nil, err
	}

	ro := &proto.Endpoint{
		Services: &proto.Service{
			Uuid: rid,
		},
	}

	objs := []stor.Object{ro}

	err = stor.DeleteObjects(objs)
	if err != nil {
		return nil, err
	}

	return &proto.CRUDDEPResponse{Uuid: rid}, nil
}

func (s *DiscoveryServer) ListDEPs(ctx context.Context, req *proto.ListDEPRequest) (*proto.ListDEPResponse, error) {

	log.Info("List Discovery Items")

	prefix := fmt.Sprintf("%s/", proto.EndpointPrefix)

	invItems := make(map[string]string)
	err := stor.WithEtcd(func(c *clientv3.Client) error {

		// arbitrary 3 second delay, should use config value
		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		resp, err := c.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
		cancel()
		if err != nil {
			return err
		}

		for _, kv := range resp.Kvs {
			keyStr := strings.Split(string(kv.Key), "/")
			if len(keyStr) < 2 {
				log.Warnf("discovery key issue: %s\n", keyStr)
				continue
			}
			ioUuid := keyStr[2]
			invItems[ioUuid] = ""
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	iiList := make([]*proto.Endpoint, 0)
	for key, _ := range invItems {
		ii := &proto.Endpoint{
			Services: &proto.Service{
				Uuid: key,
			},
		}
		so := stor.Object(ii)
		err := stor.Read(so)
		if err != nil {
			return nil, err
		}

		iiList = append(iiList, ii)
	}

	return &proto.ListDEPResponse{
		Endpoints: iiList,
	}, nil
}

func (s *DiscoveryServer) GetDEP(ctx context.Context, req *proto.GetDEPRequest) (*proto.GetDEPResponse, error) {
	log.Info("Get Discovery Item")

	uuidReq := req.Uuid

	err := checkUuid(uuidReq)
	if err != nil {
		return nil, err
	}

	ii := &proto.Endpoint{
		Services: &proto.Service{
			Uuid: uuidReq,
		},
	}
	so := stor.Object(ii)
	err = stor.Read(so)
	if err != nil {
		return nil, err
	}

	return &proto.GetDEPResponse{
		Endpoint: ii,
	}, nil
}

func main() {

	var debug bool
	var port int

	flag.IntVar(&port, "port", pkg.DefaultDiscoveryPort, "set the Discoveryd control port")
	flag.BoolVar(&debug, "debug", false, "enable extra debug logging")

	portStr := os.Getenv("MOAPORT")
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

	log.Info(fmt.Sprintf("Discoveryd starting up on port %d", port))
	log.Infof("db settings: %+v\n", etcdCfg)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterDiscoveryServer(grpcServer, &DiscoveryServer{})
	grpcServer.Serve(lis)
}
