package main

import (
	"context"
	"fmt"
	"github.com/Mellanox/rdmamap"
	"github.com/k8snetworkplumbingwg/sriovnet"
	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var vfUserArg int

func getDockerNetworkResourceForName(networkName string) *types.NetworkResource {
	cli, err := getRightClient()
	if err != nil {
		panic(err)
	}
	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		panic(err)
	}

	for _, network := range networks {
		if network.Name != networkName {
			continue
		}
		return &network
	}
	return nil
}

//Returns name of the network if provided in the docker run command.
func getNonDefaultNetwork(userCmdArgs []string) string {

	for _, item := range userCmdArgs {
		if strings.Contains(item, "--net=") == false {
			continue
		}
		name := strings.Split(item, "=")
		if len(name) < 2 {
			return ""
		}
		return name[1]
	}
	return ""
}

//Returns name of the network if provided in the docker run command.
func getUserVf(userCmdArgs []string) string {

	for _, item := range userCmdArgs {
		if strings.Contains(item, "--vf=") == false {
			continue
		}
		name := strings.Split(item, "=")
		if len(name) < 2 {
			return ""
		}
		return name[1]
	}
	return ""
}

func toCharDevCmdArgs(devices []string) []string {
	var cmds []string

	for _, dev := range devices {
		cmd := "--device=" + dev
		cmds = append(cmds, cmd)
	}
	return cmds
}

func allocateVf(pfNetdeviceName string) (string, error) {

	vfList, err := sriovnet.GetVfPciDevList(pfNetdeviceName)
	if err != nil {
		return "", err
	}
	for _, vf := range vfList {
		dirPath := filepath.Join("/sys/class/net", pfNetdeviceName, "device", vf, "net")
		fd, err := os.Open(dirPath)
		if err != nil {
			return "", err
		}
		defer fd.Close()
		fileInfos, err := fd.Readdir(-1)
		for i := range fileInfos {
			if fileInfos[i].Name() == "." || fileInfos[i].Name() == ".." {
				continue
			}
			vfNetdev := filepath.Join(dirPath, fileInfos[i].Name())
			_, err := os.Stat(vfNetdev)
			if err != nil {
				return "", err
			}
			return fileInfos[i].Name(), nil
		}
	}
	return "", fmt.Errorf("No VF are free")
}

// Allocate a specific VF, where vf can be given as index
func allocateSpecificVf(pfNetdeviceName string, vf string) (string, error) {

	vfString := "virtfn" + vf

	vfList, err := sriovnet.GetVfPciDevList(pfNetdeviceName)
	if err != nil {
		return "", err
	}
	for _, vf := range vfList {
		dirPath := filepath.Join("/sys/class/net", pfNetdeviceName, "device", vf, "net")
		fd, err := os.Open(dirPath)
		if err != nil {
			return "", err
		}
		defer fd.Close()
		if vf != vfString {
			continue
		}

		fileInfos, err := fd.Readdir(-1)
		for i := range fileInfos {
			if fileInfos[i].Name() == "." || fileInfos[i].Name() == ".." {
				continue
			}
			vfNetdev := filepath.Join(dirPath, fileInfos[i].Name())
			_, err := os.Stat(vfNetdev)
			if err != nil {
				return "", err
			}
			return fileInfos[i].Name(), nil
		}
	}
	return "", fmt.Errorf("Requested VF %q is unavailable", vf)
}

func getNetworkType(networkName string) string {

	network := getDockerNetworkResourceForName(networkName)
	if network != nil {
		return "plugin"
	}

	// Plugin based network is not provided, so check if it is
	// container:containerId format

	if strings.Contains(networkName, "container:") {
		return "container"
	}
	return "unknown"
}

func getVfDevicesByPlugin(networkName string, userCmdArgs []string) (string, string, error) {
	var vfNetdev string
	var err error

	network := getDockerNetworkResourceForName(networkName)
	if network == nil {
		return "", "", fmt.Errorf("Network not found")
	}

	pfNetdevName := network.Options["netdevice"]
	if pfNetdevName == "" {
		return "", "", fmt.Errorf("Netdevice invalid configuration")
	}

	vf := getUserVf(userCmdArgs)
	if vf != "" {
		vfNetdev, err = allocateSpecificVf(pfNetdevName, vf)
		if err != nil {
			return "", "", err
		}
	} else {
		vfNetdev, err = allocateVf(pfNetdevName)
		if err != nil {
			return "", "", err
		}
	}

	rdmaDev, err := rdmamap.GetRdmaDeviceForNetdevice(vfNetdev)
	if err != nil {
		return "", "", err
	}
	return vfNetdev, rdmaDev, nil
}

func getContainerId(networkName string) string {

	containerId := strings.Split(networkName, ":")
	if len(containerId) < 2 {
		return ""
	}
	return containerId[1]
}

// Returns rdma device for a netdev provisioned in a container
func getVfDevicesByContainer(networkName string) (string, string, error) {
	var rdmadev string

	containerId := getContainerId(networkName)
	if containerId == "" {
		return "", "", fmt.Errorf("Invalid container id format")
	}

	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	nsHandle, err := netns.GetFromDocker(containerId)
	if err != nil {
		fmt.Println("Invalid container id: ", containerId)
		return "", "", fmt.Errorf("Invalid container id")
	}
	originalHandle, err := netns.Get()
	if err != nil {
		fmt.Println("Fail to get handle of current net ns", err)
		return "", "", fmt.Errorf("Fail to get handle of current net ns")
	}
	netns.Set(nsHandle)

	ifaces, err := net.Interfaces()
	if err != nil {
		netns.Set(originalHandle)
		return "", "", fmt.Errorf("Fail to get interfaces")
	}
	for _, iface := range ifaces {
		if iface.Name == "lo" {
			continue
		}

		/* We currently support only first rdma device */
		rdmadev, err = rdmamap.GetRdmaDeviceForNetdevice(iface.Name)
		if err == nil && rdmadev != "" {
			break
		}
	}
	netns.Set(originalHandle)
	return "", rdmadev, nil
}

func allocateVfForNetwork(userCmdArgs []string) (string, string, error) {
	networkName := getNonDefaultNetwork(userCmdArgs)
	if networkName == "" {
		return "", "", fmt.Errorf("Invalid network information")
	}

	pluginType := getNetworkType(networkName)
	if pluginType == "plugin" {
		return getVfDevicesByPlugin(networkName, userCmdArgs)
	} else {
		return getVfDevicesByContainer(networkName)
	}
	return "", "", fmt.Errorf("Invalid/Unknown network information")
}

func stipNonDockerUserArgs(userCmdArgs []string) []string {
	var output []string

	privArgs := []string{"--vf="}
	var found bool

	for _, item := range userCmdArgs {
		found = false
		for _, v := range privArgs {
			found = strings.Contains(item, v)
			if found == true {
				break
			}
		}
		if found == false {
			output = append(output, item)
		}
		//userCmdArgs = append(userCmdArgs[:i], userCmdArgs[i+1:]...)
	}
	return output
}

func buildUserCmd(userCmdArgs []string) ([]string, error) {
	var runCmds []string
	var charDevCmdArgs []string

	runCmds = append(runCmds, "docker")
	runCmds = append(runCmds, "run")

	netDev, rdmaDev, err := allocateVfForNetwork(userCmdArgs)
	if err != nil {
		return nil, err
	}

	/* If netdev is setup outside for this container, there is
	 * no need to pass mac address param.
	 */
	if netDev != "" {
		handle, err := netlink.LinkByName(netDev)
		if err != nil {
			return nil, err
		}
		netAttr := handle.Attrs()
		macAddr := netAttr.HardwareAddr.String()
		macAddrArg := "--mac-address=" + macAddr
		runCmds = append(runCmds, macAddrArg)
	}

	if rdmaDev != "" {
		charDevs := rdmamap.GetRdmaCharDevices(rdmaDev)
		if len(charDevs) != 0 {
			charDevCmdArgs = toCharDevCmdArgs(charDevs)
		}
		for _, devcmdArg := range charDevCmdArgs {
			runCmds = append(runCmds, devcmdArg)
		}

		runCmds = append(runCmds, "--cap-add=IPC_LOCK")
	}

	output := stipNonDockerUserArgs(userCmdArgs)

	for _, usrCmdArg := range output {
		runCmds = append(runCmds, usrCmdArg)
	}
	return runCmds, nil
}

func execUserRunCmd(userCmdArgs []string) {
	newCmd, err := buildUserCmd(userCmdArgs)
	if err != nil {
		fmt.Println("Fail to run docker container. Error= ", err)
		return
	}

	shellCmd := exec.Command("docker")
	shellCmd.Args = newCmd
	shellCmd.Stdout = os.Stdout
	shellCmd.Stdin = os.Stdin
	shellCmd.Stderr = os.Stderr
	shellCmd.Run()
}

func execRunCmd(cmd *cobra.Command, args []string) {
	if len(os.Args) <= 2 {
		cmd.HelpFunc()(cmd, os.Args)
		return
	}
	execUserRunCmd(os.Args[2:])
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Wrapper to docker run <command>",
	Run:   execRunCmd,
	// Ignore the errors for other command line arguments
	// that this program doesn't know about.
	// Refer https://github.com/spf13/cobra/pull/284
	// Refer https://github.com/spf13/cobra/pull/662/commits/96853a4e2c2716ef0059db31d147ab7e42a89d93#diff-2fc2009ba1969a36b69136d7fb7b2072R1690
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

func init() {
	vfFlags := runCmd.Flags()
	vfFlags.IntVarP(&vfUserArg, "vf", "v", 0, "vf index")
}
