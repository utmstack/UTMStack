//go:build windows && arm64
// +build windows,arm64

package collectors

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"
	"github.com/utmstack/UTMStack/agent/utils"
	"golang.org/x/sys/windows"
)

type Event struct {
	XMLName   xml.Name     `xml:"Event"`
	System    SystemData   `xml:"System"`
	EventData []*EventData `xml:"EventData>Data"`
}

type EventData struct {
	Key   string `xml:"Name,attr"`
	Value string `xml:",chardata"`
}

type ProviderData struct {
	ProviderName string `xml:"Name,attr"`
	ProviderGUID string `xml:"Guid,attr"`
}

type TimeCreatedData struct {
	SystemTime string `xml:"SystemTime,attr"`
}

type CorrelationData struct {
	ActivityID string `xml:"ActivityID,attr"`
}

type ExecutionData struct {
	ProcessID int `xml:"ProcessID,attr"`
	ThreadID  int `xml:"ThreadID,attr"`
}

type SecurityData struct{}

type SystemData struct {
	Provider      ProviderData    `xml:"Provider"`
	EventID       int             `xml:"EventID"`
	Version       int             `xml:"Version"`
	Level         int             `xml:"Level"`
	Task          int             `xml:"Task"`
	Opcode        int             `xml:"Opcode"`
	Keywords      string          `xml:"Keywords"`
	TimeCreated   TimeCreatedData `xml:"TimeCreated"`
	EventRecordID int64           `xml:"EventRecordID"`
	Correlation   CorrelationData `xml:"Correlation"`
	Execution     ExecutionData   `xml:"Execution"`
	Channel       string          `xml:"Channel"`
	Computer      string          `xml:"Computer"`
	Security      SecurityData    `xml:"Security"`
}

type EventSubscription struct {
	Channel      string
	Query        string
	Errors       chan error
	Callback     func(event *Event)
	winAPIHandle windows.Handle
}

const (
	EvtSubscribeToFutureEvents = 1
	evtSubscribeActionError    = 0
	evtSubscribeActionDeliver  = 1
	evtRenderEventXML          = 1
)

var (
	modwevtapi       = windows.NewLazySystemDLL("wevtapi.dll")
	procEvtSubscribe = modwevtapi.NewProc("EvtSubscribe")
	procEvtRender    = modwevtapi.NewProc("EvtRender")
	procEvtClose     = modwevtapi.NewProc("EvtClose")
)

func (evtSub *EventSubscription) Create() error {
	if evtSub.winAPIHandle != 0 {
		return fmt.Errorf("windows_events: subscription has already been created")
	}

	winChannel, err := windows.UTF16PtrFromString(evtSub.Channel)
	if err != nil {
		return fmt.Errorf("windows_events: invalid channel name: %s", err)
	}

	winQuery, err := windows.UTF16PtrFromString(evtSub.Query)
	if err != nil {
		return fmt.Errorf("windows_events: invalid query: %s", err)
	}

	callback := syscall.NewCallback(evtSub.winAPICallback)

	log.Printf("Debug - Subscribing to channel: %s", evtSub.Channel)

	handle, _, err := procEvtSubscribe.Call(
		0,
		0,
		uintptr(unsafe.Pointer(winChannel)),
		uintptr(unsafe.Pointer(winQuery)),
		0,
		0,
		callback,
		uintptr(EvtSubscribeToFutureEvents),
	)

	if handle == 0 {
		return fmt.Errorf("windows_events: failed to subscribe to events: %v", err)
	}

	evtSub.winAPIHandle = windows.Handle(handle)
	return nil
}

func (evtSub *EventSubscription) Close() error {
	if evtSub.winAPIHandle == 0 {
		return fmt.Errorf("windows_events: no active subscription to close")
	}
	ret, _, err := procEvtClose.Call(uintptr(evtSub.winAPIHandle))
	if ret == 0 {
		return fmt.Errorf("windows_events: error closing handle: %s", err)
	}
	evtSub.winAPIHandle = 0
	return nil
}

func (evtSub *EventSubscription) winAPICallback(action, userContext, event uintptr) uintptr {
	switch action {
	case evtSubscribeActionError:
		evtSub.Errors <- fmt.Errorf("windows_events: error in callback, code: %x", uint16(event))
	case evtSubscribeActionDeliver:
		bufferSize := uint32(4096)
		for {
			renderSpace := make([]uint16, bufferSize/2)
			bufferUsed := uint32(0)
			propertyCount := uint32(0)
			ret, _, err := procEvtRender.Call(
				0,
				event,
				evtRenderEventXML,
				uintptr(bufferSize),
				uintptr(unsafe.Pointer(&renderSpace[0])),
				uintptr(unsafe.Pointer(&bufferUsed)),
				uintptr(unsafe.Pointer(&propertyCount)),
			)
			if ret == 0 {
				if err == windows.ERROR_INSUFFICIENT_BUFFER {
					bufferSize *= 2
					continue
				}
				evtSub.Errors <- fmt.Errorf("windows_events: failed to render event: %w", err)
				return 0
			}
			xmlStr := windows.UTF16ToString(renderSpace)
			xmlStr = cleanXML(xmlStr)

			dataParsed := new(Event)
			if err := xml.Unmarshal([]byte(xmlStr), dataParsed); err != nil {
				evtSub.Errors <- fmt.Errorf("windows_events: failed to parse XML: %s", err)
			} else {
				evtSub.Callback(dataParsed)
			}
			break
		}
	default:
		evtSub.Errors <- fmt.Errorf("windows_events: unsupported action in callback: %x", uint16(action))
	}
	return 0
}

func cleanXML(xmlStr string) string {
	xmlStr = strings.TrimSpace(xmlStr)
	if idx := strings.Index(xmlStr, "<?xml"); idx > 0 {
		xmlStr = xmlStr[idx:]
	}
	xmlStr = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, xmlStr)
	return xmlStr
}

type Windows struct{}

func getCollectorsInstances() []Collector {
	var collectors []Collector
	collectors = append(collectors, Windows{})
	return collectors
}

func (w Windows) SendLogs() {
	errorsChan := make(chan error, 10)

	callback := func(event *Event) {
		eventJSON, err := convertEventToJSON(event)
		if err != nil {
			utils.Logger.ErrorF("error converting event to JSON: %v", err)
			return
		}
		host, err := os.Hostname()
		if err != nil {
			utils.Logger.ErrorF("error getting hostname: %v", err)
			host = "unknown"
		}
		validatedLog, _, err := validations.ValidateString(eventJSON, false)
		if err != nil {
			utils.Logger.LogF(100, "error validating log: %s: %v", eventJSON, err)
			return
		}
		logservice.LogQueue <- &plugins.Log{
			DataType:   string(config.DataTypeWindowsAgent),
			DataSource: host,
			Raw:        validatedLog,
		}
	}

	channels := []string{"Security", "Application", "System"}
	var subscriptions []*EventSubscription

	for _, channel := range channels {
		sub := &EventSubscription{
			Channel:  channel,
			Query:    "*",
			Errors:   errorsChan,
			Callback: callback,
		}
		if err := sub.Create(); err != nil {
			utils.Logger.ErrorF("Error subscribing to channel %s: %s", channel, err)
			continue
		}
		subscriptions = append(subscriptions, sub)
		utils.Logger.LogF(100, "Subscribed to channel: %s", channel)
	}

	go func() {
		for err := range errorsChan {
			utils.Logger.ErrorF("Subscription error: %s", err)
		}
	}()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt)
	<-exitChan
	utils.Logger.LogF(100, "Interrupt received, closing subscriptions...")
	for _, sub := range subscriptions {
		if err := sub.Close(); err != nil {
			utils.Logger.ErrorF("Error closing subscription for %s: %v", sub.Channel, err)
		}
	}
	utils.Logger.LogF(100, "Agent finished successfully.")
}

func convertEventToJSON(event *Event) (string, error) {
	eventMap := map[string]interface{}{
		"timestamp":     event.System.TimeCreated.SystemTime,
		"provider_name": event.System.Provider.ProviderName,
		"provider_guid": event.System.Provider.ProviderGUID,
		"eventCode":     event.System.EventID,
		"version":       event.System.Version,
		"level":         event.System.Level,
		"task":          event.System.Task,
		"opcode":        event.System.Opcode,
		"keywords":      event.System.Keywords,
		"timeCreated":   event.System.TimeCreated.SystemTime,
		"recordId":      event.System.EventRecordID,
		"correlation":   event.System.Correlation,
		"execution":     event.System.Execution,
		"channel":       event.System.Channel,
		"computer":      event.System.Computer,
		"data":          make(map[string]interface{}),
	}

	dataMap := eventMap["data"].(map[string]interface{})
	for _, data := range event.EventData {
		if strings.HasPrefix(data.Value, "0x") {
			if val, err := strconv.ParseInt(data.Value[2:], 16, 64); err == nil {
				dataMap[data.Key] = val
				continue
			}
		}
		if data.Key != "" {
			value := strings.TrimSpace(data.Value)
			if value != "" {
				dataMap[data.Key] = value
			}
		}
	}

	jsonBytes, err := json.Marshal(eventMap)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (w Windows) Install() error {
	return nil
}

func (w Windows) Uninstall() error {
	return nil
}
