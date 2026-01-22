package config

import (
	"os"
	"strings"

	"github.com/utmstack/UTMStack/plugins/feeds/utils"
)

func GetInternalKey() string {
	return utils.Getenv("INTERNAL_KEY")
}

func GetBackendUrl() string {
	return utils.Getenv("BACKEND_URL")
}

func GetThreadWindsURL() string {
	if isDevEnvironment() {
		return "https://apis.dev.threatwinds.com"
	}
	return "https://apis.threatwinds.com"
}

func GetOpenSearchHost() string {
	return utils.Getenv("OPENSEARCH_HOST")
}

func GetOpenSearchPort() string {
	return utils.Getenv("OPENSEARCH_PORT")
}

func GetDBHost() string {
	return utils.Getenv("DB_HOST")
}

func GetDBPort() string {
	return utils.Getenv("DB_PORT")
}

func GetDBUser() string {
	return utils.Getenv("DB_USER")
}

func GetDBPassword() string {
	return utils.Getenv("DB_PASS")
}

func GetDBName() string {
	return utils.Getenv("DB_NAME")
}

func isDevEnvironment() bool {
	env := os.Getenv("ENV")
	if env != "" {
		if strings.Contains(env, "-dev") ||
			strings.Contains(env, "-qa") ||
			strings.Contains(env, "-rc") {
			return true
		}
	}

	return false
}
