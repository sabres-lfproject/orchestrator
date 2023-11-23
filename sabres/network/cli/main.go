package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"pulwar.isi.edu/sabres/orchestrator/sabres/network/pkg"
	"pulwar.isi.edu/sabres/orchestrator/sabres/network/protocol"
)

var (
	clientServer string
	clientPort   int
	addr         string
)

func main() {

	root := &cobra.Command{
		Use:   "snctl",
		Short: "sabres orchestrator's network controller",
	}

	root.PersistentFlags().StringVarP(
		&clientServer, "server", "s", "localhost", "network service address to use")
	root.PersistentFlags().IntVarP(
		&clientPort, "port", "p", pkg.DefaultNetworkPort, "network service port to use")

	addr = fmt.Sprintf("%s:%d", clientServer, clientPort)

	createNetworkItem := &cobra.Command{
		Use:   "create",
		Short: "Create a new graph from inventory",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			createNetworkItemFunc()
		},
	}
	root.AddCommand(createNetworkItem)

	delNetworkItem := &cobra.Command{
		Use:   "delete",
		Short: "Delete the existing graph",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			delNetworkItemFunc()
		},
	}
	root.AddCommand(delNetworkItem)

	showConfig := &cobra.Command{
		Use:   "show",
		Short: "Show the existing graph if it exists",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			showConfigFunc()
		},
	}
	root.AddCommand(showConfig)

	solve := &cobra.Command{
		Use:   "solve <file>",
		Short: "solve the constraints given network topology",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			solveFunc(args[0])
		},
	}
	root.AddCommand(solve)

	setHost := &cobra.Command{
		Use:   "set <host> <port>",
		Short: "set the cbs service's host and port",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			setHostFunc(args[0], args[1])
		},
	}
	root.AddCommand(setHost)

	root.Execute()
}

func createNetworkItemFunc() {
	pkg.WithNetwork(addr, func(c protocol.NetworkClient) error {
		fmt.Printf("sent request\n")
		resp, err := c.CreateGraph(context.TODO(), &protocol.CreateGraphRequest{})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", resp)

		return nil
	})
}

func delNetworkItemFunc() {
	pkg.WithNetwork(addr, func(c protocol.NetworkClient) error {
		fmt.Printf("sent request\n")
		resp, err := c.DeleteGraph(context.TODO(), &protocol.DeleteGraphRequest{})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", resp)

		return nil
	})
}

func showConfigFunc() {
	pkg.WithNetwork(addr, func(c protocol.NetworkClient) error {
		resp, err := c.ShowGraph(context.TODO(), &protocol.ShowGraphRequest{})

		if err != nil {
			log.Fatal(err)
		}

		if !resp.Exists {
			fmt.Printf("Graph does not exist, run create first.\n")
		}

		fmt.Printf("Graph: %+v\n", resp)

		return nil
	})
}

func setHostFunc(host, port string) {
	pkg.WithNetwork(addr, func(c protocol.NetworkClient) error {
		resp, err := c.SetCBSLocation(context.TODO(), &protocol.SetCBSRequest{
			Host: host,
			Port: port,
		})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%+v\n", resp)

		return nil
	})
}

func solveFunc(fi string) {

	//TODO: read constraints from file
	pkg.WithNetwork(addr, func(c protocol.NetworkClient) error {
		// TODO: add constraints here
		resp, err := c.RequestSolution(context.TODO(), &protocol.SolveRequest{})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%+v\n", resp.Response)

		return nil
	})
}
