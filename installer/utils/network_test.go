package utils

import (
	"fmt"
	"testing"
)

func TestGetMainIP(t *testing.T) {
	ip, err := GetMainIP()
	if err != nil{
		t.Error(err)
	}

	iface, err := GetMainIface(ip)
	if err != nil{
		t.Error(err)
	}

	fmt.Println("Main Iface is: ", iface)
}
