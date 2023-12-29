package wghelpers

import (
	"fmt"
	"net"

	"github.com/paulgmiller/wg-sync/nethelpers"
	"github.com/paulgmiller/wg-sync/pretty"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type wghelper struct {
	d       *wgtypes.Device
	cidr    *net.IPNet
	firstip net.IP
}

func (wg *wghelper) PublicKey() string { return wg.d.PublicKey.String() }

// allow passing in get outbound ip/
func (wg *wghelper) Endpoint() string {
	return fmt.Sprintf("%s:%d", nethelpers.GetOutboundIP(), wg.d.ListenPort)
}

func (wg *wghelper) LookupPeer(publickey string) (string, bool) {
	return "", false
}

func (wg *wghelper) CIDR() *net.IPNet {
	return wg.cidr
}

func (wg *wghelper) Allocate() (net.IP, error) {
	var candidate net.IP
	copy(candidate, wg.firstip)
	for {
		if !wg.cidr.Contains(candidate) {
			return net.IP{}, fmt.Errorf("no more ips left in %s", wg.cidr)
		}

		for _, p := range wg.d.Peers {
			for _, used := range p.AllowedIPs {
				if !used.Contains(candidate) {
					return candidate, nil
				}
			}
		}
		inc(candidate)
	}
}

func inc(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func WithCidr(cidr string) (*wghelper, error) {
	ip, mask, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	wghelper, err := GetDevice()
	if err != nil {
		return nil, err
	}
	wghelper.cidr = mask
	wghelper.firstip = ip
	return wghelper, nil
}

func GetDevice() (*wghelper, error) {

	wg, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	defer wg.Close()
	devices, err := wg.Devices()
	if err != nil {
		return nil, err
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no wireguard devices found")
	}

	if len(devices) != 1 {
		return nil, fmt.Errorf("multiple devices: TODO specify one as arg")
	}

	return &wghelper{d: devices[0]}, nil
}

func (wg *wghelper) SetIP(ip string) error {
	//in linux exe ip i
	//exec.Command("ip" )
	w, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer w.Close()

	return nil
}

func (wg *wghelper) AddPeer(publickey, cidr string) error {
	wgc, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer wgc.Close()
	peer, err := pretty.Peer{
		PublicKey:  publickey,
		AllowedIPs: cidr,
	}.ToConfig()
	if err != nil {
		return err
	}
	return wgc.ConfigureDevice(wg.d.Name, wgtypes.Config{

		Peers: []wgtypes.PeerConfig{
			peer,
		},
		//	ReplacePeers: true
	})
}

func (wg *wghelper) Peers() []wgtypes.Peer {
	return wg.d.Peers
}
