package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paulgmiller/wg-sync/pretty"
	"github.com/paulgmiller/wg-sync/token"
	"github.com/paulgmiller/wg-sync/udpjoin"
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

var password string

func init() {
	serveCmd.Flags().StringVarP(&password, "password", "p", "", "use a dumb password (insercure)")
	rootCmd.AddCommand(serveCmd)
	//probably have to pass in public ip and maye cidr?
}

type cidrAllocatorImpl struct{}

func (c cidrAllocatorImpl) Allocate() (net.IP, error) {
	return net.ParseIP("10.0.0.100"), nil
}

func (c cidrAllocatorImpl) CIDR() *net.IPNet {
	_, net, err := net.ParseCIDR("10.0.0.0/24")
	if err != nil {
		panic(err)
	}
	return net
}

// this is for testing please don't use
type dumbpassword string

func (p dumbpassword) Validate(token string) error {
	if string(p) != token {
		return fmt.Errorf("fool %s is not the password")
	}
	return nil
}

func serve(cmd *cobra.Command, args []string) error {
	t := token.New()

	mux := http.NewServeMux()
	mux.HandleFunc("/peers", Peers)
	mux.Handle("/token", t)

	srv := http.Server{Addr: ":8888", Handler: mux}

	//todo gracefully shut both servers  down.
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	//get this lazily for each add.
	wg, err := wghelpers.GetDevice()
	if err != nil {
		return err
	}
	alloc := cidrAllocatorImpl{}
	joiner := udpjoin.New(t, alloc, wg)
	if password != "" {
		joiner = udpjoin.New(dumbpassword(password), alloc, wg)
	}
	err = joiner.HaddleJoins(ctx)
	if err != nil {
		log.Printf("udp handler exited with %s", err)
	}
	log.Printf("up and seving")
	<-ctx.Done()
	log.Printf("got term signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		// Error from closing listeners, or context timeout:
		log.Printf("HTTP server Shutdown: %v", err)
	}

	return err
}

func Peers(resp http.ResponseWriter, req *http.Request) {
	d0, err := wghelpers.GetDevice()
	if err != nil {
		log.Printf("error getting wg device %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := pretty.Yaml(resp, d0.Peers()...); err != nil {
		log.Printf("error marsalling peers %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
}
