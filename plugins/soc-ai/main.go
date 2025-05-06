package main

import "github.com/utmstack/UTMStack/plugins/soc-ai/processor"

func main() {
	processor := processor.NewProcessor()
	processor.ProcessData()
}
