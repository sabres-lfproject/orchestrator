package pkg

import (
	"fmt"

	"google.golang.org/grpc"
	"pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
)

var (
	DefaultInventoryPort = 15005
)

func Endpoint(server string, port int) string {
	return fmt.Sprintf("%s:%d", server, port)
}

func WithInventory(endpoint string, f func(protocol.InventoryClient) error) error {

	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to moa service: %v", err)
	}
	client := protocol.NewInventoryClient(conn)
	defer conn.Close()

	return f(client)
}
