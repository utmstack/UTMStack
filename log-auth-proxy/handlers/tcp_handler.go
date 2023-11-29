package handlers

import (
	"bufio"
	"log"
	"net"
	"strings"

	"github.com/utmstack/UTMStack/log-auth-proxy/logservice"
	"github.com/utmstack/UTMStack/log-auth-proxy/middleware"
	"github.com/utmstack/UTMStack/log-auth-proxy/model"
)

func HandleRequest(conn net.Conn, interceptor *middleware.LogAuthInterceptor, logOutputService *logservice.LogOutputService) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()

		parts := strings.Split(message, ",LOG:")
		if len(parts) != 2 {
			conn.Write([]byte("INVALID FORMAT expecting AUTH:<token>,LOG:<log>\n"))
			continue
		}

		token := strings.TrimPrefix(parts[0], "AUTH:")
		logData := parts[1]

		if !interceptor.AuthService.IsConnectionKeyValid(token) {
			conn.Write([]byte("Authentication FAIL\n"))
			continue
		}

		go logOutputService.SendLog(model.WebHookGithub, logData)
		conn.Write([]byte("RECEIVED\n"))

	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading from connection:", err.Error())
	}
}
