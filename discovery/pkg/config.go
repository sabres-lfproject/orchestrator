package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	dobj "pulwar.isi.edu/sabres/orchestrator/discovery/protocol"
)

type ServiceConfig struct {
	Endpoints []*dobj.Endpoint `json:"endpoints,omitempty"`
}

func LoadServicesConfig(configPath string) ([]*dobj.Endpoint, error) {

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("could not read configuration file %s", configPath)
		return nil, err
	}

	log.Infof("%s", data)

	cfg := &ServiceConfig{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		log.Errorf("could not parse configuration file")
		return nil, err
	}

	log.WithFields(log.Fields{
		"config": fmt.Sprintf("%+v", *cfg),
	}).Debug("config")

	return cfg.Endpoints, nil
}
