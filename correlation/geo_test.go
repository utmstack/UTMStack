package main

import (
	"log"
	"net"
	"testing"

	"github.com/AtlasInsideCorp/UTMStackCorrelationEngine/geo"
	"github.com/levigross/grequests"
)

func TestIsLocal(t *testing.T) {
	ip := net.ParseIP("192.168.0.1")
	if !geo.IsLocal(ip) {
		t.Fatalf("%s should be local", ip)
	}
}

func TestMyExternalIp(t *testing.T) {
	resp, err := grequests.Get("http://myexternalip.com/raw", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Log(resp.String())
}

func TestGetExternal(t *testing.T) {
	ip := geo.GetExternal()
	t.Log(ip)
	if ip == nil {
		log.Fatalln("external ip can't be nil")
	}

}
func TestGetExternalOnce(t *testing.T) {
	ip := geo.GetExternalOnce()
	t.Log(ip)
	if ip == nil {
		log.Fatalln("external ip can't be nil")
	}

}
