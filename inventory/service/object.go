package inventory

import "fmt"

var (
	InvPrefix = "/inventory"
	ResPrefix = "/resource"
)

// Required functions for stor
// InventoryItem definitions
func (x *InventoryItem) Key() string {
	return fmt.Sprintf("%s/%s", InvPrefix, x.Uuid)
}

func (x *InventoryItem) SetVersion(v int64) { x.Version = v }

func (x *InventoryItem) Value() interface{} { return x }

// ResourceItem definitions
func (x *ResourceItem) Key() string {
	return fmt.Sprintf("%s/%s", ResPrefix, x.Uuid)
}

func (x *ResourceItem) SetVersion(v int64) { x.Version = v }

func (x *ResourceItem) Value() interface{} { return x }
