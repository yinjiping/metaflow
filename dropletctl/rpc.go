/*
 * rpc is a subcommands of droplet-ctl
 * which pulls policy information from controller by rpc.
 * now it contains 3 subcommands:
 *   flowAcls     get flowAcls from controller
 *   ipGroups     get ipGroups from controller
 *   platformData get platformData from controller
 */
package dropletctl

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"gitlab.x.lan/yunshan/droplet-libs/logger"
	"gitlab.x.lan/yunshan/droplet/config"
	"gitlab.x.lan/yunshan/message/trident"
)

type CmdExecute func(response *trident.SyncResponse)

var configPath string

func SetConfigPath(path string) {
	configPath = path
}

func regiterCommand() []*cobra.Command {
	platformDataCmd := &cobra.Command{
		Use:   "platformData",
		Short: "get platformData from controller, press Ctrl^c to end it",
		Run: func(cmd *cobra.Command, args []string) {
			initCmd(platformData)
		},
	}
	ipGroupsCmd := &cobra.Command{
		Use:   "ipGroups",
		Short: "get ipGroups from controller, press Ctrl^c to end it",
		Run: func(cmd *cobra.Command, args []string) {
			initCmd(ipGroups)
		},
	}
	flowAclsCmd := &cobra.Command{
		Use:   "flowAcls",
		Short: "get flowAcls from controller, press Ctrl^c to end it",
		Run: func(cmd *cobra.Command, args []string) {
			initCmd(flowAcls)
		},
	}

	commands := []*cobra.Command{platformDataCmd, ipGroupsCmd, flowAclsCmd}
	return commands
}

func RegisterRpcCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "rpc",
		Short: "pull policy from controller by rpc",
	}
	cmds := regiterCommand()
	for _, handler := range cmds {
		root.AddCommand(handler)
	}

	return root
}

func initCmd(cmd CmdExecute) {
	cfg := config.Load(configPath)
	cfg.LogLevel = "error"
	logger.InitLog(cfg.LogFile, cfg.LogLevel)

	controllers := make([]net.IP, 0, len(cfg.ControllerIps))
	for _, ipString := range cfg.ControllerIps {
		ip := net.ParseIP(ipString)
		controllers = append(controllers, ip)
	}

	synchronizer := config.NewRpcConfigSynchronizer(controllers, cfg.ControllerPort)
	synchronizer.Register(func(response *trident.SyncResponse) {
		cmd(response)
		fmt.Println("press Ctrl^c to end it !!")
	})

	synchronizer.Start()

	wait := make(chan os.Signal)
	signal.Notify(wait, os.Interrupt)
	if sig := <-wait; sig != os.Interrupt {
		fmt.Println("press Ctrl^c to end it !!")
	}
}

func jsonFormat(index int, v interface{}) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		fmt.Println("json encode failed")
	}
	fmt.Printf("\t%v: %s\n", index, jsonBytes)
}

func flowAcls(response *trident.SyncResponse) {
	if flowAcls := response.GetFlowAcls(); flowAcls != nil {
		fmt.Println("acl data:")
		for index, entry := range flowAcls {
			jsonFormat(index+1, entry)
		}
	}
}

func ipGroups(response *trident.SyncResponse) {
	if plarformData := response.GetPlatformData(); plarformData != nil {
		if ipGroups := plarformData.GetIpGroups(); ipGroups != nil {
			fmt.Println("ipGroups data:")
			for index, entry := range ipGroups {
				jsonFormat(index+1, entry)
			}
		}
	}
}

func platformData(response *trident.SyncResponse) {
	if plarformData := response.GetPlatformData(); plarformData != nil {
		if interfaces := plarformData.GetInterfaces(); interfaces != nil {
			fmt.Println("platform data:")
			for index, entry := range interfaces {
				jsonFormat(index+1, entry)
			}
		}
	}
}
