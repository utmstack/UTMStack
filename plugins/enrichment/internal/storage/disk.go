package storage

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"

	"github.com/utmstack/UTMStack/plugins/enrichment/config"
	"github.com/utmstack/UTMStack/plugins/enrichment/internal/csvio"
	"github.com/utmstack/UTMStack/plugins/enrichment/internal/registry"
)

func DatasetsDir() (utils.Folder, error) {
	return utils.MkdirJoin(plugins.WorkDir, "pipeline", config.CSVSubdir)
}

func SyncFromDisk() error {
	csvDir, err := DatasetsDir()
	if err != nil {
		return fmt.Errorf("could not create csv-datasets dir: %w", err)
	}

	entries, err := os.ReadDir(csvDir.String())
	if err != nil {
		if os.IsNotExist(err) {
			removeVanishedDatasets(map[string]struct{}{})
			return nil
		}
		return err
	}

	onDisk := make(map[string]struct{}, len(entries))
	loaded := 0
	skipped := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".csv") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".csv")
		filePath := csvDir.FileJoin(entry.Name())
		onDisk[id] = struct{}{}

		if !loadOrRefreshDataset(id, filePath) {
			skipped++
			continue
		}
		loaded++
	}

	removed := removeVanishedDatasets(onDisk)

	catcher.Info("sync from disk complete", map[string]any{
		"process": config.ProcessName,
		"loaded":  loaded,
		"skipped": skipped,
		"removed": removed,
	})
	return nil
}

func loadOrRefreshDataset(id, filePath string) bool {
	lineCount, err := countLines(filePath)
	if err != nil {
		catcher.Warn("could not count lines on csv — skipping", map[string]any{
			"process": config.ProcessName,
			"path":    filePath,
			"id":      id,
			"error":   err.Error(),
		})
		return false
	}

	if lineCount > config.MaxCSVRows+1 {
		_ = catcher.Error("csv exceeds max rows limit — skipping", nil, map[string]any{
			"process":  config.ProcessName,
			"path":     filePath,
			"id":       id,
			"lines":    lineCount,
			"maxRows":  config.MaxCSVRows,
		})
		return false
	}

	if lineCount > config.WarnCSVRows+1 {
		catcher.Warn("large csv dataset", map[string]any{
			"process": config.ProcessName,
			"id":      id,
			"lines":   lineCount,
		})
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		catcher.Warn("could not read csv — skipping", map[string]any{
			"process": config.ProcessName,
			"path":    filePath,
			"id":      id,
			"error":   err.Error(),
		})
		return false
	}

	if !utf8.Valid(raw) {
		catcher.Warn("invalid UTF-8 in csv — skipping", map[string]any{
			"process": config.ProcessName,
			"path":    filePath,
			"id":      id,
		})
		return false
	}

	var firstLine string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) != "" {
			firstLine = line
			break
		}
	}
	sep := csvio.DetectSeparator(firstLine)

	info, _ := os.Stat(filePath)
	var sizeBytes int64
	if info != nil {
		sizeBytes = info.Size()
	}

	ds, err := csvio.LoadCSV(bytes.NewReader(raw), sep, sizeBytes)
	if err != nil {
		catcher.Warn("could not parse csv — skipping", map[string]any{
			"process": config.ProcessName,
			"path":    filePath,
			"id":      id,
			"error":   err.Error(),
		})
		return false
	}
	ds.ID = id

	registry.Swap(id, ds)
	return true
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	count := 0
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func removeVanishedDatasets(onDisk map[string]struct{}) int {
	removed := 0
	for _, id := range registry.IDs() {
		if _, exists := onDisk[id]; !exists {
			if registry.Delete(id) {
				catcher.Info("dataset removed (file no longer on disk)", map[string]any{
					"process": config.ProcessName,
					"id":      id,
				})
				removed++
			}
		}
	}
	return removed
}
