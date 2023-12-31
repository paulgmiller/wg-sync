/*
Copyright © 2023 NAME HERE paul.miller@gmail.com
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/paulgmiller/wg-sync/pretty"
	"github.com/paulgmiller/wg-sync/wghelpers"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "wg-sync",
		Short: "syncs peers from a central url ",
		Long:  `replaces all peers with those from a central url`,
		RunE:  syncPeers,
	}
	cfgFile string
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "url", "", "config file (default is $HOME/.wg-sync.yaml)")
}

func syncPeers(cmd *cobra.Command, args []string) error {

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
		log.Fatal(err)
	}

	d, err := wghelpers.GetDevice()
	if err != nil {
		return err
	}

	for _, peer := range peers {
		if peer.PublicKey == d.PublicKey() {
			continue
		}
		//todo make sure we don't duplocate ips?
		if _, found := d.LookupPeer(peer.PublicKey); !found {
			//todo endpoint?
			d.AddPeer(peer.PublicKey, peer.AllowedIPs)
		}

	}

	return nil

}
