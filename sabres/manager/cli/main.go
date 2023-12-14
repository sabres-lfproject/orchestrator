package main

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cbs "pulwar.isi.edu/sabres/orchestrator/sabres/cbs/service/pkg"
	"pulwar.isi.edu/sabres/orchestrator/sabres/manager/pkg"
	"pulwar.isi.edu/sabres/orchestrator/sabres/manager/protocol"
	network "pulwar.isi.edu/sabres/orchestrator/sabres/network/pkg"
	proto "pulwar.isi.edu/sabres/orchestrator/sabres/network/protocol"
)

var (
	clientServer    string
	clientPort      int
	cbsServer       string
	cbsPort         int
	networkServer   string
	networkPort     int
	inventoryServer string
	inventoryPort   int
)

func main() {

	root := &cobra.Command{
		Use:   "smgmt",
		Short: "sabres orchestrator management controller",
	}

	root.PersistentFlags().StringVarP(
		&clientServer, "server", "s", "localhost", "manager service address to use")
	root.PersistentFlags().IntVarP(
		&clientPort, "port", "p", pkg.DefaultManagementPort, "manager service port to use")

	root.PersistentFlags().StringVarP(
		&cbsServer, "cbsserver", "c", "localhost", "cbs service address to use")
	root.PersistentFlags().IntVarP(
		&cbsPort, "cbsport", "d", cbs.DefaultCBSPort, "cbs service port to use")

	root.PersistentFlags().StringVarP(
		&networkServer, "networkserver", "e", "localhost", "network service address to use")
	root.PersistentFlags().IntVarP(
		&networkPort, "networkport", "f", network.DefaultNetworkPort, "network service port to use")

	root.PersistentFlags().StringVarP(
		&inventoryServer, "inventoryserver", "g", "localhost", "inventory service address to use")
	root.PersistentFlags().IntVarP(
		&inventoryPort, "inventoryport", "h", inventory.DefaultInventoryPort, "inventory service port to use")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a sabres slice",
	}
	root.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Create a sabres slice",
	}
	root.AddCommand(deleteCmd)

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Create a sabres slices",
	}
	root.AddCommand(showCmd)

	configureCmd := &cobra.Command{
		Use:   "configure",
		Short: "Create a sabres slice",
	}
	root.AddCommand(configureCmd)

	createNetworkSlice := &cobra.Command{
		Use:   "slice <request-file>",
		Short: "Create a slice given a cbs formatted request file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
			cbsAddr := fmt.Sprintf("%s:%d", cbsServer, cbsPort)
			networkAddr := fmt.Sprintf("%s:%d", networkServer, networkPort)
			inventoryAddr := fmt.Sprintf("%s:%d", inventoryServer, inventoryPort)
			createNetworkSliceFunc(addr, cbsAddr, networkAddr, inventoryAddr, args[0])
		},
	}
	createCmd.AddCommand(createNetworkSlice)

	deleteNetworkSlice := &cobra.Command{
		Use:   "slice <uuid>",
		Short: "Delete the existing graph",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
			inventoryAddr := fmt.Sprintf("%s:%d", inventoryServer, inventoryPort)
			deleteNetworkSliceFunc(addr, inventoryAddr, args[0])
		},
	}
	deleteCmd.AddCommand(deleteNetworkSlice)

	showConfig := &cobra.Command{
		Use:   "slices",
		Short: "Show current slices",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
			inventoryAddr := fmt.Sprintf("%s:%d", inventoryServer, inventoryPort)
			showNetworkSliceFunc(addr, inventoryAddr)
		},
	}
	showCmd.AddCommand(showConfig)

	configureConfig := &cobra.Command{
		Use:   "slice <uuid>",
		Short: "Show current slices",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
			inventoryAddr := fmt.Sprintf("%s:%d", inventoryServer, inventoryPort)
			configureNetworkSliceFuncs(addr, inventoryAddr, args[0])
		},
	}
	configureCmd.AddCommand(configureConfig)

	root.Execute()
}

func createNetworkSliceFunc(mgmtAddr, cbsAddr, netAddr, invAddr, fileName string) {

	contents, err := io.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	constraints := make([]*proto.Constraint, 0)

	err = json.Marshal(constraints, contents)
	if err != nil {
		log.Fatal(err)
	}

	pkg.WithManager(mgmtAddr, func(c protocol.NetworkClient) error {
		fmt.Printf("sending request\n")
		resp, err := c.CreateSlice(context.TODO(), &protocol.CreateSliceRequest{
			CbsAddr: cbsAddr,
			NetAddr: netAddr,
			InvAddr: invAddr,
			Request: constraints,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("response: %v\n", resp)

		return nil
	})
}

func deleteNetworkSliceFunc(mgmtAddr, invAddr, uuid string) {
	pkg.WithNetwork(addr, func(c protocol.NetworkClient) error {
		fmt.Printf("sent request\n")
		resp, err := c.DeleteSlice(context.TODO(), &protocol.DeleteSliceRequest{
			InvAddr: invAddr,
			Uuid:    uuid,
		})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", resp)

		return nil
	})
}

func showConfigFunc(mgmtAddr, invAddr string) {
	pkg.WithNetwork(addr, func(c protocol.NetworkClient) error {
		resp, err := c.ShowSlice(context.TODO(), &protocol.ShowSliceRequest{
			InvAddr: invAddr,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Slice: %+v\n", resp)

		return nil
	})
}

func configureConfigFunc(mgmtAddr, invAddr, uuid string) {
	pkg.WithNetwork(addr, func(c protocol.NetworkClient) error {
		resp, err := c.ConfigureSlice(context.TODO(), &protocol.ConfigureSliceRequest{
			InvAddr: invAddr,
			Uuid:    uuid,
		})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Slice: %+v\n", resp)

		return nil
	})
}
