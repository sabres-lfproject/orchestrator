package pkg

import (
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	cbspkg "pulwar.isi.edu/sabres/orchestrator/sabres/cbs/service/pkg"
)

type IPManagement struct {
	Name string
	IP   string
	Uuid string
}

type Slice struct {
	Name    string
	Uuid    string
	Devices []*cbspkg.CBSNode
	Edges   []*cbspkg.CBSEdge
}

var (
	IPPrefix    = "/ipmgmt"
	SlicePrefix = "/slice"
)

// Required functions for stor
// IPManagement definitions
func (x *IPManagement) Key() string {
	_, err := uuid.Parse(x.Uuid)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s/%s", InvPrefix, x.Uuid)
}
func (x *IPManagement) SetVersion(v int64) { x.Version = v }
func (x *IPManagement) GetVersion() int64  { return x.Version }
func (x *IPManagement) Value() interface{} { return x }

// Slice definitions
func (x *Slice) Key() string {
	_, err := uuid.Parse(x.Uuid)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s/%s", ResPrefix, x.Uuid)
}
func (x *Slice) SetVersion(v int64) { x.Version = v }
func (x *Slice) GetVersion() int64  { return x.Version }
func (x *Slice) Value() interface{} { return x }
