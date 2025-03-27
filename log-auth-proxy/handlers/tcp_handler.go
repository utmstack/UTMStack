package handlers

import (
	"bufio"
	"net"
	"strings"

	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/logservice"
	"github.com/utmstack/UTMStack/log-auth-proxy/middleware"
	"github.com/utmstack/UTMStack/log-auth-proxy/utils"
)

func HandleRequest(conn net.Conn, interceptor *middleware.LogAuthInterceptor, logOutputService *logservice.LogOutputService) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()

		parts := strings.Split(message, ",LOG:")
		if len(parts) != 2 {
			utils.Logger.ErrorF("INVALID FORMAT expecting AUTH:<token>,LOG:<log>")
			conn.Write([]byte("INVALID FORMAT expecting AUTH:<token>,LOG:<log>\n"))
			continue
		}

		token := strings.TrimPrefix(parts[0], "AUTH:")
		logData := parts[1]

		if !interceptor.AuthService.IsConnectionKeyValid(token) {
			conn.Write([]byte("Authentication FAIL\n"))
			continue
		}

		go logOutputService.SendLog(config.WebHookGithub, logData)
		conn.Write([]byte("RECEIVED\n"))

	}

	if err := scanner.Err(); err != nil {
		utils.Logger.ErrorF("Error reading from connection: %s", err.Error())
	}
}
