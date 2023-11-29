package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/log-auth-proxy/config"
	"github.com/utmstack/UTMStack/log-auth-proxy/logservice"
	"github.com/utmstack/UTMStack/log-auth-proxy/model"
)

func HttpLog(logOutputService *logservice.LogOutputService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logType, source, err := getHeaderAndSource(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		body["dataSource"] = source

		if logType != model.JsonInput {
			body["dataType"] = logType
		}

		jsonBytes, err := json.Marshal(body)
		if err != nil {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		logType, source, err := getHeaderAndSource(c)
		if err != nil {
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
			if logType != model.JsonInput {
				dataMap["dataType"] = logType
			}

			str, err := json.Marshal(v)
			if err != nil {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		jsonBytes, err := json.Marshal(body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to convert JSON to string"})
			return
		}

		jsonString := string(jsonBytes)

		go logOutputService.SendLog(model.WebHookGithub, jsonString)

		c.JSON(http.StatusOK, gin.H{"status": "received", "sendTo": model.WebHookGithub})
	}
}

func getHeaderAndSource(c *gin.Context) (model.LogType, string, error) {
	var logKind model.LogType
	logType := c.GetHeader(config.ProxyLogTypeHeader)
	source := c.GetHeader(config.ProxySourceHeader)
	if source == "" {
		return "", "", errors.New("source not provided")
	}

	if logType == "" {
		logKind = model.JsonInput
	} else {
		logKind = model.LogType(logType)
	}

	return logKind, source, nil
}
