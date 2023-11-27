package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	srv := http.Server{Addr: ":8888"}

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	err := HaddleJoins(ctx)
	if err != nil {
		log.Printf("udp handler exited with %s", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		// Error from closing listeners, or context timeout:
		log.Printf("HTTP server Shutdown: %v", err)
	}

	return err
}

type joinRequest struct {
	PublicKey string
	AuthToken string
}

type joinResponse struct {
	Assignedip string
	Peers      []pretty.Peer
}

func HaddleJoins(ctx context.Context) error {
	udpaddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+defaultJoinPort)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		return err
	}
	log.Printf("Waiting for joins on %s", udpaddr.String())
	go func() {
		for {
			buf := make([]byte, 4096) //how big should we be? will we go over multiple packets?
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Printf("Failed to read from udp: %s", err)
				}
				return
			}
			// Deserialize the JSON data into a Message struct
			var jreq joinRequest
			err = json.Unmarshal(buf[:n], &jreq)
			if err != nil {
				log.Printf("Failed to unmarshal: %s, %s", buf, err)

				continue
			}

			log.Printf("got join request from %v, %s", remoteAddr, jreq.PublicKey)
			jResp := joinResponse{
				Assignedip: "10.0.0.100",
				Peers: []pretty.Peer{
					{
						PublicKey:  "amMRWDvsLUmNHn52xer2yl/UaAkXnDrd/HxUTRkEGXc=",
						AllowedIPs: "10.0.0.0/24",
					},
				},
			}
			respbuf, err := json.Marshal(jResp)
			if err != nil {
				log.Printf("Failed to enode: %s", err)
				continue
			}
			_, err = conn.WriteToUDP(respbuf, remoteAddr)
			if err != nil {
				log.Printf("Failed to send: %s, %s", buf, err)
				continue
			}

		}
	}()
	<-ctx.Done()
	conn.Close()
	log.Println("Listener closed")
	return nil

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
