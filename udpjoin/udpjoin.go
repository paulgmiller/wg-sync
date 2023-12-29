package udpjoin

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"time"

	"github.com/paulgmiller/wg-sync/pretty"
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

type cidrAllocator interface {
	Allocate() (net.IP, error)
	CIDR() *net.IPNet
}

type wgDevice interface {
	PublicKey() string
	Endpoint() string
	LookupPeer(publickey string) (string, bool)
	AddPeer(publickey, cidr string) error
}

type joiner struct {
	auth  authorizer
	alloc cidrAllocator
	dev   wgDevice
}

func New(auth authorizer, alloc cidrAllocator, dev wgDevice) *joiner {
	return &joiner{
		auth:  auth,
		alloc: alloc,
		dev:   dev,
	}
}

// TODO listen on all ips
func (j *joiner) HaddleJoins(ctx context.Context) error {
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
			// todo byte fields rather than json just to make it really tight? binary/reader/writer? proto
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
			jResp, err := j.GenerateResponse(jreq)
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

func (j *joiner) GenerateResponse(jreq Request) (Response, error) {

	assignedip, found := j.dev.LookupPeer(jreq.PublicKey)
	if found { //should we also check that the ip is the same?
		log.Printf("peer %s already exists", jreq.PublicKey)
	} else {
		ip, err := j.alloc.Allocate()
		if err != nil {
			//not nice to not tell them sorry? But then we need an error protocol
			return Response{}, err
		}
		assignedip = ip.String() + "/32"
		//if we crash here we lose the ip. Combine allocator and wg device?
		//so that allocate takes a public key and adds the peer
		//wierd to add the slash /32.
		j.dev.AddPeer(jreq.PublicKey, assignedip)
	}

	//add the peer to us before we return anything

	//cidr, err := nethelpers.GetWireGaurdCIDR(j.dev.Name)

	return Response{
		Assignedip: assignedip,
		Peers: []pretty.Peer{
			{
				PublicKey: j.dev.PublicKey(),
				//is the right ting to do?
				AllowedIPs: j.alloc.CIDR().String(), //too much throttle down to /32?
				Endpoint:   j.dev.Endpoint(),
			},
		},
	}, nil
}
