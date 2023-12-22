package udpjoin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/paulgmiller/wg-sync/nethelpers"
	"github.com/paulgmiller/wg-sync/pretty"
	"github.com/paulgmiller/wg-sync/wghelpers"
	"github.com/samber/lo"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const defaultJoinPort = ":5000"

type Request struct {
	PublicKey string
	AuthToken string
}

type Response struct {
	Assignedip string
	Peers      []pretty.Peer
}

type cidrAllocator interface {
	Allocate() (net.IP, error)
}

const sendTimeout = time.Second * 3

func Send(server string, jReq Request) (Response, error) {

	conn, err := net.Dial("udp", server)
	if err != nil {
		return Response{}, err
	}

	if err := conn.SetDeadline(time.Now().Add(sendTimeout)); err != nil {
		return Response{}, err
	}

	log.Printf("dialing %s, %s", server, conn.LocalAddr().String())
	defer conn.Close()
	err = json.NewEncoder(conn).Encode(jReq)
	if err != nil {
		return Response{}, err
	}

	var jResp Response
	err = json.NewDecoder(conn).Decode(&jResp)

	return jResp, err
}

type authorizer interface {
	Validate(token string) error
}

type joiner struct {
	lock sync.Mutex
	auth authorizer
}

func New(a authorizer) *joiner {
	return &joiner{}
}

// TODO listen on all ips
func (j *joiner) HaddleJoins(ctx context.Context, alloc cidrAllocator) error {
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
		defer conn.Close()
		buf := make([]byte, 4096)
		for {
			//how big should we be? will we go over multiple packets?
			n, remoteAddr, err := conn.ReadFromUDP(buf) //has to be this ratehr than desrialize because we need the remote addr or we get  write: destination address required
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Printf("Failed to read from udp: %s", err)
				}
				continue
			}
			// Deserialize the JSON data into a Message struct
			// todo byte fields rather than json just to make it really tight?
			var jreq Request
			err = json.Unmarshal(buf[:n], &jreq)
			if err != nil {
				log.Printf("Failed to unmarshal: %s, %s", buf, err)

				continue
			}

			if err := j.auth.Validate(jreq.AuthToken); err != nil {
				log.Printf("bad auth token from %v, %s", remoteAddr, jreq.PublicKey)
				//ban them for a extended period? Just backoff?
				continue
			}

			log.Printf("got join request from %v, %s", remoteAddr, jreq.PublicKey)
			jResp, err := j.GenerateResponse(jreq, alloc)
			if err != nil {
				log.Printf("Failed to generate response %s", err)
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
	return nil

}

func (j *joiner) GenerateResponse(jreq Request, alloc cidrAllocator) (Response, error) {
	j.lock.Lock()
	defer j.lock.Unlock()

	d0, err := wghelpers.GetDevice()
	if err != nil {
		return Response{}, err
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
			return Response{}, err
		}
		asssignedip = ip.String()
	}

	//ad the peer to us before we return anything

	cidr, err := nethelpers.GetWireGaurdCIDR(d0.Name)
	if err != nil {
		return Response{}, err
	}

	//ip, cinet.ParseCIDR(cidr.String())

	return Response{
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
