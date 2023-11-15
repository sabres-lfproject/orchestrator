package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"pulwar.isi.edu/sabres/orchestrator/discovery/pkg"
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

	var mockSliceData []map[string]interface{}
	mockSliceStr, err := ioutil.ReadFile(fmt.Sprintf("%s/slice.json", datadir))
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(mockSliceStr, &mockSliceData)
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/aether-roc-api/aether/v2.1.x/enterprise/site", func(c *gin.Context) {
		c.JSON(http.StatusOK, mockSliceData)
	})

	var mockResourceData []map[string]interface{}
	mockResourceStr, err := ioutil.ReadFile(fmt.Sprintf("%s/resources.json", datadir))
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(mockResourceStr, &mockResourceData)
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/resources", func(c *gin.Context) {
		c.JSON(http.StatusOK, mockResourceData)
	})

	mockAddr := fmt.Sprintf("localhost:%d", port)
	log.Infof("starting mock on: %s\n", mockAddr)

	r.Run(mockAddr)
}
