package updates

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

var (
	maxQuerySize = 20
)

func HandleAgentUpdates(c *gin.Context) {
	version := c.Query("version")
	os := c.Query("os")
	dependencyType := c.Query("type")
	partIndex := c.Query("partIndex")
	partSize := c.Query("partSize")

	if err := validateParams(version, os, dependencyType, partSize, partIndex); err != nil {
		c.JSON(http.StatusBadRequest, models.DependencyUpdateResponse{Message: err.Error()})
		return
	}

	updater := NewAgentUpdater()
	handleUpdate(c, updater, version, os, dependencyType, partIndex, partSize)
}

func HandleCollectorUpdates(c *gin.Context) {
	collectorType := c.Query("collectorType")
	version := c.Query("version")
	os := c.Query("os")
	dependencyType := c.Query("type")
	partIndex := c.Query("partIndex")
	partSize := c.Query("partSize")

	if err := validateParams(version, os, dependencyType, partSize, partIndex); err != nil {
		c.JSON(http.StatusBadRequest, models.DependencyUpdateResponse{Message: err.Error()})
		return
	}

	var updater UpdaterInterf
	if dependencyType == config.DependInstallerLabel {
		updater = NewCollectorUpdater()
	} else {
		switch collectorType {
		case "as400":
			updater = NewAs400Updater()
		default:
			c.JSON(http.StatusBadRequest, models.DependencyUpdateResponse{Message: "invalid collector type"})
			return
		}
	}

	handleUpdate(c, updater, version, os, dependencyType, partIndex, partSize)
}

func validateParams(version, os, dependencytype, partSize, partIndex string) error {
	if version == "" || os == "" || dependencytype == "" || partSize == "" || partIndex == "" {
		return errors.New("missing required parameters")
	} else {
		if !utils.ValueExistsInList(os, config.DependOsAllows) {
			return errors.New("invalid os parameter")
		}
		if !utils.ValueExistsInList(dependencytype, config.DependTypes) {
			return errors.New("invalid dependency type parameter")
		}
		partSizeint, err := strconv.Atoi(partSize)
		if err != nil {
			return errors.New("invalid partSizeQuery parameter. Must be an integer")
		}
		if partSizeint > maxQuerySize {
			return fmt.Errorf("partSize parameter cannot be greater than %d", maxQuerySize)
		}
		partIndexint, err := strconv.Atoi(partIndex)
		if err != nil {
			return errors.New("invalid partIndex parameter. Must be an integer")
		}
		if partIndexint < 1 {
			return errors.New("partIndex parameter must be greater than 0")
		}
	}

	return nil
}

func handleUpdate(c *gin.Context, updater UpdaterInterf, version, os, dependencyType, partIndex, partSize string) {
	latestVersion := updater.GetLatestVersion(dependencyType)
	if latestVersion == "" {
		c.JSON(http.StatusNotFound, models.DependencyUpdateResponse{Message: "dependency not found"})
		return
	}
	if latestVersion == version {
		c.JSON(http.StatusPartialContent, models.DependencyUpdateResponse{
			Message: "dependency already up to date",
			Version: latestVersion,
		})
		return
	}

	fileContent, nParts, size, err := updater.GetFileContent(os, dependencyType, partIndex, partSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.DependencyUpdateResponse{Message: fmt.Sprintf("error getting dependency file: %v", err)})
		return
	}

	c.JSON(http.StatusPartialContent, models.DependencyUpdateResponse{
		Message:     "dependency update available",
		Version:     latestVersion,
		TotalSize:   size,
		PartIndex:   partIndex,
		NParts:      nParts,
		FileContent: fileContent,
	})
}
