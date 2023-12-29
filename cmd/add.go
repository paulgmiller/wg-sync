/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/go-ini/ini"
	"github.com/paulgmiller/wg-sync/pretty"
	"github.com/paulgmiller/wg-sync/udpjoin"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "sends a join request to a listening wg-sync server",
	Long:  `sends a join request then takes the returned assigned ip and peer and generates a wq-quick config`,
	RunE:  add,
}

func init() {
	rootCmd.AddCommand(addCmd)
	joinServer = *addCmd.Flags().StringP("server", "s", "127.0.0.1"+defaultJoinPort, "server ip  to send udp request to")
	thetoken = *addCmd.Flags().StringP("token", "t", "empty", "otop autenticator token to send to server")
}

var joinServer, thetoken string

func add(cmd *cobra.Command, args []string) error {
	private, err := wgtypes.GenerateKey()
	if err != nil {
		return err
	}

	//todo reuse device if exists?
	//d0, err := wghelpers.GetDevice()

	jreq := udpjoin.Request{
		PublicKey: private.PublicKey().String(),
		AuthToken: thetoken,
	}

	resp, err := udpjoin.Send(joinServer, jreq)
	if err != nil {
		return err
	}
	log.Printf("got %+v", resp)

	conf := ini.Empty()

	iface, err := conf.NewSection("Interface")
	if err != nil {
		return err
	}
	iface.NewKey("PrivateKey", private.String())
	iface.NewKey("Address", resp.Assignedip)

	pretty.Ini(conf, resp.Peers...)

	conf.WriteTo(os.Stdout)

	return nil

}

/* old and busted

	resp, err := http.Get(cfgFile)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	peers := map[string]pretty.Peer{}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got %d from %s", resp.StatusCode, cfgFile)
	}
	decoder := yaml.NewDecoder(resp.Body)
	err = decoder.Decode(&peers)
	if err != nil {
		return err
	}

	d0, err := wghelpers.GetDevice()
	if err != nil {
		return err
	}

	me := pretty.Peer{
		PublicKey:  d0.PublicKey.String(),
		AllowedIPs: nethelpers.GetWireGaurdIP(d0.Name),
	}

	if lo.FromPtr(server) {
		me.Endpoint = fmt.Sprintf("%s:%d", nethelpers.GetOutboundIP(), d0.ListenPort)
	}

	peers[namegen.GetName(0)] = me

	stdout := yaml.NewEncoder(os.Stdout)
	return stdout.Encode(peers)
}
*/
