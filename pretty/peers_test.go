package pretty

import (
	"fmt"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestToConfig(t *testing.T) {
	key, _ := wgtypes.GenerateKey()

	fmt.Println(key)
	p := Peer{
		PublicKey:  key.String(),
		AllowedIPs: "10.0.0.0/24,10.0.2.0/24",
	}
	pcfg, err := p.ToConfig()
	if err != nil {
		t.Fatalf("error %s", err)
	}
	if len(pcfg.AllowedIPs) != 2 {
		t.Fatalf("wrong number of allowed ips")
	}

}
