package main

import (
	"fmt"

	"github.com/spf13/cobra"

	// "github.com/onosproject/aether-roc-api/pkg/aether_2_1_0/types"
	"pulwar.isi.edu/sabres/orchestrator/sabres/sd-core-interface/pkg"
)

var (
	aetherrocServer string
	aetherrocPort   int
	addr            string
)

func main() {

	root := &cobra.Command{
		Use:   "sdcore",
		Short: "sabres orchestrator's sdcore command line interface",
	}

	root.PersistentFlags().StringVarP(
		&aetherrocServer, "server", "s", "localhost", "network service address to use")
	root.PersistentFlags().IntVarP(
		&aetherrocPort, "port", "p", 31194, "network service port to use")

	list := &cobra.Command{
		Use:   "list",
		Short: "list sdcore things",
	}

	root.AddCommand(list)

	listEnterprises := &cobra.Command{
		Use:   "enterprise",
		Short: "list all sd-core enterprises",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			enters, err := pkg.GetEnterprises(addr)
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			for _, e := range enters {
				// log.Infof("Enterprise: %v\n", e)
				fmt.Printf("Enterprise: %v\n", e)
			}

		},
	}
	list.AddCommand(listEnterprises)

	listSites := &cobra.Command{
		Use:   "site <enterprise>",
		Short: "list all sites for an sd-core enterprises",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			enters, err := pkg.GetSites(addr, args[0])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			for _, e := range enters {
				// log.Infof("Site: %v\n", e)
				fmt.Printf("Site: %v\n", e)
			}

		},
	}
	list.AddCommand(listSites)

	listDeviceGroup := &cobra.Command{
		Use:   "device-group <enterprise> <site>",
		Short: "list all device-groups in a site",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			gsdr, err := pkg.GetSiteDetails(addr, args[0], args[1], "device-group")
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Device Groups:\n%v\n", gsdr)

		},
	}
	list.AddCommand(listDeviceGroup)

	listDevices := &cobra.Command{
		Use:   "devices <enterprise> <site>",
		Short: "list all devices in a site",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			gsdr, err := pkg.GetSiteDetails(addr, args[0], args[1], "device")
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Devices:\n%v\n", gsdr)

		},
	}
	list.AddCommand(listDevices)

	listIpDomain := &cobra.Command{
		Use:   "ip <enterprise> <site>",
		Short: "list all ip information for a site",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			gsdr, err := pkg.GetSiteDetails(addr, args[0], args[1], "ip-domain")
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("IP Domains:\n%v\n", gsdr)

		},
	}
	list.AddCommand(listIpDomain)

	listSimCard := &cobra.Command{
		Use:   "sim-card <enterprise> <site>",
		Short: "list all sim-card information for a site",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			gsdr, err := pkg.GetSiteDetails(addr, args[0], args[1], "sim-card")
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Sim Cards:\n%v\n", gsdr)

		},
	}
	list.AddCommand(listSimCard)

	listSlice := &cobra.Command{
		Use:   "slice <enterprise> <site>",
		Short: "list all slice information for a site",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			gsdr, err := pkg.GetSiteDetails(addr, args[0], args[1], "slice")
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Slices:\n%v\n", gsdr)

		},
	}
	list.AddCommand(listSlice)

	listSmallCell := &cobra.Command{
		Use:   "small-cell <enterprise> <site>",
		Short: "list all small-cell information for a site",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			gsdr, err := pkg.GetSiteDetails(addr, args[0], args[1], "small-cell")
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Small Cells:\n%v\n", gsdr)

		},
	}
	list.AddCommand(listSmallCell)

	listUPF := &cobra.Command{
		Use:   "upf <enterprise> <site>",
		Short: "list all upf information for a site",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			gsdr, err := pkg.GetSiteDetails(addr, args[0], args[1], "upf")
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("UPFs:\n%v\n", gsdr)

		},
	}
	list.AddCommand(listUPF)

	/* Create */

	create := &cobra.Command{
		Use:   "create",
		Short: "create sdcore things",
	}

	root.AddCommand(create)

	createEnterprises := &cobra.Command{
		Use:   "enterprise",
		Short: "create sd-core enterprise",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Create Enterprise Not implemented at SD-Core openapi\n")
		},
	}
	create.AddCommand(createEnterprises)

	createSite := &cobra.Command{
		Use:   "site <enterprise-name> <site-name>",
		Short: "create a site for an sd-core enterprises",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			enters, err := pkg.CreateSite(addr, args[0], args[1])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Create Site: %v\n", enters)

		},
	}
	create.AddCommand(createSite)

	createFile := &cobra.Command{
		Use:   "file",
		Short: "create sdcore things from a file",
	}
	create.AddCommand(createFile)

	createSiteFromFile := &cobra.Command{
		Use:   "site <enterprise-name> <site-name> <file-name-with-site-info>",
		Short: "create a site for an sd-core enterprises",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			enters, err := pkg.CreateSiteFromFile(addr, args[0], args[1], args[2])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Create Site: %v\n", enters)

		},
	}
	createFile.AddCommand(createSiteFromFile)

	createSliceFromFile := &cobra.Command{
		Use:   "slice <enterprise-name> <site-name> <file-name-with-site-info>",
		Short: "create a slice for an sd-core enterprises",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			enters, err := pkg.CreateSliceFromFile(addr, args[0], args[1], args[2])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Create Slice: %v\n", enters)

		},
	}
	createFile.AddCommand(createSliceFromFile)

	createDevicesFromFile := &cobra.Command{
		Use:   "device <enterprise-name> <site-name> <file-name-with-site-info>",
		Short: "create devices for an sd-core enterprises",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			enters, err := pkg.CreateDevicesFromFile(addr, args[0], args[1], args[2])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Create Devices: %v\n", enters)

		},
	}
	createFile.AddCommand(createDevicesFromFile)

	createDeviceGroupFromFile := &cobra.Command{
		Use:   "device-group <enterprise-name> <site-name> <file-name-with-site-info>",
		Short: "create device group for an sd-core enterprises",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			//log.Infof("calling get enterprise: %s", addr)
			enters, err := pkg.CreateDeviceGroupFromFile(addr, args[0], args[1], args[2])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Create DeviceGroup: %v\n", enters)

		},
	}
	createFile.AddCommand(createDeviceGroupFromFile)

	deleteSD := &cobra.Command{
		Use:   "delete",
		Short: "delete sdcore things",
	}
	root.AddCommand(deleteSD)

	deleteDeviceGroup := &cobra.Command{
		Use:   "device-group <enterprise-name> <site-name> <device-group-name>",
		Short: "delete device group from sd-core site",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			enters, err := pkg.DeleteAetherObject(addr, args[0], args[1], "device-group", args[2])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Delete DeviceGroup: %v\n", enters)
		},
	}
	deleteSD.AddCommand(deleteDeviceGroup)

	deleteSlice := &cobra.Command{
		Use:   "slice <enterprise-name> <site-name> <slice-name>",
		Short: "delete slice from sd-core site",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			enters, err := pkg.DeleteAetherObject(addr, args[0], args[1], "slice", args[2])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Delete Slice: %v\n", enters)
		},
	}
	deleteSD.AddCommand(deleteSlice)

	deleteDevice := &cobra.Command{
		Use:   "device <enterprise-name> <site-name> <device-name>",
		Short: "delete devices from sd-core site",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			enters, err := pkg.DeleteAetherObject(addr, args[0], args[1], "device", args[2])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Delete Device: %v\n", enters)
		},
	}
	deleteSD.AddCommand(deleteDevice)

	deleteSite := &cobra.Command{
		Use:   "site <enterprise-name> <site-name>",
		Short: "delete site from sd-core site",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			addr = fmt.Sprintf("%s:%d", aetherrocServer, aetherrocPort)
			enters, err := pkg.DeleteAetherObject(addr, args[0], args[1], "site", args[2])
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				return
			}

			fmt.Printf("Delete Site: %v\n", enters)
		},
	}
	deleteSD.AddCommand(deleteSite)

	root.Execute()
}
