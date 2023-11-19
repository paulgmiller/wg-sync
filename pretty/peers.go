package pretty

import (
	"fmt"
	"net"
	"strings"

	"github.com/samber/lo"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Peer struct {
	PublicKey  string `yaml:"PublicKey,omitempty"`
	AllowedIPs string `yaml:"AllowedIPs,omitempty"`
	Endpoint   string `yaml:"Endpoint,omitempty"`
	Server     bool   `yaml:"Server,omitempty"`
}

//lets try https://raw.githubusercontent.com/paulgmiller/paulgmiller.github.io/master/peers.yaml

func New(p wgtypes.Peer) Peer {
	return Peer{
		PublicKey:  p.PublicKey.String(),
		AllowedIPs: strings.Join(lo.Map(p.AllowedIPs, func(item net.IPNet, _ int) string { return item.String() }), ","),
		//how do we know if these are public or temporary? is it fine to guess?
		//Endpoint: p.Endpoint.String(),
	}
}

func (p Peer) ToConfig() (wgtypes.PeerConfig, error) {
	pkey, err := wgtypes.ParseKey(p.PublicKey)
	if err != nil {
		return wgtypes.PeerConfig{}, err
	}
	if len(pkey) != wgtypes.KeyLen {
		return wgtypes.PeerConfig{}, fmt.Errorf("key length inocrrect %d should be %d", len(pkey), wgtypes.KeyLen)
	}

	var allowdedips []net.IPNet
	strIps := strings.Split(p.AllowedIPs, ",")
	for _, item := range strIps {
		ip, vnet, err := net.ParseCIDR(item)
		if err != nil {
			return wgtypes.PeerConfig{}, err
		}
		allowdedips = append(allowdedips, net.IPNet{IP: ip, Mask: vnet.Mask})
	}

	return wgtypes.PeerConfig{
		PublicKey:  wgtypes.Key(pkey),
		UpdateOnly: true,
		AllowedIPs: allowdedips,
	}, nil
}
