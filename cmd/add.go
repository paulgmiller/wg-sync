/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/paulgmiller/wg-sync/nethelpers"
	"github.com/paulgmiller/wg-sync/pretty"
	"github.com/paulgmiller/wg-sync/wghelpers"
	"github.com/rb-go/namegen"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add current as a peer to sync list",
	Long:  `pulls the current device's public key and adds it to the sync list merging with others`,
	RunE:  add,
}

func init() {
	rootCmd.AddCommand(addCmd)
	server = addCmd.Flags().BoolP("server", "s", false, "publish as a server which means we add an endpoint")
}

func add(cmd *cobra.Command, args []string) error {

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
