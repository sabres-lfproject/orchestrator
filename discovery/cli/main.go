package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"pulwar.isi.edu/sabres/orchestrator/discovery/pkg"
	proto "pulwar.isi.edu/sabres/orchestrator/discovery/protocol"
)

var (
	clientServer string
	clientPort   int
	addr         string
)

func main() {

	root := &cobra.Command{
		Use:   "dctl",
		Short: "orchestrator's discovery controller",
	}

	root.PersistentFlags().StringVarP(
		&clientServer, "server", "s", "localhost", "discovery service address to use")
	root.PersistentFlags().IntVarP(
		&clientPort, "port", "p", pkg.DefaultDiscoveryPort, "discovery service port to use")

	addr = fmt.Sprintf("%s:%d", clientServer, clientPort)

	create := &cobra.Command{
		Use:   "create",
		Short: "Create discovery endpoint",
	}
	root.AddCommand(create)

	createDEP := &cobra.Command{
		Use:   "disc <file>",
		Short: "Create an discovery endpoint through config file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			createDEPFunc(args[0])
		},
	}
	create.AddCommand(createDEP)

	del := &cobra.Command{
		Use:   "del",
		Short: "Destroy discovery endpoint",
	}
	root.AddCommand(del)

	delDEP := &cobra.Command{
		Use:   "del disc <endpoint>",
		Short: "Delete a discovery endpoint",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			delDEPFunc(args[0])
		},
	}
	del.AddCommand(delDEP)

	showConfig := &cobra.Command{
		Use:   "show",
		Short: "Show system config",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showConfigFunc(args[0])
		},
	}
	root.AddCommand(showConfig)

	listDiscovery := &cobra.Command{
		Use:   "list",
		Short: "Show system config",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			listDiscoveryFunc()
		},
	}
	root.AddCommand(listDiscovery)

	root.Execute()
}

func createDEPFunc(fi string) {

	fmt.Printf("test")

	eps, err := pkg.LoadServicesConfig(fi)
	if err != nil {
		log.Fatal(err)
	}

	if len(eps) > 1 {
		log.Fatal("function only takes one endpoint currently")
	}

	ep := eps[0]
	fmt.Printf("service config: %s\n", ep)

	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	pkg.WithDiscovery(addr, func(c proto.DiscoveryClient) error {
		req := &proto.CreateDEPRequest{
			Endpoint: ep,
		}

		fmt.Printf("sent request: %v\n", req)
		resp, err := c.CreateDEP(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", resp)

		return nil
	})
}

func delDEPFunc(Uuid string) {

	_, err := uuid.Parse(Uuid)
	if err != nil {
		fmt.Print(err)
		return
	}

	pkg.WithDiscovery(addr, func(c proto.DiscoveryClient) error {
		req := &proto.DeleteDEPRequest{
			Uuid: Uuid,
		}

		resp, err := c.DeleteDEP(context.TODO(), req)
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
	pkg.WithDiscovery(addr, func(c proto.DiscoveryClient) error {

		resp, err := c.GetDEP(context.TODO(), &proto.GetDEPRequest{
			Uuid: Uuid,
		})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%+v\n", resp)

		return nil
	})
}

func listDiscoveryFunc() {

	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	pkg.WithDiscovery(addr, func(c proto.DiscoveryClient) error {

		resp, err := c.ListDEPs(context.TODO(), &proto.ListDEPRequest{})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%+v\n", resp)

		return nil
	})
}
