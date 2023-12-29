package wghelpers

import (
	"net"
	"testing"
)

func TestIncrement(t *testing.T) {

	first, _, err := net.ParseCIDR("10.0.0.0/24")
	if err != nil {
		t.Fatal(err.Error())
	}

	inc(first)
	if first.String() != "10.0.0.1" {
		t.Fatalf("%s wasn't as expected ", first.String())
	}

}
