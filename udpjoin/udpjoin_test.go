package udpjoin

import (
	"context"
	"log"
	"net"
	"testing"
)

type boolValidator struct {
	e error
}

func (b *boolValidator) Validate(token string) error {
	return b.e
}

type fakeDevice struct{}

func (*fakeDevice) CIDR() *net.IPNet {
	_, net, _ := net.ParseCIDR("10.1.0.0/24")
	return net
}

func (*fakeDevice) Allocate() (net.IP, error) {
	return net.IPv4(10, 1, 0, 5), nil
}

func (*fakeDevice) PublicKey() string { return "ABCDEFABCDEF" }
func (*fakeDevice) Endpoint() string  { return "20.20.20.20:5000" }
func (*fakeDevice) LookupPeer(publickey string) (string, bool) {
	return "", false
}
func (*fakeDevice) AddPeer(publickey, cidr string) error {
	return nil
}

var _ cidrAllocator = &fakeDevice{}
var _ wgDevice = &fakeDevice{}

func TestJoin(t *testing.T) {

	d := &fakeDevice{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	j := New(&boolValidator{}, d, d)

	if err := j.HaddleJoins(ctx); err != nil {
		t.Fatalf("can't start listner %s", err)
	}

	jreq := Request{
		PublicKey: "DEADBEEFDEADBEEF", //d0.PublicKey.String(),
		AuthToken: "HOKEYPOKEYSMOKEY",
	}

	resp, err := Send("127.0.0.1:5000", jreq)
	if err != nil {
		t.Fatalf("failed send %s", err)
	}

	log.Printf("got %s", resp.Assignedip)

	if resp.Assignedip != net.IPv4(10, 1, 0, 5).String() {
		t.Fatalf("ddidn't get back expeced ip")
	}

	if len(resp.Peers) != 1 {
		t.Fatal("didn't get expected peers")
	}
	if resp.Peers[0].AllowedIPs != "10.1.0.0/24" {
		t.Fatalf("ddidn't get back expeced allowed ips %s", resp.Peers[0].AllowedIPs)
	}

	if resp.Peers[0].Endpoint != "20.20.20.20:5000" {
		t.Fatalf("didn't get back expeced endpoint: %s", resp.Peers[0].Endpoint)
	}

	jreq2 := Request{
		PublicKey: "amMRWDvsLUmNHn52xer2yl/UaAkXnDrd/HxUTRkEGXc=", //d0.PublicKey.String(),
		AuthToken: "TOTALLYSECRET",
	}
	resp, err = Send("127.0.0.1:5000", jreq2)
	if err != nil {
		t.Fatalf("failed send %s", err)
	}
	log.Printf("got %s", resp.Assignedip)

}
