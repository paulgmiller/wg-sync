package pretty

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/go-ini/ini"
	"github.com/samber/lo"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/yaml.v2"
)

type Peer struct {
	PublicKey  string `yaml:"PublicKey,omitempty"`
	AllowedIPs string `yaml:"AllowedIPs,omitempty"`
	Endpoint   string `yaml:"Endpoint,omitempty"`
	//Server     bool   `yaml:"Server,omitempty"`
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

func Ini(w io.Writer, peers ...Peer) error {
	iniFile := ini.Empty()
	for _, p := range peers {
		sec, err := iniFile.NewSection("PEER")
		if err != nil {
			return err
		}
		//tecnically new key returns an err
		sec.NewKey("PublicKey", p.PublicKey)
		sec.NewKey("AllowedIPs", p.AllowedIPs)
		if p.Endpoint != "" {
			sec.NewKey("AllowedIPs", p.AllowedIPs)
		}
	}
	_, err := iniFile.WriteTo(w)
	return err
}

func Yaml(w io.Writer, peers ...wgtypes.Peer) error {
	enc := yaml.NewEncoder(w)
	var prettypeers []Peer
	for _, p := range peers {
		prettypeers = append(prettypeers, New(p))
	}
	return enc.Encode(prettypeers)
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
