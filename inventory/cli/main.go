package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"pulwar.isi.edu/sabres/orchestrator/inventory/pkg"
	"pulwar.isi.edu/sabres/orchestrator/inventory/protocol"
)

var (
	clientServer string
	clientPort   int
	addr         string
)

func main() {

	root := &cobra.Command{
		Use:   "ictl",
		Short: "orchestrator's inventory controller",
	}

	root.PersistentFlags().StringVarP(
		&clientServer, "server", "s", "localhost", "inventory service address to use")
	root.PersistentFlags().IntVarP(
		&clientPort, "port", "p", pkg.DefaultInventoryPort, "inventory service port to use")

	addr = fmt.Sprintf("%s:%d", clientServer, clientPort)

	create := &cobra.Command{
		Use:   "create",
		Short: "Create inventory item",
	}
	root.AddCommand(create)

	createInventoryItem := &cobra.Command{
		Use:   "inv <file>",
		Short: "Create an inventory item through config file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			createInventoryItemFunc(args[0])
		},
	}
	create.AddCommand(createInventoryItem)

	del := &cobra.Command{
		Use:   "del",
		Short: "Destroy inventory item",
	}
	root.AddCommand(del)

	delInventoryItem := &cobra.Command{
		Use:   "del inv <item>",
		Short: "Delete a inventory item",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			delInventoryItemFunc(args[0])
		},
	}
	del.AddCommand(delInventoryItem)

	showConfig := &cobra.Command{
		Use:   "show",
		Short: "Show system config",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showConfigFunc(args[0])
		},
	}
	root.AddCommand(showConfig)

	listInventory := &cobra.Command{
		Use:   "list",
		Short: "Show system config",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			listInventoryFunc()
		},
	}
	root.AddCommand(listInventory)

	root.Execute()
}

func createInventoryItemFunc(fi string) {

	fmt.Printf("test")

	ii_list, err := pkg.LoadInventoryItemConfig(fi)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("service config: %s\n", ii_list)

	if len(ii_list) == 0 {
		fmt.Printf("invalid config\n")
		return
	}

	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	for _, inv := range ii_list {
		pkg.WithInventory(addr, func(c protocol.InventoryClient) error {
			req := &protocol.CreateInventoryItemRequest{
				Request: inv,
			}

			fmt.Printf("sent request: %v\n", req)
			resp, err := c.CreateInventoryItem(context.TODO(), req)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%v\n", resp)

			return nil
		})
	}
}

func delInventoryItemFunc(Uuid string) {

	_, err := uuid.Parse(Uuid)
	if err != nil {
		fmt.Print(err)
		return
	}

	ii := &protocol.InventoryItem{}
	ii.Uuid = Uuid

	pkg.WithInventory(addr, func(c protocol.InventoryClient) error {
		req := &protocol.DeleteInventoryItemRequest{
			Request: ii,
		}

		resp, err := c.DeleteInventoryItem(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", resp)

		return nil
	})

}

func showConfigFunc(Uuid string) {

	_, err := uuid.Parse(Uuid)
	if err != nil {
		fmt.Print(err)
		return
	}

	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	pkg.WithInventory(addr, func(c protocol.InventoryClient) error {

		resp, err := c.GetInventoryItem(context.TODO(), &protocol.GetInventoryItemRequest{
			Uuid: Uuid,
		})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%+v\n", resp)

		return nil
	})
}

func listInventoryFunc() {

	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	pkg.WithInventory(addr, func(c protocol.InventoryClient) error {

		resp, err := c.ListInventoryItems(context.TODO(), &protocol.ListInventoryItemsRequest{})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%+v\n", resp)

		return nil
	})
}
