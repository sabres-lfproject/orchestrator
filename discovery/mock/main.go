package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"pulwar.isi.edu/sabres/orchestrator/discovery/pkg"

	"gitlab.com/mergetb/tech/stor"
	inventory "pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
	config "pulwar.isi.edu/sabres/orchestrator/pkg"
)

var (
	EtcdConfigPath string = "/var/orchestrator/config.cfg"
)

func main() {

	var debug bool
	var port int
	var datadir string

	flag.IntVar(&port, "port", pkg.DefaultMockDiscoveryPort, "set the Inventoryd control port")
	flag.BoolVar(&debug, "debug", false, "enable extra debug logging")
	flag.StringVar(&datadir, "dir", ".", "directory with mock data")

	flag.Parse()

	portStr := os.Getenv("MOCKPORT")
	if portStr != "" {
		portInt, err := strconv.Atoi(portStr)
		if err != nil {
			log.Warningf("Failed to convert MOCKPORT to int, ignored: %v", portStr)
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

	datadirStr := os.Getenv("DATADIR")
	if datadirStr != "" {
		datadir = datadirStr
	}

	r := gin.Default()

	if debug {
		log.SetLevel(log.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		log.SetLevel(log.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	}

	r.GET("/ping", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", []byte("pong"))
	})

	r.GET("/aether-roc-api/aether/v2.1.x/enterprise/site", func(c *gin.Context) {
		var mockSliceData []map[string]interface{}
		mockSliceStr, err := ioutil.ReadFile(fmt.Sprintf("%s/slice.json", datadir))
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(mockSliceStr, &mockSliceData)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, mockSliceData)
	})

	r.GET("/resources", func(c *gin.Context) {
		var mockResourceData []map[string]interface{}
		mockResourceStr, err := ioutil.ReadFile(fmt.Sprintf("%s/resources.json", datadir))
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(mockResourceStr, &mockResourceData)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, mockResourceData)
	})

	r.GET("/mock", func(c *gin.Context) {
		err := AddMockToEtcd()
		if err != nil {
			c.JSON(404, "failed")
			log.Errorf("%v\n", err)
		} else {
			c.JSON(http.StatusOK, "ok")
		}
	})

	mockAddr := fmt.Sprintf("localhost:%d", port)
	log.Infof("starting mock on: %s\n", mockAddr)

	r.Run(mockAddr)
}

func AddMockToEtcd() error {

	configPath := fmt.Sprintf("/data/discovery/mock/resources.json")

	log.Infof("%s\n", configPath)

	data, err := os.Open(configPath)
	if err != nil {
		return err
	}

	jsonData, err := io.ReadAll(data)
	if err != nil {
		return err
	}

	rec := make([]inventory.ResourceItem, 0)
	err = json.Unmarshal(jsonData, &rec)
	if err != nil {
		return err
	}

	log.Infof("%#v\n", rec)

	objs := make([]stor.Object, 0)
	for x, r := range rec {
		i := &inventory.InventoryItem{
			Uuid:     r.Uuid,
			Resource: &rec[x],
			Entity: &inventory.Entity{
				Idtype: inventory.Entity_IP,
			},
			Notes: fmt.Sprintf("manual entry %d", x),
		}
		objs = append(objs, i)
	}

	for _, o := range objs {
		log.Infof("%#v\n", o)
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

	err = stor.WriteObjects(objs, false)
	if err != nil {
		return err
	}

	log.Infof("wrote objs to db\n")

	return nil
}
