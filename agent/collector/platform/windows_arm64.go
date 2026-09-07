//go:build windows && arm64
// +build windows,arm64

package platform

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/config"
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

// SecurityData holds the SID of the security context the event was raised
// in. Windows only populates <Security UserID="..."> for events tied to a
// specific user's session — many System-category events leave it empty.
type SecurityData struct {
	UserID string `xml:"UserID,attr"`
}

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
	bookmark     windows.Handle
	bookmarks    *bookmarkStore
	// events is this channel's own delivery queue (see A5 in
	// GAPS_AND_IMPROVEMENTS.md) — a burst on one Windows Event Log channel
	// (e.g. verbose PowerShell script block logging) fills only its own
	// queue, instead of a single shared one that would also delay events
	// from every other channel (including Security).
	events chan windowsEventEnvelope

	mu      sync.Mutex
	running bool
}

const (
	EvtSubscribeToFutureEvents     = 1
	EvtSubscribeStartAfterBookmark = 3
	evtSubscribeActionError        = 0
	evtSubscribeActionDeliver      = 1
	evtRenderEventXML              = 1
	evtRenderBookmark              = 2

	// perChannelQueueSize is the buffer size of each channel's own events
	// queue (see EventSubscription.events).
	perChannelQueueSize = 256
)

var (
	modwevtapi            = windows.NewLazySystemDLL("wevtapi.dll")
	procEvtSubscribe      = modwevtapi.NewProc("EvtSubscribe")
	procEvtRender         = modwevtapi.NewProc("EvtRender")
	procEvtClose          = modwevtapi.NewProc("EvtClose")
	procEvtCreateBookmark = modwevtapi.NewProc("EvtCreateBookmark")
	procEvtUpdateBookmark = modwevtapi.NewProc("EvtUpdateBookmark")
)

type windowsEventEnvelope struct {
	channel  string
	xml      string
	bookmark string
}

func (evtSub *EventSubscription) Create(initialBookmarkXML string) error {
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

	flags := uintptr(EvtSubscribeToFutureEvents)
	var bookmarkArg uintptr
	if initialBookmarkXML != "" {
		if bookmarkXMLPtr, perr := windows.UTF16PtrFromString(initialBookmarkXML); perr == nil {
			if h, _, cerr := procEvtCreateBookmark.Call(uintptr(unsafe.Pointer(bookmarkXMLPtr))); h != 0 {
				evtSub.bookmark = windows.Handle(h)
				bookmarkArg = h
				flags = EvtSubscribeStartAfterBookmark
			} else {
				utils.Logger.ErrorF("windows_events: failed to create bookmark for channel %s, starting from now: %v", evtSub.Channel, cerr)
			}
		} else {
			utils.Logger.ErrorF("windows_events: invalid saved bookmark for channel %s, starting from now: %v", evtSub.Channel, perr)
		}
	}

	if evtSub.bookmark == 0 {
		h, _, cerr := procEvtCreateBookmark.Call(0)
		if h == 0 {
			return fmt.Errorf("windows_events: failed to create bookmark: %v", cerr)
		}
		evtSub.bookmark = windows.Handle(h)
	}

	callback := syscall.NewCallback(evtSub.winAPICallback)

	log.Printf("Debug - Subscribing to channel: %s", evtSub.Channel)

	handle, _, err := procEvtSubscribe.Call(
		0,
		0,
		uintptr(unsafe.Pointer(winChannel)),
		uintptr(unsafe.Pointer(winQuery)),
		bookmarkArg,
		0,
		callback,
		flags,
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
		return fmt.Errorf("windows_events: no active subscription to close")
	}
	ret, _, err := procEvtClose.Call(uintptr(evtSub.winAPIHandle))
	if ret == 0 {
		return fmt.Errorf("windows_events: error closing handle: %s", err)
	}
	evtSub.winAPIHandle = 0

	if evtSub.bookmark != 0 {
		procEvtClose.Call(uintptr(evtSub.bookmark))
		evtSub.bookmark = 0
	}

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
				initial := ""
				if evtSub.bookmarks != nil {
					initial = evtSub.bookmarks.get(channel)
				}
				if err := evtSub.Create(initial); err != nil {
					utils.Logger.ErrorF("Retry failed for channel %s: %s", channel, err)
				} else {
					utils.Logger.LogF(100, "Resubscribed to channel: %s", channel)
					break
				}
			}
		}(evtSub.Channel)

	case evtSubscribeActionDeliver:
		utils.Logger.LogF(100, "Received event from channel: %s", evtSub.Channel)
		xmlStr, err := renderHandle(event, evtRenderEventXML)
		if err != nil {
			evtSub.Errors <- fmt.Errorf("render in callback: %v", err)
			break
		}

		bookmarkXML := ""
		if evtSub.bookmark != 0 {
			if ok, _, uerr := procEvtUpdateBookmark.Call(uintptr(evtSub.bookmark), event); ok == 0 {
				utils.Logger.ErrorF("failed to update bookmark for channel %s: %v", evtSub.Channel, uerr)
			} else if rendered, rerr := renderHandle(uintptr(evtSub.bookmark), evtRenderBookmark); rerr == nil {
				bookmarkXML = rendered
			} else {
				utils.Logger.ErrorF("failed to render bookmark for channel %s: %v", evtSub.Channel, rerr)
			}
		}

		select {
		case evtSub.events <- windowsEventEnvelope{channel: evtSub.Channel, xml: xmlStr, bookmark: bookmarkXML}:
		default:
			utils.Logger.ErrorF("events queue full for channel %s: event discarded", evtSub.Channel)
		}
	default:
		evtSub.Errors <- fmt.Errorf("windows_events: unsupported action in callback: %x", uint16(action))
	}
	return 0
}

func renderHandle(h uintptr, flag uintptr) (string, error) {
	bufSize := uint32(4096)
	for {
		space := make([]uint16, bufSize/2)
		used := uint32(0)
		prop := uint32(0)

		ret, _, err := procEvtRender.Call(
			0, h, flag,
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

type Windows struct {
	stopChan      chan struct{}
	subscriptions []*EventSubscription
	mu            sync.Mutex
}

var windowsCollector = &Windows{
	stopChan: make(chan struct{}),
}

func GetCollectors() []Collector {
	return []Collector{windowsCollector}
}

func (w *Windows) Name() string {
	return "windows-arm64"
}

func (w *Windows) Start(ctx context.Context, enqueue func(*plugins.Log) error) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.ErrorF("panic in Windows ARM64 collector: %v", r)
		}
	}()

	bookmarks := loadBookmarkStore()
	go bookmarks.flushLoop(ctx)

	errorsChan := make(chan error, 10)

	channels := []string{
		"Security", "Application", "System", "Windows Powershell", "Microsoft-Windows-Powershell/Operational", "ForwardedEvents",
		"Microsoft-Windows-WinLogon/Operational", "Microsoft-Windows-Windows Firewall With Advanced Security/Firewall",
		"Microsoft-Windows-Windows Defender/Operational",
		// Added for A10 (see agent/GAPS_AND_IMPROVEMENTS.md) — high
		// forensic value channels missing from the original list.
		// TaskScheduler and PrintService are disabled by default on a
		// stock Windows install; dependency.configureWindowsAuditPolicy
		// enables them at startup (via wevtutil), or subscribing here
		// would just never see any events.
		"Microsoft-Windows-TaskScheduler/Operational",
		"Microsoft-Windows-TerminalServices-RDPClient/Operational",
		"Microsoft-Windows-TerminalServices-LocalSessionManager/Operational",
		"Microsoft-Windows-TerminalServices-RemoteConnectionManager/Operational",
		"Microsoft-Windows-WinRM/Operational",
		"Microsoft-Windows-NTLM/Operational",
		"Microsoft-Windows-CodeIntegrity/Operational",
		"Microsoft-Windows-WMI-Activity/Operational",
		"Microsoft-Windows-GroupPolicy/Operational",
		"Microsoft-Windows-SmbClient/Security",
		"Microsoft-Windows-PrintService/Operational",
		// Only exists if Sysmon is installed; subscribing to a channel
		// that doesn't exist on this host fails gracefully below (logged,
		// skipped) like any other missing channel.
		"Microsoft-Windows-Sysmon/Operational",
	}

	w.mu.Lock()
	w.subscriptions = nil
	for _, channel := range channels {
		sub := &EventSubscription{
			Channel:   channel,
			Query:     "*",
			Errors:    errorsChan,
			bookmarks: bookmarks,
			events:    make(chan windowsEventEnvelope, perChannelQueueSize),
		}
		if err := sub.Create(bookmarks.get(channel)); err != nil {
			utils.Logger.ErrorF("Error subscribing to channel %s: %s", channel, err)
			continue
		}
		w.subscriptions = append(w.subscriptions, sub)
		go eventWorker(sub.events, enqueue, bookmarks)
		utils.Logger.LogF(100, "Subscribed to channel: %s", channel)
	}
	w.mu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in error handler: %v", r)
			}
		}()
		for err := range errorsChan {
			utils.Logger.ErrorF("Subscription error: %s", err)
		}
	}()

	// Block until context is cancelled or Stop() is called
	select {
	case <-ctx.Done():
	case <-w.stopChan:
	}

	utils.Logger.Info("Windows ARM64 collector stopping...")
	w.mu.Lock()
	for _, sub := range w.subscriptions {
		if err := sub.Close(); err != nil {
			utils.Logger.ErrorF("Error closing subscription for %s: %v", sub.Channel, err)
		}
	}
	w.subscriptions = nil
	w.mu.Unlock()
	utils.Logger.Info("Windows ARM64 collector stopped.")
}

func eventWorker(events chan windowsEventEnvelope, enqueue func(*plugins.Log) error, bookmarks *bookmarkStore) {
	host, err := os.Hostname()
	if err != nil {
		utils.Logger.ErrorF("error getting hostname: %v", err)
		host = "unknown"
	}

	for envelope := range events {
		ev := new(Event)
		if err := xml.Unmarshal([]byte(envelope.xml), ev); err != nil {
			utils.Logger.ErrorF("unmarshal error: %v", err)
			continue
		}

		eventJSON, err := convertEventToJSON(ev)
		if err != nil {
			utils.Logger.ErrorF("toJSON error: %v", err)
			continue
		}

		validatedLog, _, err := entities.ValidateString(eventJSON, false)
		if err != nil {
			utils.Logger.LogF(100, "validation error: %s: %v", eventJSON, err)
			continue
		}

		if err := enqueue(&plugins.Log{
			DataSource: host,
			DataType:   string(config.DataTypeWindowsAgent),
			Raw:        validatedLog,
		}); err != nil {
			utils.Logger.ErrorF("failed to persist windows event from channel %s, not advancing bookmark: %v", envelope.channel, err)
			continue
		}

		if envelope.bookmark != "" {
			bookmarks.set(envelope.channel, envelope.bookmark)
		}
	}
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

	if event.System.Security.UserID != "" {
		eventMap["userId"] = event.System.Security.UserID
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

func (w *Windows) Install() error {
	return nil
}

func (w *Windows) Uninstall() error {
	return nil
}

func (w *Windows) Stop() {
	select {
	case w.stopChan <- struct{}{}:
	default:
		// Already stopped or not started
	}
}
