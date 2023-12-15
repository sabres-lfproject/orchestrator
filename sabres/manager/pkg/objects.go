package pkg

import (
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type IPManagement struct {
	Name    string
	IP      string
	Uuid    string
	Version int64
}

type Slice struct {
	Name    string
	Uuid    string
	Devices []map[string]string
	Edges   []map[string]string
	Version int64
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
	return fmt.Sprintf("%s/%s", IPPrefix, x.Uuid)
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

	return fmt.Sprintf("%s/%s", SlicePrefix, x.Uuid)
}
func (x *Slice) SetVersion(v int64) { x.Version = v }
func (x *Slice) GetVersion() int64  { return x.Version }
func (x *Slice) Value() interface{} { return x }
