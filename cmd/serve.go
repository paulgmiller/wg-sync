package cmd

import (
	"log"
	"net/http"

	"github.com/paulgmiller/wg-sync/pretty"
	"github.com/paulgmiller/wg-sync/wghelpers"
	"github.com/spf13/cobra"
)

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

	/*l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Echo all incoming data.
			io.Copy(c, c)
			// Shut down the connection.
			c.Close()
		}(conn)
	*/

	return http.ListenAndServe(":8888", nil)

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
