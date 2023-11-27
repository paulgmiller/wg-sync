package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/paulgmiller/wg-sync/nethelpers"
	"github.com/paulgmiller/wg-sync/pretty"
	"github.com/paulgmiller/wg-sync/wghelpers"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/peers", Peers)
	srv := http.Server{Addr: ":8888", Handler: mux}

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	err := HaddleJoins(ctx, cidrAllocatorImpl{})
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

type cidrAllocator interface {
	Allocate() (net.IP, error)
}

type cidrAllocatorImpl struct{}

func (c cidrAllocatorImpl) Allocate() (net.IP, error) {
	return net.ParseIP("10.0.0.100"), nil
}

var lock sync.Mutex

func HaddleJoins(ctx context.Context, alloc cidrAllocator) error {
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
			buf := make([]byte, 4096)                   //how big should we be? will we go over multiple packets?
			n, remoteAddr, err := conn.ReadFromUDP(buf) //has to be this ratehr than desrialize because we need the remote addr or we get  write: destination address required
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

			//obviously bad.
			if jreq.AuthToken != "HOKEYPOKEYSMOKEY" {
				log.Printf("bad auth token from %v, %s", remoteAddr, jreq.PublicKey)
				//ban them for a extended period?
				continue
			}

			log.Printf("got join request from %v, %s", remoteAddr, jreq.PublicKey)
			jResp, err := GenerateResponse(jreq, alloc)
			if err != nil {
				log.Printf("Failed to generate response %s", err)
				//ban them for a extended period?
				continue
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

func GenerateResponse(jreq joinRequest, alloc cidrAllocator) (joinResponse, error) {
	lock.Lock()
	defer lock.Unlock()

	d0, err := wghelpers.GetDevice()
	if err != nil {
		return joinResponse{}, err
	}

	var asssignedip string
	existing, found := lo.Find(d0.Peers, func(p wgtypes.Peer) bool { return p.PublicKey.String() == jreq.PublicKey })
	if found { //should we also check that the ip is the same?
		log.Printf("peer %s already exists", jreq.PublicKey)
		asssignedip = existing.AllowedIPs[0].String()
	} else {
		ip, err := alloc.Allocate()
		if err != nil {
			//not nice to not tell them sorry? But then we need an error protocol
			return joinResponse{}, err
		}
		asssignedip = ip.String()
	}

	//ad the peer to us before we return anything

	cidr, err := nethelpers.GetWireGaurdCIDR(d0.Name)
	if err != nil {
		return joinResponse{}, err
	}

	//ip, cinet.ParseCIDR(cidr.String())

	return joinResponse{
		Assignedip: asssignedip,
		Peers: []pretty.Peer{
			{
				PublicKey:  d0.PublicKey.String(),
				AllowedIPs: cidr.String(),                                                   //too much throttle down to /32?
				Endpoint:   fmt.Sprintf("%s:%d", nethelpers.GetOutboundIP(), d0.ListenPort), //just pass this in instead of trying to detect it?
			},
		},
	}, nil
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
