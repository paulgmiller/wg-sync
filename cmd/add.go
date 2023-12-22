/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/paulgmiller/wg-sync/udpjoin"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "sends a join request to a listening wg-sync server",
	Long:  `sends a joint request then takes the returned assigned ip and peer and updates wg config`,
	RunE:  add,
}

func init() {
	rootCmd.AddCommand(addCmd)
	joinServer = *addCmd.Flags().StringP("server", "s", "127.0.0.1"+defaultJoinPort, "server ip  to send udp request to")
}

var joinServer string

func add(cmd *cobra.Command, args []string) error {
	/*d0, err := wghelpers.GetDevice()
	if err != nil {
		return err
	}*/
	jreq := udpjoin.Request{
		PublicKey: "DEADBEEFDEADBEEF", //d0.PublicKey.String(),
		AuthToken: "TOTALLYSECRET",
	}

	resp, err := udpjoin.Send(joinServer, jreq)
	if err != nil {
		return err
	}
	log.Printf("got %s", resp.Assignedip)

	jreq2 := udpjoin.Request{
		PublicKey: "amMRWDvsLUmNHn52xer2yl/UaAkXnDrd/HxUTRkEGXc=", //d0.PublicKey.String(),
		AuthToken: "TOTALLYSECRET",
	}
	resp, err = udpjoin.Send(joinServer, jreq2)
	if err != nil {
		return err
	}
	log.Printf("got %s", resp.Assignedip)
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
