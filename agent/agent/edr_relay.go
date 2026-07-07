package agent

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

var (
	edrSpoolFile  = filepath.Join(fs.GetExecutablePath(), "edr-spool", "events.ndjson")
	edrOffsetFile = filepath.Join(fs.GetExecutablePath(), "edr-spool", ".offset")
	edrDataType   = "utmstack_edr"
)

// drainSpool reads unread lines from spool (starting at the persisted byte
// offset), calls emit for each, and persists the advanced offset.
func drainSpool(spoolPath, offsetPath string, emit func(raw string)) (int64, error) {
	start := readOffset(offsetPath)

	f, err := os.Open(spoolPath)
	if err != nil {
		if os.IsNotExist(err) {
			return start, nil
		}
		return start, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return start, err
	}
	if fi.Size() < start {
		// spool rotated/truncated; restart from the beginning
		start = 0
	}
	if _, err := f.Seek(start, 0); err != nil {
		return start, err
	}

	r := bufio.NewReader(f)
	pos := start
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 && strings.HasSuffix(line, "\n") {
			pos += int64(len(line))
			trimmed := strings.TrimRight(line, "\n")
			if trimmed != "" {
				emit(trimmed)
			}
		}
		if err != nil {
			break // EOF or read error: stop; partial (no newline) line stays unread
		}
	}
	writeOffset(offsetPath, pos)
	return pos, nil
}

func readOffset(p string) int64 {
	b, err := os.ReadFile(p)
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	return n
}

func writeOffset(p string, n int64) {
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(strconv.FormatInt(n, 10)), 0o644)
}

// EDRRelay drains the EDR spool into the agent LogQueue every second until ctx is done.
func EDRRelay(ctx context.Context) {
	host, _ := os.Hostname()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := drainSpool(edrSpoolFile, edrOffsetFile, func(raw string) {
				select {
				case LogQueue <- &plugins.Log{DataType: edrDataType, Raw: raw, DataSource: host}:
				default:
					utils.Logger.LogF(100, "EDR relay: LogQueue full, retrying next tick")
				}
			})
			if err != nil {
				utils.Logger.LogF(100, "EDR relay drain error: %v", err)
			}
		}
	}
}
