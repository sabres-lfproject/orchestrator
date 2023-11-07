package protocol

import (
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	InvPrefix = "/inventory"
	ResPrefix = "/resource"
)

// Required functions for stor
// InventoryItem definitions
func (x *InventoryItem) Key() string {
	_, err := uuid.Parse(x.Uuid)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s/%s", InvPrefix, x.Uuid)
}

func (x *InventoryItem) SetVersion(v int64) { x.Version = v }

func (x *InventoryItem) Value() interface{} { return x }

// ResourceItem definitions
func (x *ResourceItem) Key() string {
	_, err := uuid.Parse(x.Uuid)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s/%s", ResPrefix, x.Uuid)
}

func (x *ResourceItem) SetVersion(v int64) { x.Version = v }

func (x *ResourceItem) Value() interface{} { return x }
