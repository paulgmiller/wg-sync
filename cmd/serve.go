package cmd

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/paulgmiller/wg-sync/pretty"
	"github.com/paulgmiller/wg-sync/wghelpers"
	"github.com/spf13/cobra"
)

const defaultJoinPort = ":5000"

// addCmd represents the add command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "act as a gateway server",
	Long:  `serve up both peers over yaml for syncing and a udp add connection`, //sould we also do dns?
	RunE:  serve,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	//probably have to pass in public ip and maye cidr?
}

func serve(cmd *cobra.Command, args []string) error {

	http.HandleFunc("/peers", Peers)
	cmd.Context()
	//todo pass a context? figure out cancelation?
	HaddleJoins(cmd.Context())

	return nil
	//return http.ListenAndServe(":8888", nil)

}

type joinRequest struct {
	PublicKey string
}

type joinResponse struct {
	Assignedip string
	Peer       []pretty.Peer
}

func HaddleJoins(ctx context.Context) {
	udpaddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+defaultJoinPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waitig for joins on %s", udpaddr.String())
	for {

		conn, err := net.ListenUDP("udp", udpaddr)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Got  joins on %s", conn.RemoteAddr())

		reader := io.TeeReader(conn, os.Stdout)
		var jreq joinRequest
		err := json.NewDecoder(reader).Decode()
		if err != nil {
			log.Printf()
			conn.Close()
			return
		}
		conn.Close()
	}

}

func Peers(resp http.ResponseWriter, req *http.Request) {
	d0, err := wghelpers.GetDevice()
	if err != nil {
		log.Printf("error getting wg device %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := pretty.Yaml(resp, d0.Peers...); err != nil {
		log.Printf("error marsalling peers %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
}
