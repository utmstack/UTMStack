//go:build windows && arm64
// +build windows,arm64

package collectors

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	twsdk "github.com/threatwinds/go-sdk/entities"
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
	winAPIHandle windows.Handle

	mu      sync.Mutex
	running bool
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
	incomingEvents   = make(chan string, 1024)
)

func (evtSub *EventSubscription) Create() error {
	evtSub.mu.Lock()
	defer evtSub.mu.Unlock()

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
	evtSub.mu.Lock()
	defer evtSub.mu.Unlock()

	if evtSub.winAPIHandle == 0 {
		return nil
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
		err := fmt.Errorf("windows_events: error in callback, code: %x", uint16(event))
		evtSub.Errors <- err

		go func(channel string) {
			utils.Logger.LogF(100, "Attempting to resubscribe to channel: %s after error: %v", channel, err)
			evtSub.mu.Lock()
			defer evtSub.mu.Unlock()

			_ = evtSub.Close()

			for {
				time.Sleep(5 * time.Second)
				if err := evtSub.Create(); err != nil {
					utils.Logger.ErrorF("Retry failed for channel %s: %s", channel, err)
				} else {
					utils.Logger.LogF(100, "Resubscribed to channel: %s", channel)
					break
				}
			}
		}(evtSub.Channel)

	case evtSubscribeActionDeliver:
		utils.Logger.LogF(100, "Received event from channel: %s", evtSub.Channel)
		xmlStr, err := quickRenderXML(event)
		if err != nil {
			evtSub.Errors <- fmt.Errorf("render in callback: %v", err)
			break
		}
		select {
		case incomingEvents <- xmlStr:
		default:
			utils.Logger.ErrorF("incomingEvents lleno: evento descartado")
		}
	default:
		evtSub.Errors <- fmt.Errorf("windows_events: unsupported action in callback: %x", uint16(action))
	}
	return 0
}

func eventWorker() {
	for xmlStr := range incomingEvents {
		ev := new(Event)
		if err := xml.Unmarshal([]byte(xmlStr), ev); err != nil {
			utils.Logger.ErrorF("unmarshal error: %v", err)
			continue
		}

		eventJSON, err := convertEventToJSON(ev)
		if err != nil {
			utils.Logger.ErrorF("toJSON error: %v", err)
			continue
		}

		validatedLog, _, err := twsdk.ValidateString(eventJSON, false)
		if err != nil {
			utils.Logger.LogF(100, "validation error: %s: %v", eventJSON, err)
			continue
		}

		select {
		case logservice.LogQueue <- logservice.LogPipe{
			Src:  string(config.DataTypeWindowsAgent),
			Logs: []string{validatedLog},
		}:
		default:
			utils.Logger.LogF(100, "LogQueue full: event discarded")
		}
	}
}

func quickRenderXML(h uintptr) (string, error) {
	bufSize := uint32(4096)
	for {
		space := make([]uint16, bufSize/2)
		used := uint32(0)
		prop := uint32(0)

		ret, _, err := procEvtRender.Call(
			0, h, evtRenderEventXML,
			uintptr(bufSize),
			uintptr(unsafe.Pointer(&space[0])),
			uintptr(unsafe.Pointer(&used)),
			uintptr(unsafe.Pointer(&prop)),
		)
		if ret == 0 {
			if err == windows.ERROR_INSUFFICIENT_BUFFER {
				bufSize *= 2
				continue
			}
			return "", err
		}
		return cleanXML(windows.UTF16ToString(space)), nil
	}
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
	go eventWorker()

	channels := []string{
		"Security", "Application", "System", "Microsoft-Windows-Sysmon/Operational", "Windows Powershell",
		"Microsoft-Windows-Powershell/Operational", "ForwardedEvents", "Microsoft-Windows-WinLogon/Operational",
		"Microsoft-Windows-Windows Firewall With Advanced Security/Firewall", "Microsoft-Windows-Windows Defender/Operational",
	}
	var subscriptions []*EventSubscription

	for _, channel := range channels {
		sub := &EventSubscription{
			Channel: channel,
			Query:   "*",
			Errors:  errorsChan,
		}
		if err := sub.Create(); err != nil {
			utils.Logger.LogF(100, "Error subscribing to channel %s: %s", channel, err)
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
	close(incomingEvents)
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
		"timestamp":    event.System.TimeCreated.SystemTime,
		"providerName": event.System.Provider.ProviderName,
		"providerGuid": event.System.Provider.ProviderGUID,
		"eventCode":    event.System.EventID,
		"version":      event.System.Version,
		"level":        event.System.Level,
		"task":         event.System.Task,
		"opcode":       event.System.Opcode,
		"keywords":     event.System.Keywords,
		"timeCreated":  event.System.TimeCreated.SystemTime,
		"recordId":     event.System.EventRecordID,
		"correlation":  event.System.Correlation,
		"execution":    event.System.Execution,
		"channel":      event.System.Channel,
		"computer":     event.System.Computer,
		"data":         make(map[string]interface{}),
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
