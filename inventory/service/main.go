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
	pkg "pulwar.isi.edu/sabres/orchestrator/inventory/pkg"
	inv "pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
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

type InventoryServer struct {
	inv.UnimplementedInventoryServer
}

func checkRequest(ii *inv.InventoryItem) error {

	if ii.Resource == nil {
		errMsg := fmt.Sprintf("Check Request: Empty Resource")
		log.Errorf("%s", errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	if ii.Resource.Name == "" {
		errMsg := fmt.Sprintf("Check Request: Resource with no name")
		log.Errorf("%s", errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

func checkUuid(u string) error {
	_, err := uuid.Parse(u)
	return err
}

func (s *InventoryServer) CreateInventoryItem(ctx context.Context, req *inv.CreateInventoryItemRequest) (*inv.InventoryItemResponse, error) {

	var ro *inv.ResourceItem
	var io *inv.InventoryItem

	if req == nil {
		errMsg := fmt.Sprintf("inv.Createinv.InventoryItem: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	if req.Request == nil {
		errMsg := fmt.Sprintf("inv.Createinv.InventoryItem: Empty Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	err := checkRequest(req.Request)
	if err != nil {
		return nil, err
	}

	// for now - with mocked data, especially resources, let the uuid be
	// passed in, otherwise we need a mechanism to later attribute names
	// to uuids due to adding edges or other referencial resources
	io = req.Request
	if err = checkUuid(io.Uuid); err != nil {
		io.Uuid = uuid.New().String()
	}
	ro = req.Request.Resource
	if err = checkUuid(ro.Uuid); err != nil {
		ro.Uuid = uuid.New().String()
	}
	ro.Parent = io.Uuid
	io.Resource = ro

	log.WithFields(log.Fields{
		"Name":     io.Uuid,
		"Resource": ro.Uuid,
	}).Info("inv.Createinv.InventoryItem")

	objs := []stor.Object{io, ro}

	err = stor.WriteObjects(objs, true)
	if err != nil {
		return nil, err
	}

	return &inv.InventoryItemResponse{IoUuid: io.Uuid, RoUuid: ro.Uuid}, nil
}

func (s *InventoryServer) ModifyInventoryItem(ctx context.Context, req *inv.ModifyInventoryItemRequest) (*inv.InventoryItemResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("inv.Modifyinv.InventoryItem: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	if req.Request == nil {
		errMsg := fmt.Sprintf("inv.Modifyinv.InventoryItem: Empty Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	err := checkRequest(req.Request)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"Name":     req.Request.Resource.Name,
		"Resource": req.Request.Resource,
	}).Info("inv.Modifyinv.InventoryItem")

	io := req.Request
	err = checkUuid(io.Uuid)
	if err != nil {
		return nil, err
	}

	ro := req.Request.Resource
	err = checkUuid(ro.Uuid)
	if err != nil {
		return nil, err
	}

	objs := []stor.Object{io, ro}

	err = stor.WriteObjects(objs, true)
	if err != nil {
		return nil, err
	}

	return &inv.InventoryItemResponse{IoUuid: io.Uuid, RoUuid: ro.Uuid}, nil
}

func (s *InventoryServer) DeleteInventoryItem(ctx context.Context, req *inv.DeleteInventoryItemRequest) (*inv.InventoryItemResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("inv.Deleteinv.InventoryItem: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	if req.Request == nil {
		errMsg := fmt.Sprintf("inv.Deleteinv.InventoryItem: Empty Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	err := checkRequest(req.Request)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"Name":     req.Request.Resource.Name,
		"Resource": req.Request.Resource,
	}).Info("inv.Deleteinv.InventoryItem")

	io := req.Request
	err = checkUuid(io.Uuid)
	if err != nil {
		return nil, err
	}

	ro := req.Request.Resource
	err = checkUuid(ro.Uuid)
	if err != nil {
		return nil, err
	}

	objs := []stor.Object{io, ro}

	err = stor.DeleteObjects(objs)
	if err != nil {
		return nil, err
	}

	return &inv.InventoryItemResponse{IoUuid: io.Uuid, RoUuid: ro.Uuid}, nil
}

func (s *InventoryServer) ListInventoryItems(ctx context.Context, req *inv.ListInventoryItemsRequest) (*inv.ListInventoryItemsResponse, error) {

	log.Info("List Inventory Items")

	prefix := fmt.Sprintf("%s/", inv.InvPrefix)

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
				log.Warnf("inventory key issue: %s\n", keyStr)
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

	iiList := make([]*inv.InventoryItem, 0)
	for k, _ := range invItems {
		ii := &inv.InventoryItem{Uuid: k}
		so := stor.Object(ii)
		err := stor.Read(so)
		if err != nil {
			return nil, err
		}

		iiList = append(iiList, ii)
	}

	return &inv.ListInventoryItemsResponse{
		Items: iiList,
	}, nil

}

func (s *InventoryServer) GetInventoryItem(ctx context.Context, req *inv.GetInventoryItemRequest) (*inv.GetItemResponse, error) {
	log.Info("Get Inventory Item")

	uuidReq := req.Uuid

	err := checkUuid(uuidReq)
	if err != nil {
		return nil, err
	}

	ii := &inv.InventoryItem{Uuid: uuidReq}
	so := stor.Object(ii)
	err = stor.Read(so)
	if err != nil {
		return nil, err
	}

	return &inv.GetItemResponse{
		Item: ii,
	}, nil
}

func (s *InventoryServer) GetResourceItem(ctx context.Context, req *inv.GetResourceItemRequest) (*inv.GetItemResponse, error) {
	log.Info("Get Resource Item")

	uuidReq := req.Uuid

	err := checkUuid(uuidReq)
	if err != nil {
		return nil, err
	}

	ri := &inv.ResourceItem{Uuid: uuidReq}
	ro := stor.Object(ri)
	err = stor.Read(ro)
	if err != nil {
		return nil, err
	}

	ii := &inv.InventoryItem{Uuid: ri.Parent}
	io := stor.Object(ii)
	err = stor.Read(io)
	if err != nil {
		return nil, err
	}

	return &inv.GetItemResponse{
		Item: ii,
	}, nil
}

func (s *InventoryServer) UpdateInventoryManagement(ctx context.Context, req *inv.UpdateInventoryRequest) (*inv.InventoryItemResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("inv.Modifyinv.InventoryItem: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	iouuid := req.IoUuid
	err := checkUuid(iouuid)
	if err != nil {
		return nil, err
	}

	_, _, err = net.ParseCIDR(req.MgmtAddr)
	if err != nil {
		return nil, err
	}

	io := &inv.InventoryItem{Uuid: iouuid}

	obj := stor.Object(io)

	err = stor.Read(obj)
	if err != nil {
		return nil, err
	}

	io.Entity.Idtype = inv.Entity_IP
	io.Entity.Identification = req.MgmtAddr

	objs := []stor.Object{io}

	err = stor.WriteObjects(objs, false)
	if err != nil {
		return nil, err
	}

	return &inv.InventoryItemResponse{IoUuid: iouuid}, nil
}

func (s *InventoryServer) BulkUpdateManagement(ctx context.Context, req *inv.BulkUpdateRequest) (*inv.BulkUpdateResponse, error) {
	if req == nil {
		errMsg := fmt.Sprintf("inv.Modifyinv.InventoryItem: Nil Request")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	if len(req.Bur) <= 0 {
		return nil, fmt.Errorf("empty bulk request")
	}

	respList := make([]*inv.InventoryItemResponse, 0)
	for _, bur := range req.Bur {
		resp, err := s.UpdateInventoryManagement(ctx, bur)
		if err != nil {
			return nil, err
		}
		respList = append(respList, resp)
	}

	return &inv.BulkUpdateResponse{Response: respList}, nil
}

func main() {

	var debug bool
	var port int

	flag.IntVar(&port, "port", pkg.DefaultInventoryPort, "set the Inventoryd control port")
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

	log.Info(fmt.Sprintf("Inventoryd starting up on port %d", port))
	log.Infof("db settings: %+v\n", etcdCfg)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	inv.RegisterInventoryServer(grpcServer, &InventoryServer{})
	grpcServer.Serve(lis)
}
