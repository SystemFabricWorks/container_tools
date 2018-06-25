package main

import (
	"fmt"
	"github.com/Mellanox/sriovnet"
	"github.com/spf13/cobra"
)

var sriovCmds = &cobra.Command{
	Use:   "sriov",
	Short: "sriov management commands for netdevices",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return nil
	},
}

var pfNetdev string
var vfIndex int

func enableSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specific valid PF netdevice")
		return
	}

	err1 := sriovnet.EnableSriov(pfNetdev)
	if err1 != nil {
		return
	}

	handle, err2 := sriovnet.GetPfNetdevHandle(pfNetdev)
	if err2 != nil {
		return
	}
	err3 := sriovnet.ConfigVfs(handle, false)
	if err3 != nil {
		return
	}
}

func disableSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specific valid PF netdevice")
		return
	}
	sriovnet.DisableSriov(pfNetdev)
}

func listSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specific valid PF netdevice")
		return
	}
	handle, err2 := sriovnet.GetPfNetdevHandle(pfNetdev)
	if err2 != nil {
		return
	}

	for _, vf := range handle.List {
		vfName := sriovnet.GetVfNetdevName(handle, vf)
		fmt.Printf("%v ", vfName)
	}
	fmt.Printf("\n")
}

func unbindSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specific valid PF netdevice")
		return
	}
	handle, err2 := sriovnet.GetPfNetdevHandle(pfNetdev)
	if err2 != nil {
		return
	}

	if vfIndex != -1 {
		var found bool
		var err error
		for _, vf := range handle.List {
			if vfIndex == vf.Index {
				found = true
				fmt.Printf("Unbinding VF: %d\n", vf.Index)
				err = sriovnet.UnbindVf(handle, vf)
				if err != nil {
					fmt.Printf("Fail to Unbind VF: ", err)
					break
				}
			}
		}
		if found == false {
			fmt.Println("VF index = %d not found\n", vfIndex)
		}
	} else {
		for _, vf := range handle.List {
			fmt.Printf("Unbinding VF: %d\n", vf.Index)
			err := sriovnet.UnbindVf(handle, vf)
			if err != nil {
				fmt.Println("Fail to unbind VF: ", err)
				fmt.Printf("Continu to bind other VFs\n")
			}
		}
	}

	for _, vf := range handle.List {
		vfName := sriovnet.GetVfNetdevName(handle, vf)
		fmt.Printf("%v ", vfName)
	}
	fmt.Printf("\n")
}

func bindSriovFunc(cmd *cobra.Command, args []string) {
	if pfNetdev == "" {
		fmt.Println("Please specific valid PF netdevice")
		return
	}
	handle, err2 := sriovnet.GetPfNetdevHandle(pfNetdev)
	if err2 != nil {
		return
	}

	if vfIndex != -1 {
		var found bool
		var err error
		for _, vf := range handle.List {
			if vfIndex == vf.Index {
				found = true
				fmt.Printf("Binding VF: %d\n", vf.Index)
				err = sriovnet.BindVf(handle, vf)
				if err != nil {
					fmt.Println("Fail to bind VF: ", err)
					break
				}

				mode, _ := GetDevlinkMode(pfNetdev)
				if mode != "switchdev" {
					fmt.Println("Skipping VF rep link config")
					continue
				}
				err = SetVfRepresentorLinkUp(pfNetdev, vf.Index)
				if err != nil {
					fmt.Println("Fail to bind VF: ", err)
					break
				}
			}
		}
		if found == false {
			fmt.Println("VF index = %d not found\n", vfIndex)
		}
	} else {
		for _, vf := range handle.List {
			fmt.Printf("Binding VF: %d\n", vf.Index)
			err := sriovnet.BindVf(handle, vf)
			if err != nil {
				fmt.Println("Fail to bind VF: ", err)
				fmt.Printf("Continu to bind other VFs\n")
			}
			mode, _ := GetDevlinkMode(pfNetdev)
			if mode != "switchdev" {
				fmt.Println("Skipping VF rep link config")
				continue
			}
			err = SetVfRepresentorLinkUp(pfNetdev, vf.Index)
			if err != nil {
				fmt.Println("Fail to bind VF: ", err)
				break
			}
		}
	}

	for _, vf := range handle.List {
		vfName := sriovnet.GetVfNetdevName(handle, vf)
		fmt.Printf("%v ", vfName)
	}
	fmt.Printf("\n")
}

var enableSriovCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable sriov for PF netdevice",
	Run:   enableSriovFunc,
}

var disableSriovCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable sriov for PF netdevice",
	Run:   disableSriovFunc,
}

var listSriovCmd = &cobra.Command{
	Use:   "list",
	Short: "List sriov netdevices for PF netdevice",
	Run:   listSriovFunc,
}

var unbindSriovCmd = &cobra.Command{
	Use:   "unbind",
	Short: "Unbind a specific or all VFs of a PF netdevice",
	Run:   unbindSriovFunc,
}

var bindSriovCmd = &cobra.Command{
	Use:   "bind",
	Short: "bind a specific or all VFs of a PF netdevice",
	Run:   bindSriovFunc,
}

func init() {
	enableFlags := enableSriovCmd.Flags()
	enableFlags.StringVarP(&pfNetdev, "netdev", "n", "", "enable sriov for the PF netdevice")

	disableFlags := disableSriovCmd.Flags()
	disableFlags.StringVarP(&pfNetdev, "netdev", "n", "", "disable sriov for the PF netdevice")

	listFlags := listSriovCmd.Flags()
	listFlags.StringVarP(&pfNetdev, "netdev", "n", "", "List netdevices of the PF netdevice")

	unbindFlags := unbindSriovCmd.Flags()
	unbindFlags.IntVarP(&vfIndex, "vf", "v", -1, "vf index to unbind")
	unbindFlags.StringVarP(&pfNetdev, "netdev", "n", "", "PF netdevice whose VFs to unbind")

	bindFlags := bindSriovCmd.Flags()
	bindFlags.IntVarP(&vfIndex, "vf", "v", -1, "vf index to bind")
	bindFlags.StringVarP(&pfNetdev, "netdev", "n", "", "PF netdevice whose VFs to bind")
}

/* add new sriov command here */
var sriovCmdList = [...]*cobra.Command{
	enableSriovCmd,
	disableSriovCmd,
	listSriovCmd,
	unbindSriovCmd,
	bindSriovCmd,
}

func init() {
	for _, cmds := range sriovCmdList {
		sriovCmds.AddCommand(cmds)
	}
}
