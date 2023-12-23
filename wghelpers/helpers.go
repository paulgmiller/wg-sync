package wghelpers

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type wghelper struct {
	d *wgtypes.Device
}

func (wg *wghelper) PublicKey() string { return wg.d.PublicKey.String() }
func (wg *wghelper) ListenPort() int   { return wg.d.ListenPort }

func (wg *wghelper) LookupPeer(publickey string) (string, bool) {
	return "", false
}

func (wg *wghelper) Peers() []wgtypes.Peer {
	return wg.d.Peers
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
func (wg *wghelper) AddPeer(publickey, cidr string) error {
	return nil
}

func SavePeers(name string, peers []wgtypes.PeerConfig) error {
	wg, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer wg.Close()

	return wg.ConfigureDevice(name, wgtypes.Config{
		Peers: peers,
		//	ReplacePeers: true
	})
}
