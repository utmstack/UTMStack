package elastic

import (
	"errors"

	"github.com/threatwinds/go-sdk/catcher"
)

func RegisterError(message string, id string) {
	err := IndexStatus(id, "Error", "update")
	if err != nil {
		_ = catcher.Error("error while indexing error in elastic: %v", err, map[string]any{"process": "plugin_com.utmstack.soc-ai"})
	}
	_ = catcher.Error("TODO: Explain this error", errors.New(message), map[string]any{"process": "plugin_com.utmstack.soc-ai"})
}
