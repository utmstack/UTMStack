package processor

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/soc-ai/configurations"
	"github.com/utmstack/UTMStack/plugins/soc-ai/elastic"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

var (
	processor     *Processor
	processorOnce sync.Once
)

type Processor struct {
	AlertInfoQueue chan schema.Alert
	GPTQueue       chan schema.AlertGPTDetails
	ElasticQueue   chan schema.AlertGPTDetails
}

func NewProcessor() *Processor {
	processorOnce.Do(func() {
		processor = &Processor{
			AlertInfoQueue: make(chan schema.Alert, 1000),
			GPTQueue:       make(chan schema.AlertGPTDetails, 1000),
			ElasticQueue:   make(chan schema.AlertGPTDetails, 1000),
		}
	})
	return processor
}

func (p *Processor) ProcessData() {

	go configurations.GetPluginConfig().UpdateGPTConfigurations()
	go p.socAiServer()
	go p.processAlertsInfo()

	for i := 0; i < 10; i++ {
		go p.processGPTRequests()
	}

	go p.processAlertToElastic()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}

func (p *Processor) RegisterError(message string, id string) {
	err := elastic.IndexStatus(id, "Error", "update")
	if err != nil {
		_ = catcher.Error("error while indexing error in elastic: %v", err, nil)
	}
	_ = catcher.Error("%s", errors.New(message), nil)
}
