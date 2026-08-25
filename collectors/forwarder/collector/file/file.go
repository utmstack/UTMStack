package file

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/configwatcher"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/schema"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

const pollInterval = 1 * time.Second

// fileWatcher represents an active file being tailed.
type fileWatcher struct {
	dataType string
	path     string
	cancel   context.CancelFunc
}

// FileCollector manages file-based log collection.
type FileCollector struct {
	watchers map[string]*fileWatcher // key: dataType+path
	mu       sync.RWMutex
	queue    chan *plugins.Log
}

// New creates a new FileCollector.
func New() *FileCollector {
	return &FileCollector{
		watchers: make(map[string]*fileWatcher),
	}
}

func (fc *FileCollector) Name() string {
	return "file"
}

func (fc *FileCollector) Stop() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	for key, w := range fc.watchers {
		if w.cancel != nil {
			w.cancel()
		}
		delete(fc.watchers, key)
	}
}

// Start begins watching for configuration changes using fsnotify.
// It performs an initial reconciliation and then reacts to config file changes.
func (fc *FileCollector) Start(ctx context.Context, queue chan *plugins.Log) {
	fc.queue = queue
	configwatcher.Watch(ctx, "file collector", func() {
		fc.reconcile(ctx)
	})
	fc.Stop()
}

func (fc *FileCollector) reconcile(ctx context.Context) {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		utils.Logger.ErrorF("file collector: error reading config: %v", err)
		return
	}

	// Build set of desired watchers from enabled file integrations
	desiredWatchers := make(map[string]struct{})

	for intType, integration := range cnf.FileIntegrations {
		// Only handle file-type integrations (present in FilePaths)
		if !isFileType(intType) {
			continue
		}

		if !integration.Enabled {
			// Stop watchers for disabled integrations
			fc.stopWatchersForDataType(intType)
			continue
		}

		// Expand glob patterns and start watchers
		for _, pathPattern := range integration.Paths {
			matches, err := filepath.Glob(pathPattern)
			if err != nil {
				utils.Logger.ErrorF("file collector: invalid glob pattern %s: %v", pathPattern, err)
				continue
			}

			if len(matches) == 0 {
				utils.Logger.LogF(100, "file collector: no files match pattern %s for %s", pathPattern, intType)
				continue
			}

			for _, filePath := range matches {
				key := intType + ":" + filePath
				desiredWatchers[key] = struct{}{}
				fc.ensureWatcher(ctx, intType, filePath)
			}
		}
	}

	// Stop watchers that are no longer needed
	fc.mu.Lock()
	for key, w := range fc.watchers {
		if _, exists := desiredWatchers[key]; !exists {
			utils.Logger.Info("file collector: stopping watcher for %s", key)
			if w.cancel != nil {
				w.cancel()
			}
			delete(fc.watchers, key)
		}
	}
	fc.mu.Unlock()
}

// isFileType returns true if the dataType is a file-based integration
// (i.e., it exists in FilePaths).
func isFileType(dataType string) bool {
	_, ok := config.FilePaths[config.DataType(dataType)]
	return ok
}

func (fc *FileCollector) stopWatchersForDataType(dataType string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	for key, w := range fc.watchers {
		if w.dataType == dataType {
			utils.Logger.Info("file collector: stopping watcher for %s", key)
			if w.cancel != nil {
				w.cancel()
			}
			delete(fc.watchers, key)
		}
	}
}

func (fc *FileCollector) ensureWatcher(ctx context.Context, dataType, filePath string) {
	key := dataType + ":" + filePath

	fc.mu.Lock()
	if _, exists := fc.watchers[key]; exists {
		fc.mu.Unlock()
		return
	}

	// Create a new watcher
	watchCtx, cancel := context.WithCancel(ctx)
	w := &fileWatcher{
		dataType: dataType,
		path:     filePath,
		cancel:   cancel,
	}
	fc.watchers[key] = w
	fc.mu.Unlock()

	// Start tailing in a goroutine
	go fc.tailFile(watchCtx, w)
	utils.Logger.Info("file collector: started watching %s for %s", filePath, dataType)
}

func (fc *FileCollector) tailFile(ctx context.Context, w *fileWatcher) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	var file *os.File
	var offset int64
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Open the file if not already open or if it was rotated
		if file == nil {
			opened, err := os.Open(w.path)
			if err != nil {
				utils.Logger.LogF(100, "file collector: error opening %s: %v", w.path, err)
				time.Sleep(pollInterval)
				continue
			}
			file = opened

			// Seek to end to only read new content
			end, err := file.Seek(0, io.SeekEnd)
			if err != nil {
				utils.Logger.ErrorF("file collector: error seeking %s: %v", w.path, err)
				file.Close()
				file = nil
				time.Sleep(pollInterval)
				continue
			}
			offset = end
		}

		// Check for file rotation (file was replaced)
		stat, err := os.Stat(w.path)
		if err != nil {
			// File may have been deleted, close and retry
			file.Close()
			file = nil
			time.Sleep(pollInterval)
			continue
		}

		fileStat, err := file.Stat()
		if err != nil || !os.SameFile(stat, fileStat) {
			// File was rotated, reopen
			file.Close()
			file = nil
			offset = 0
			continue
		}

		// Check if file was truncated
		if stat.Size() < offset {
			offset = 0
			file.Seek(0, io.SeekStart)
		}

		// Read new lines
		reader := bufio.NewReader(file)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					utils.Logger.ErrorF("file collector: error reading %s: %v", w.path, err)
				}
				break
			}

			if len(line) == 0 {
				continue
			}

			// Update offset
			offset += int64(len(line))

			// Validate and send log
			validatedLog, _, err := entities.ValidateString(line, false)
			if err != nil {
				utils.Logger.LogF(100, "file collector: validation error for %s: %v", w.path, err)
				continue
			}

			select {
			case fc.queue <- &plugins.Log{
				DataType:   w.dataType,
				DataSource: hostname,
				Raw:        validatedLog,
			}:
			default:
				utils.Logger.LogF(100, "file collector: queue full, discarding log from %s", w.path)
			}
		}

		time.Sleep(pollInterval)
	}
}
