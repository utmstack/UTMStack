package processor

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/utmstack/soc-ai/configurations"
	"github.com/utmstack/soc-ai/elastic"
	"github.com/utmstack/soc-ai/schema"
	"github.com/utmstack/soc-ai/utils"
)

var (
	processor     *Processor
	processorOnce sync.Once
)

type Processor struct {
	AlertInfoQueue chan schema.AlertGPTDetails
	GPTQueue       chan schema.AlertGPTDetails
	ElasticQueue   chan schema.AlertGPTDetails
}

func NewProcessor() *Processor {
	processorOnce.Do(func() {
		processor = &Processor{
			AlertInfoQueue: make(chan schema.AlertGPTDetails, 1000),
			GPTQueue:       make(chan schema.AlertGPTDetails, 1000),
			ElasticQueue:   make(chan schema.AlertGPTDetails, 1000),
		}
	})
	return processor
}

func (p *Processor) ProcessData() {
	utils.Logger.Info("Starting SOC-AI Processor...")

	go configurations.UpdateGPTConfigurations()
	go p.restRouter()
	go p.processAlertsInfo()

	for i := 0; i < 10; i++ {
		go p.processGPTRequests()
	}

	go p.processAlertToElastic()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}

func (p *Processor) RegisterError(message, id string) {
	err := elastic.IndexStatus(id, "Error", "update")
	if err != nil {
		utils.Logger.ErrorF("error while indexing error in elastic: %v", err)
	}
	utils.Logger.ErrorF("%s", message)
}
