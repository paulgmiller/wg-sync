package nethelpers

import (
	"io"
	"log"
	"net"
	"net/http"

	"github.com/samber/lo"
)

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
