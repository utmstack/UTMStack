package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/updates"
)

func validateParams(version, os, dependencytype string) error {
	if version == "" || os == "" || dependencytype == "" {
		return errors.New("missing required parameters")
	} else {
		if os != "windows" && os != "linux" {
			return errors.New("invalid os parameter")
		}
		if dependencytype != "installer" && dependencytype != "service" && dependencytype != "depend_zip" {
			return errors.New("invalid dependency type parameter")
		}
	}

	return nil
}

func handleUpdate(c *gin.Context, updater updates.UpdaterInterf, version, os, dependencyType string) {
	latestVersion := updater.GetLatestVersion(dependencyType)
	if latestVersion == "" {
		c.JSON(http.StatusNotFound, models.DependencyUpdateResponse{Message: "dependency not found"})
		return
	}
	if latestVersion == version {
		c.JSON(http.StatusOK, models.DependencyUpdateResponse{
			Message: "dependency already up to date",
			Version: latestVersion,
		})
		return
	}

	fileContent, err := updater.GetFileContent(os, dependencyType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.DependencyUpdateResponse{Message: fmt.Sprintf("error getting dependency file: %v", err)})
		return
	}

	c.JSON(http.StatusOK, models.DependencyUpdateResponse{
		Message:     "dependency update available",
		Version:     latestVersion,
		FileContent: fileContent,
	})
}

func HandleAgentUpdates(c *gin.Context) {
	version := c.Query("version")
	os := c.Query("os")
	dependencyType := c.Query("type")

	if err := validateParams(version, os, dependencyType); err != nil {
		c.JSON(http.StatusBadRequest, models.DependencyUpdateResponse{Message: err.Error()})
		return
	}

	updater := updates.NewAgentUpdater()
	handleUpdate(c, updater, version, os, dependencyType)
}

func HandleCollectorUpdates(c *gin.Context) {
	collectorType := c.Query("collectorType")
	version := c.Query("version")
	os := c.Query("os")
	dependencytype := c.Query("type")

	if err := validateParams(version, os, dependencytype); err != nil {
		c.JSON(http.StatusBadRequest, models.DependencyUpdateResponse{Message: err.Error()})
		return
	}

	var updater updates.UpdaterInterf
	if dependencytype == "installer" {
		updater = updates.NewCollectorUpdater()
	} else {
		switch collectorType {
		case "as400":
			updater = updates.NewAs400Updater()
		default:
			c.JSON(http.StatusBadRequest, models.DependencyUpdateResponse{Message: "invalid collector type"})
			return
		}
	}

	handleUpdate(c, updater, version, os, dependencytype)
}
