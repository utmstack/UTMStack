package client

import (
	"github.com/threatwinds/go-sdk/catcher"
	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
)

func ConnectOpenSearch() error {
	osUrl := plugins.PluginCfg("org.opensearch", false).Get("opensearch").String()

	err := sdkos.Connect([]string{osUrl}, "", "")
	if err != nil {
		return catcher.Error("failed to connect to OpenSearch", err, map[string]any{
			"url": osUrl,
		})
	}

	catcher.Info("Connected to OpenSearch", map[string]any{
		"url": osUrl,
	})

	return nil
}
