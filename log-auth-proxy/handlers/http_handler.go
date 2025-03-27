package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/logservice"
	"github.com/utmstack/UTMStack/log-auth-proxy/utils"
)

func HttpLog(logOutputService *logservice.LogOutputService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}

		if err := c.ShouldBindJSON(&body); err != nil {
			utils.Logger.ErrorF("Error binding http JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logType, source, err := getHeaderAndSource(c)
		if err != nil {
			utils.Logger.ErrorF("Error getting header and source: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		body["dataSource"] = source

		if logType != config.JsonInput {
			body["dataType"] = logType
		}

		jsonBytes, err := json.Marshal(body)
		if err != nil {
			utils.Logger.ErrorF("Error marshalling http JSON: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to convert JSON to string"})
			return
		}

		jsonString := string(jsonBytes)

		go logOutputService.SendLog(logType, jsonString)

		c.JSON(http.StatusOK, gin.H{"status": "received", "sendTo": logType})
	}
}

func HttpBulkLog(logOutputService *logservice.LogOutputService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body []interface{}

		if err := c.ShouldBindJSON(&body); err != nil {
			utils.Logger.ErrorF("Error binding bulk JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		logType, source, err := getHeaderAndSource(c)
		if err != nil {
			utils.Logger.ErrorF("Error getting header and source: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		strs := make([]string, 0)
		for _, v := range body {
			dataMap, isMap := v.(map[string]interface{})
			if !isMap {
				continue
			}

			dataMap["dataSource"] = source
			if logType != config.JsonInput {
				dataMap["dataType"] = logType
			}

			str, err := json.Marshal(v)
			if err != nil {
				utils.Logger.ErrorF("Error marshalling bulk JSON: %v", err)
				continue
			}
			log := string(str)
			strs = append(strs, log)
		}
		go logOutputService.SendBulkLog(logType, strs)

		c.JSON(http.StatusOK, gin.H{"status": "received"})
	}
}

func HttpPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ping": "ok"})
}

func HttpGitHubHandler(logOutputService *logservice.LogOutputService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body interface{}

		if err := c.ShouldBindJSON(&body); err != nil {
			utils.Logger.ErrorF("Error binding github JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		jsonBytes, err := json.Marshal(body)
		if err != nil {
			utils.Logger.ErrorF("Error marshalling github JSON: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to convert JSON to string"})
			return
		}

		jsonString := string(jsonBytes)

		go logOutputService.SendLog(config.WebHookGithub, jsonString)

		c.JSON(http.StatusOK, gin.H{"status": "received", "sendTo": config.WebHookGithub})
	}
}

func getHeaderAndSource(c *gin.Context) (config.LogType, string, error) {
	var logKind config.LogType
	logType := c.GetHeader(config.ProxyLogTypeHeader)
	source := c.GetHeader(config.ProxySourceHeader)
	if source == "" {
		return "", "", errors.New("source not provided")
	}

	if logType == "" {
		logKind = config.JsonInput
	} else {
		logKind = config.LogType(logType)
	}

	return logKind, source, nil
}
