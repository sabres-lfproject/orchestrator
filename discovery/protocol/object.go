package protocol

import (
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	EndpointPrefix = "/endpoints"
)

func (x *Endpoint) Key() string {
	_, err := uuid.Parse(x.Services.Uuid)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s/%s", EndpointPrefix, x.Services.Uuid)
}

func (x *Endpoint) SetVersion(v int64) { x.Version = v }

func (x *Endpoint) Value() interface{} { return x }
