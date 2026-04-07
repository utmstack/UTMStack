package utils

import (
	"fmt"
	"net"
	"time"
)

func ArePortsReachable(ip string, ports ...string) error {
	var conn net.Conn
	var err error

external:
	for _, port := range ports {
		for i := 0; i < 3; i++ {
			conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ip, port), 5*time.Second)
			if err == nil {
				conn.Close()
				continue external
			}
			time.Sleep(5 * time.Second)
		}
		if err != nil {
			return Logger.ErrorF("cannot connect to %s on port %s: %v", ip, port, err)
		}
	}

	return nil
}
