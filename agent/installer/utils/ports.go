package utils

import (
	"fmt"
	"net"
	"time"
)

func IsPortReachable(ip, port string) error {
	var (
		conn net.Conn
		err  error
	)

	for i := 0; i < 3; i++ {
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ip, port), 5*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(5 * time.Second)
	}

	return err
}
