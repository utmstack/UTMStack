//go:build windows && arm64
// +build windows,arm64

package platform

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/collector/configwatcher"
	"github.com/utmstack/UTMStack/agent/collector/schema"
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

	id   uintptr
	ctx  context.Context
	stop chan struct{}
	once sync.Once

	mu       sync.Mutex
	retrying bool
}

func newEventSubscription(ctx context.Context, channel string, errors chan error) *EventSubscription {
	return &EventSubscription{
		Channel: channel,
		Query:   "*",
		Errors:  errors,
		ctx:     ctx,
		stop:    make(chan struct{}),
	}
}

const (
	EvtSubscribeToFutureEvents = 1
	evtSubscribeActionError    = 0
	evtSubscribeActionDeliver  = 1
	evtRenderEventXML          = 1
)

const (
	maxResubscribeAttempts = 10
	resubscribeBaseDelay   = 5 * time.Second
	resubscribeMaxDelay    = 5 * time.Minute

	// Coalesces the burst of events an editor emits on save.
	channelReloadDebounce = 1 * time.Second
)

var (
	modwevtapi       = windows.NewLazySystemDLL("wevtapi.dll")
	procEvtSubscribe = modwevtapi.NewProc("EvtSubscribe")
	procEvtRender    = modwevtapi.NewProc("EvtRender")
	procEvtClose     = modwevtapi.NewProc("EvtClose")
	incomingEvents   = make(chan string, 1024)
)

// syscall.NewCallback allocates from a pool of 2000 process-wide slots that are
// never released, and exhausting it aborts the process with a runtime throw that
// recover cannot catch. This single callback is shared by every subscription:
// EvtSubscribe carries the subscription id as its context argument, and
// dispatchEvent uses it to find the owner.
//
// Assigned in init because dispatchEvent reaches createLocked, which reads
// sharedCallback, and Go rejects that as a variable initialization cycle.
var sharedCallback uintptr

func init() {
	sharedCallback = syscall.NewCallback(dispatchEvent)
}

var (
	registryMu sync.Mutex
	registry   = map[uintptr]*EventSubscription{}
	lastID     uintptr
)

func dispatchEvent(action, userContext, event uintptr) uintptr {
	registryMu.Lock()
	evtSub := registry[userContext]
	registryMu.Unlock()

	if evtSub == nil {
		return 0
	}
	return evtSub.handleEvent(action, event)
}

// registerSubscription assigns an id on first use. Ids start at 1, because
// EvtSubscribe treats 0 as "no context".
func registerSubscription(evtSub *EventSubscription) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if evtSub.id == 0 {
		lastID++
		evtSub.id = lastID
	}
	registry[evtSub.id] = evtSub
}

func unregisterSubscription(evtSub *EventSubscription) {
	registryMu.Lock()
	defer registryMu.Unlock()
	delete(registry, evtSub.id)
}

func (evtSub *EventSubscription) Create() error {
	evtSub.mu.Lock()
	defer evtSub.mu.Unlock()
	return evtSub.createLocked()
}

// createLocked requires evtSub.mu, so the resubscribe path can reuse it without
// relocking a non-reentrant mutex.
func (evtSub *EventSubscription) createLocked() error {
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

	utils.Logger.LogF(100, "Subscribing to channel: %s", evtSub.Channel)

	// Registered before subscribing, because the callback can fire immediately.
	registerSubscription(evtSub)

	handle, _, err := procEvtSubscribe.Call(
		0,
		0,
		uintptr(unsafe.Pointer(winChannel)),
		uintptr(unsafe.Pointer(winQuery)),
		0,
		evtSub.id,
		sharedCallback,
		uintptr(EvtSubscribeToFutureEvents),
	)

	if handle == 0 {
		unregisterSubscription(evtSub)
		return fmt.Errorf("windows_events: failed to subscribe to events: %v", err)
	}

	evtSub.winAPIHandle = windows.Handle(handle)
	return nil
}

// Close releases the subscription and aborts a pending resubscribe. Safe to call
// more than once, since teardown may follow a reconcile that already closed it.
func (evtSub *EventSubscription) Close() error {
	evtSub.once.Do(func() { close(evtSub.stop) })

	evtSub.mu.Lock()
	defer evtSub.mu.Unlock()
	return evtSub.closeLocked()
}

func (evtSub *EventSubscription) closeLocked() error {
	unregisterSubscription(evtSub)

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

func (evtSub *EventSubscription) handleEvent(action, event uintptr) uintptr {
	switch action {
	case evtSubscribeActionError:
		err := fmt.Errorf("windows_events: error in callback, code: %x", uint16(event))
		evtSub.Errors <- err
		evtSub.scheduleResubscribe(err)

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

// scheduleResubscribe retries a failed subscription with backoff. Only one
// attempt loop runs per subscription, so a burst of error callbacks on the same
// channel cannot spawn competing goroutines.
func (evtSub *EventSubscription) scheduleResubscribe(cause error) {
	evtSub.mu.Lock()
	if evtSub.retrying {
		evtSub.mu.Unlock()
		return
	}
	evtSub.retrying = true
	_ = evtSub.closeLocked()
	evtSub.mu.Unlock()

	go func(channel string) {
		defer func() {
			evtSub.mu.Lock()
			evtSub.retrying = false
			evtSub.mu.Unlock()
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic resubscribing to channel %s: %v", channel, r)
			}
		}()

		utils.Logger.LogF(100, "Attempting to resubscribe to channel: %s after error: %v", channel, cause)

		delay := resubscribeBaseDelay
		for attempt := 1; attempt <= maxResubscribeAttempts; attempt++ {
			// Waiting outside the lock keeps teardown from blocking on a backoff.
			select {
			case <-evtSub.ctx.Done():
				return
			case <-evtSub.stop:
				return
			case <-time.After(delay):
			}

			evtSub.mu.Lock()
			err := evtSub.createLocked()
			evtSub.mu.Unlock()

			if err == nil {
				utils.Logger.LogF(100, "Resubscribed to channel: %s", channel)
				return
			}
			utils.Logger.ErrorF("Retry %d/%d failed for channel %s: %s",
				attempt, maxResubscribeAttempts, channel, err)

			delay = utils.IncrementReconnectDelay(delay, resubscribeMaxDelay)
		}

		utils.Logger.ErrorF("windows_events: channel %s unavailable after %d attempts, giving up",
			channel, maxResubscribeAttempts)
	}(evtSub.Channel)
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

func (w *Windows) Start(ctx context.Context, queue chan *plugins.Log) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.ErrorF("panic in Windows ARM64 collector: %v", r)
		}
	}()

	// Cancelled by the caller or by Stop(), so the watcher and any pending
	// resubscribe observe both paths.
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-runCtx.Done():
		case <-w.stopChan:
			cancel()
		}
	}()

	// Started once: reconcile runs again on every config change.
	errorsChan := make(chan error, 10)
	go eventWorker(queue)

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

	configwatcher.WatchFile(runCtx, "windows collector", config.WindowsEventChannelsFile,
		channelReloadDebounce, func() { w.reconcile(runCtx, errorsChan) })

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

// reconcile aligns the live subscriptions with the configured channels.
func (w *Windows) reconcile(ctx context.Context, errorsChan chan error) {
	channels, skipped, err := schema.ResolveWindowsEventChannels(
		config.WindowsEventChannelsFile, config.DefaultWindowsEventChannels)
	if err != nil {
		// Keep what is running. Editors save by truncating, so a read can catch
		// the file mid-write, and treating that as an empty list would drop every
		// custom subscription and rebuild it a moment later.
		utils.Logger.ErrorF("Error reading custom event channels, keeping current subscriptions: %s", err)
		return
	}
	for _, reason := range skipped {
		utils.Logger.Info("Custom event channel skipped: %s", reason)
	}

	desired := make(map[string]struct{}, len(channels))
	for _, channel := range channels {
		desired[strings.ToLower(channel)] = struct{}{}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	active := make(map[string]struct{}, len(w.subscriptions))
	kept := w.subscriptions[:0:0]
	for _, sub := range w.subscriptions {
		key := strings.ToLower(sub.Channel)
		if _, wanted := desired[key]; wanted {
			active[key] = struct{}{}
			kept = append(kept, sub)
			continue
		}
		// Defaults are always in desired, so only client-added channels get here.
		utils.Logger.Info("Unsubscribing from removed channel: %s", sub.Channel)
		if err := sub.Close(); err != nil {
			utils.Logger.ErrorF("Error closing subscription for %s: %v", sub.Channel, err)
		}
	}
	w.subscriptions = kept

	for _, channel := range channels {
		if _, running := active[strings.ToLower(channel)]; running {
			continue
		}
		sub := newEventSubscription(ctx, channel, errorsChan)
		if err := sub.Create(); err != nil {
			utils.Logger.ErrorF("Error subscribing to channel %s: %s", channel, err)
			continue
		}
		w.subscriptions = append(w.subscriptions, sub)
		utils.Logger.LogF(100, "Subscribed to channel: %s", channel)
	}
}

func eventWorker(queue chan *plugins.Log) {
	host, err := os.Hostname()
	if err != nil {
		utils.Logger.ErrorF("error getting hostname: %v", err)
		host = "unknown"
	}

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

		validatedLog, _, err := entities.ValidateString(eventJSON, false)
		if err != nil {
			utils.Logger.LogF(100, "validation error: %s: %v", eventJSON, err)
			continue
		}

		select {
		case queue <- &plugins.Log{
			DataSource: host,
			DataType:   string(config.DataTypeWindowsAgent),
			Raw:        validatedLog,
		}:
		default:
			utils.Logger.LogF(100, "LogQueue full: event discarded")
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
