package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/rb-go/namegen"
	"github.com/samber/lo"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/yaml.v2"
)

type prettyPeer struct {
	PublicKey  string `yaml:"PublicKey,omitempty"`
	AllowedIPs string `yaml:"AllowedIPs,omitempty"`
	Endpoint   string `yaml:"Endpoint,omitempty"`
	//KeepAlive
}

func PrettyPeer(p wgtypes.Peer) prettyPeer {
	return prettyPeer{
		PublicKey:  base64.StdEncoding.EncodeToString(p.PublicKey[:]),
		AllowedIPs: strings.Join(lo.Map(p.AllowedIPs, func(item net.IPNet, _ int) string { return item.String() }), ","),
		//how do we know if these are public or temporary? is it fine to guess?
		//Endpoint: p.Endpoint.String(),
	}
}
func main() {
	wg, err := wgctrl.New()
	if err != nil {
		log.Fatal(err)
	}
	defer wg.Close()
	devices, err := wg.Devices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		log.Fatal("no wireguard devices found")
	}

	if len(devices) != 1 {
		log.Fatal("multiple devices: TODO specify one as arg")
	}

	d0 := devices[0]

	me := prettyPeer{
		PublicKey:  d0.PublicKey.String(),
		AllowedIPs: GetWireGaurdIP(d0.Name),
		Endpoint:   fmt.Sprintf("%s:%d", GetOutboundIP().String(), d0.ListenPort),
	}

	peers := map[string]prettyPeer{}
	peers[namegen.GetName(0)] = me

	for _, peer := range devices[0].Peers {
		peers[namegen.GetName(0)] = PrettyPeer(peer)
	}

	stdout := yaml.NewEncoder(os.Stdout)
	err = stdout.Encode(peers)
	if err != nil {
		log.Fatal(err)
	}

}

func GetWireGaurdIP(interfacename string) string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("can't get interfaces: %v", err)
	}

	wginterface, found := lo.Find(ifaces, func(iface net.Interface) bool { return iface.Name == interfacename })
	if !found {
		log.Fatalf("can't get interfaces: %v", err)
	}

	addrs, err := wginterface.Addrs()
	if err != nil {
		log.Fatalf("can't get interface addrs: %v", err)
	}
	return addrs[0].String()
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
