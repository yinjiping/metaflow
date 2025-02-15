package utils

import (
	"net"
	"testing"
)

func TestIPv4ToBinary(t *testing.T) {
	expect := "10000001000101000000000111111111"
	if ret := IPv4ToBinary(IpToUint32(net.ParseIP("129.20.1.255").To4())); ret != expect {
		t.Errorf("IPv4ToBinary处理不正确，expect %v, return %v", expect, ret)
	}
}

func TestIPv6ToBinary(t *testing.T) {
	expect := "00000000000000010000000000100011000001000101011001111000100110100000000000000000000000000000000000000000000000001001100000011111"
	if ret := IPv6ToBinary(net.ParseIP("1:23:456:789a:0::981f")); ret != expect {
		t.Errorf("IPv6ToBinary处理不正确，expect %v, return %v", expect, ret)
	}
}
