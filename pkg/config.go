package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mergetb/tech/stor"
	"gopkg.in/yaml.v2"
)

// ServiceConfig encapsulates information for communicating with services.
type ServiceConfig struct {
	Address string          `yaml:address",omitempty"`
	Port    int             `yaml:port",omitempty"`
	TLS     *stor.TLSConfig `yaml:tls",omitempty"`
	Timeout int             `yaml:timeout",omitempty"`
}

// ServicesConfig encapsulates information for communicating with services.
type ServicesConfig struct {
	Etcd *ServiceConfig `yaml:",omitempty"`
}

// Endpoint returns the endpoint string of a service config.
func (s *ServiceConfig) Endpoint() string {
	return fmt.Sprintf("%s:%d", s.Address, s.Port)
}

func LoadConfig(configPath string) (*ServicesConfig, error) {

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("could not read configuration file %s", configPath)
		return nil, err
	}

	log.Infof("%s", data)

	cfg := &ServicesConfig{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Errorf("could not parse configuration file")
		return nil, err
	}

	log.Infof("%v", cfg.Etcd)

	log.WithFields(log.Fields{
		"config": fmt.Sprintf("%+v", *cfg),
	}).Debug("config")

	return cfg, nil
}

func ReadENVSettings(config *ServicesConfig) error {

	if config.Etcd != nil {
		etcdPort := os.Getenv("ETCDPORT")
		log.Debugf("etcd port env: %s", etcdPort)
		if etcdPort != "" {
			intPort, err := strconv.Atoi(etcdPort)
			if err != nil {
				log.Warningf("ETCDPORT ENV unable to be set, cannot convert to int: %v", err)
			} else {
				config.Etcd.Port = intPort
			}
		}

		etcdHost := os.Getenv("ETCDHOST")
		log.Debugf("etcd host env: %s", etcdHost)
		if etcdHost != "" {
			config.Etcd.Address = etcdHost
		}
	}

	return nil
}

func SetEtcdSettings(config *ServicesConfig) (*stor.Config, error) {
	cfg := &stor.Config{}

	if config.Etcd != nil {
		cfg.Address = config.Etcd.Address
		cfg.Port = config.Etcd.Port
		cfg.TLS = config.Etcd.TLS
		cfg.Timeout = time.Duration(config.Etcd.Timeout) * time.Millisecond
	} else {
		return nil, fmt.Errorf("No ETCD config found.\n")
	}

	return cfg, nil
}

type ClickSettings struct {
	Host      string `yaml:host",omitempty"`
	Interface string `yaml:interface",omitempty"`
	Config    string `yaml:config",omitempty"`
}

type ClickFile struct {
	Interfaces []*ClickSettings `yaml:interfaces",omitempty"`
}

func LoadClickConfig(configPath string) ([]*ClickSettings, error) {

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("could not read configuration file %s", configPath)
		return nil, err
	}

	log.Infof("%s", data)

	cfg := &ClickFile{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Errorf("could not parse configuration file")
		return nil, err
	}

	log.WithFields(log.Fields{
		"config": fmt.Sprintf("%+v", *cfg),
	}).Debug("config")

	return cfg.Interfaces, nil
}
