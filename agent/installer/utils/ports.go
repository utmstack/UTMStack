package utils

import (
	"fmt"
	"net"
	"time"
)

func IsPortOpen(ip, port string) bool {
	for i := 0; i < 3; i++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ip, port), 5*time.Second)
		if err == nil {
			defer conn.Close()
			return true
		}
		time.Sleep(5 * time.Second)
	}
	return false
}
