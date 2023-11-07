package pkg

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	protocol "pulwar.isi.edu/sabres/orchestrator/inventory/protocol"

	"gopkg.in/yaml.v2"
)

type ServiceConfig struct {
	Inventory []*protocol.InventoryItem `yaml:"inventory"`
}

func LoadInventoryItemConfig(configPath string) ([]*protocol.InventoryItem, error) {

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("could not read configuration file %s", configPath)
		return nil, err
	}

	log.Infof("%s", data)

	cfg := &ServiceConfig{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Errorf("could not parse configuration file")
		return nil, err
	}

	log.WithFields(log.Fields{
		"config": fmt.Sprintf("%+v", *cfg),
	}).Debug("config")

	return cfg.Inventory, nil
}
