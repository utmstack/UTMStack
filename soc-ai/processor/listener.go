package processor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/soc-ai/configurations"
	"github.com/utmstack/soc-ai/elastic"
	"github.com/utmstack/soc-ai/schema"
)

func (p *Processor) restRouter() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	r.NoRoute(notFound)

	r.POST("/process", p.handleAlerts)

	r.Run(":" + configurations.SOC_AI_SERVER_PORT)
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusBadRequest, "url not found")
}

func (p *Processor) handleAlerts(c *gin.Context) {
	err := elastic.CreateIndexIfNotExist(configurations.SOC_AI_INDEX)
	if err != nil {
		catcher.Error("error creating index", err, map[string]any{"index": configurations.SOC_AI_INDEX})
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("error creating index %s", configurations.SOC_AI_INDEX))
		return
	}

	if !configurations.GetGPTConfig().ModuleActive {
		catcher.Error("Droping request to /process, GPT module is not active", nil, nil)
		c.JSON(http.StatusOK, "GPT module is not active")
		return
	}

	var ids []string
	if err := c.BindJSON(&ids); err != nil {
		catcher.Error("error binding JSON", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, id := range ids {
		result, err := elastic.ElasticSearch(configurations.SOC_AI_INDEX, "activityId", id)
		if err != nil {
			if !strings.Contains(err.Error(), "no such host") {
				catcher.Error("error while searching alert in elastic", err, map[string]any{"alert": id})
				c.JSON(http.StatusInternalServerError, fmt.Sprintf("error while searching alert %s in elastic: %v", id, err))
				return
			}
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("error while searching alert %s in elastic", id))
			return
		}

		var gptResponses []schema.GPTAlertResponse
		err = json.Unmarshal(result, &gptResponses)
		if err != nil {
			catcher.Error("error decoding response from elastic", err, nil)
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("error decoding response: %v", err))
			return
		}

		if len(gptResponses) == 0 {
			err = elastic.IndexStatus(id, "Processing", "create")
			if err != nil {
				catcher.Error("error creating doc in index", err, nil)
				c.JSON(http.StatusInternalServerError, fmt.Sprintf("error creating doc in index: %v", err))
				return
			}
		} else {
			err = elastic.IndexStatus(id, "Processing", "update")
			if err != nil {
				catcher.Error("error updating doc in index", err, nil)
				c.JSON(http.StatusInternalServerError, fmt.Sprintf("error updating doc in index: %v", err))
				return
			}
		}

		catcher.Info("alert received", map[string]any{"alert": id})
		p.AlertInfoQueue <- schema.AlertGPTDetails{AlertID: id}
	}

	c.JSON(http.StatusOK, "alerts received successfully")
}
