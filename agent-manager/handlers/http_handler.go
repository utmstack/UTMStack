package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/UTMStack/agent-manager/updates"
)

func HandleInstallerRequest(c *gin.Context) {
	installerType := c.Param("installerType")
	os := c.Param("os")

	if (installerType != "agent" && installerType != "collector") || (os != "windows" && os != "linux") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid installer type or os"})
		return
	}

	filepath := getFile(installerType, os)
	if filepath == "file not found" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file not found"})
		return
	}

	c.File(filepath)
}

func getFile(installerType, os string) string {
	switch installerType {
	case "agent":
		updater := updates.NewAgentUpdater()
		file, path := updater.GetFileName(fixOS(os), "DEPEND_TYPE_INSTALLER")
		return filepath.Join(path, file)
	case "collector":
		updater := updates.NewCollectorUpdater()
		file, path := updater.GetFileName(fixOS(os), "DEPEND_TYPE_INSTALLER")
		return filepath.Join(path, file)
	}
	return "file not found"
}

func fixOS(os string) string {
	if os == "windows" {
		return "OS_WINDOWS"
	}
	return "OS_LINUX"
}
