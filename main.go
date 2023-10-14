package main

import (
	"log"

	"golang.zx2c4.com/wireguard/wgctrl"
)

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
	for _, d := range devices {
		log.Printf("%s %s %s\n", d.Name, d.PublicKey, d.ListenPort)
		for _, p := range d.Peers {
			log.Printf("  %s %s %s\n", p.PublicKey, p.AllowedIPs, p.Endpoint)
		}
	}

}
