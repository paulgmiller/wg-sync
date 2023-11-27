package nethelpers

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/samber/lo"
)

func GetWireGaurdCIDR(interfacename string) (net.Addr, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	wginterface, found := lo.Find(ifaces, func(iface net.Interface) bool { return iface.Name == interfacename })
	if !found {
		return nil, err
	}

	addrs, err := wginterface.Addrs()
	if err != nil {
		return nil, err
	}
	return addrs[0], nil
}

func GetOutboundIP() string {
	resp, err := http.Get("https://ifconfig.me")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

var NotFound error = errors.New("PublicIPNotFound")

// this won't work with azure vms where the public up is dont thgouh nat
func FindPublicIp() (net.Addr, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var publicaddrs []net.Addr
	for _, a := range addrs {
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			return nil, err
		}

		//come back and support ipv6
		if ip.IsPrivate() {
			log.Printf("private %s", ip)
			continue
		}
		if ip.IsLoopback() {
			continue
		}

		if ip.To4() == nil {
			continue
		}
		publicaddrs = append(publicaddrs, a)

	}
	if len(publicaddrs) == 0 {
		return nil, NotFound
	}
	return publicaddrs[0], nil

}
