package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"math"
	"net/netip"
	"os"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/go-sdk/v2/plugins"
)

const (
	coordScale   = 1e6
	csvBufBytes  = 256 * 1024
	geoProcessID = "plugin_com.utmstack.geolocation"
)

// loadGeolocationData streams the 5 GeoLite2 CSV files from <WorkDir>/geolocation
// into compact, pointer-free sorted tables and publishes an immutable snapshot.
func loadGeolocationData() {
	geoDir, err := utils.MkdirJoin(plugins.WorkDir, "geolocation")
	if err != nil {
		_ = catcher.Error("could not create geolocation directory", err, map[string]any{"process": geoProcessID})
		return
	}

	snap := loadGeolocationFromDir(geoDir.String())
	if snap != nil {
		geoData.Store(snap)
		// Reclaim temporary load-time maps and CSV read buffers immediately.
		runtime.GC()
		debug.FreeOSMemory()
	}
}

func estimateCSVRows(filePath string, avgBytesPerRow int64) int {
	info, err := os.Stat(filePath)
	if err != nil || info.Size() <= 0 || avgBytesPerRow <= 0 {
		return 0
	}
	est := int(info.Size()/avgBytesPerRow) + 16
	if est < 0 {
		return 0
	}
	return est
}

func shrinkSlice[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	if cap(s)-len(s) > len(s)/16+64 {
		exact := make([]T, len(s))
		copy(exact, s)
		return exact
	}
	return slices.Clip(s)
}

func loadGeolocationFromDir(dir string) *geoSnapshot {
	sep := string(os.PathSeparator)
	locPath := dir + sep + "locations-en.csv"
	asnV4Path := dir + sep + "asn-blocks-v4.csv"
	asnV6Path := dir + sep + "asn-blocks-v6.csv"
	cityV4Path := dir + sep + "blocks-v4.csv"
	cityV6Path := dir + sep + "blocks-v6.csv"

	// Load locations-en.csv first so city blocks can resolve geoname_id -> compact uint32 index directly.
	locations, geoIDToIdx := loadCityLocationsWithRetry(locPath)
	if len(locations) <= 1 {
		return nil
	}

	// Intern repeated ASO names into a compact string slice (index 0 = "-").
	asos := []string{"-"}
	asoToIdx := map[string]uint32{"-": 0}

	asnV4 := make([]asnBlockV4, 0, estimateCSVRows(asnV4Path, 34))
	asnV6 := make([]asnBlockV6, 0, estimateCSVRows(asnV6Path, 48))
	loadASNBlocksWithRetry(asnV4Path, &asnV4, &asnV6, &asos, asoToIdx)
	loadASNBlocksWithRetry(asnV6Path, &asnV4, &asnV6, &asos, asoToIdx)

	cityV4 := make([]cityBlockV4, 0, estimateCSVRows(cityV4Path, 52))
	cityV6 := make([]cityBlockV6, 0, estimateCSVRows(cityV6Path, 68))
	loadCityBlocksWithRetry(cityV4Path, &cityV4, &cityV6, geoIDToIdx)
	loadCityBlocksWithRetry(cityV6Path, &cityV4, &cityV6, geoIDToIdx)

	// Only publish if ASN, city blocks, and locations were populated (matches original contract).
	if (len(asnV4) == 0 && len(asnV6) == 0) || (len(cityV4) == 0 && len(cityV6) == 0) {
		return nil
	}

	// Shrink over-allocated slices to exact length before publishing.
	asnV4 = shrinkSlice(asnV4)
	asnV6 = shrinkSlice(asnV6)
	cityV4 = shrinkSlice(cityV4)
	cityV6 = shrinkSlice(cityV6)
	locations = shrinkSlice(locations)
	asos = shrinkSlice(asos)

	return &geoSnapshot{
		asnV4:         asnV4,
		asnV4Overlap:  finalizeASNV4(asnV4),
		asnV6:         asnV6,
		asnV6Overlap:  finalizeASNV6(asnV6),
		cityV4:        cityV4,
		cityV4Overlap: finalizeCityV4(cityV4),
		cityV6:        cityV6,
		cityV6Overlap: finalizeCityV6(cityV6),
		locations:     locations,
		asos:          asos,
	}
}

func streamCSVWithRetry(filePath string, handleRow func(line []string)) {
	const maxRetries = 3
	var lastErr error

	for retry := 0; retry < maxRetries; retry++ {
		lastErr = streamCSVFile(filePath, handleRow)
		if lastErr == nil {
			return
		}

		_ = catcher.Error("could not read geolocation file, retrying", lastErr, map[string]any{
			"process":    geoProcessID,
			"file":       filePath,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(time.Duration(1<<uint(retry)) * time.Second)
		}
	}

	_ = catcher.Error("all retries failed when reading geolocation file", lastErr, map[string]any{
		"process": geoProcessID,
		"file":    filePath,
	})
}

func streamCSVFile(filePath string, handleRow func(line []string)) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	r := csv.NewReader(bufio.NewReaderSize(f, csvBufBytes))
	r.ReuseRecord = true
	r.FieldsPerRecord = -1

	// Skip CSV header row
	if _, err := r.Read(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	for {
		line, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		handleRow(line)
	}
}

func loadCityLocationsWithRetry(filePath string) ([]cityLocation, map[uint64]uint32) {
	est := estimateCSVRows(filePath, 80) + 1
	// Index 0 is reserved as the "no location" sentinel.
	locations := make([]cityLocation, 1, est)
	geoIDToIdx := make(map[uint64]uint32, est)
	countryCodePool := make(map[string]string, 300)
	countryNamePool := make(map[string]string, 300)

	intern := func(pool map[string]string, s string) string {
		if s == "" {
			return ""
		}
		if v, ok := pool[s]; ok {
			return v
		}
		cp := string([]byte(s))
		pool[cp] = cp
		return cp
	}

	streamCSVWithRetry(filePath, func(line []string) {
		if len(line) < 11 {
			return
		}
		geonameID, err := strconv.ParseUint(line[0], 10, 64)
		if err != nil {
			_ = catcher.Error("could not parse geonameID", err, map[string]any{
				"process":   geoProcessID,
				"geonameID": line[0],
			})
			return
		}

		// Copy cityName out of ReuseRecord buffer so backing bytes can be reused.
		cityName := ""
		if line[10] != "" {
			cityName = string([]byte(line[10]))
		}

		idx := uint32(len(locations))
		locations = append(locations, cityLocation{
			countryISOCode: intern(countryCodePool, line[4]),
			countryName:    intern(countryNamePool, line[5]),
			cityName:       cityName,
		})
		geoIDToIdx[geonameID] = idx
	})

	return locations, geoIDToIdx
}

func loadASNBlocksWithRetry(
	filePath string,
	outV4 *[]asnBlockV4,
	outV6 *[]asnBlockV6,
	asos *[]string,
	asoToIdx map[string]uint32,
) {
	streamCSVWithRetry(filePath, func(line []string) {
		if len(line) < 3 {
			return
		}
		prefix, err := netip.ParsePrefix(line[0])
		if err != nil {
			_ = catcher.Error("could not parse CIDR", err, map[string]any{
				"process": geoProcessID,
				"cidr":    line[0],
			})
			return
		}

		var asn uint64
		if line[1] != "" {
			asn, err = strconv.ParseUint(line[1], 10, 64)
			if err != nil {
				_ = catcher.Error("could not parse ASN", err, map[string]any{
					"process": geoProcessID,
					"asn":     line[1],
				})
				return
			}
		}

		var asoIdx uint32
		if line[2] != "" {
			if idx, ok := asoToIdx[line[2]]; ok {
				asoIdx = idx
			} else {
				cp := string([]byte(line[2]))
				asoIdx = uint32(len(*asos))
				*asos = append(*asos, cp)
				asoToIdx[cp] = asoIdx
			}
		}

		prefix = prefix.Masked()
		addr := prefix.Addr()
		if addr.Is4() {
			start, end := prefixRangeV4(prefix)
			*outV4 = append(*outV4, asnBlockV4{
				start:  start,
				end:    end,
				asn:    uint32(asn),
				asoIdx: asoIdx,
			})
		} else {
			startHi, startLo, endHi, endLo := prefixRangeV6(prefix)
			*outV6 = append(*outV6, asnBlockV6{
				startHi: startHi,
				startLo: startLo,
				endHi:   endHi,
				endLo:   endLo,
				asn:     uint32(asn),
				asoIdx:  asoIdx,
			})
		}
	})
}

func loadCityBlocksWithRetry(
	filePath string,
	outV4 *[]cityBlockV4,
	outV6 *[]cityBlockV6,
	geoIDToIdx map[uint64]uint32,
) {
	streamCSVWithRetry(filePath, func(line []string) {
		if len(line) < 10 {
			return
		}
		prefix, err := netip.ParsePrefix(line[0])
		if err != nil {
			_ = catcher.Error("could not parse CIDR", err, map[string]any{
				"process": geoProcessID,
				"cidr":    line[0],
			})
			return
		}

		var geonameID uint64
		if line[1] != "" {
			geonameID, err = strconv.ParseUint(line[1], 10, 64)
			if err != nil {
				_ = catcher.Error("could not parse geonameID", err, map[string]any{
					"process":   geoProcessID,
					"geonameID": line[1],
				})
				return
			}
		}

		var lat float64
		if line[7] != "" {
			lat, err = strconv.ParseFloat(line[7], 64)
			if err != nil {
				_ = catcher.Error("could not parse latitude", err, map[string]any{
					"process":  geoProcessID,
					"latitude": line[7],
				})
				return
			}
		}

		var lon float64
		if line[8] != "" {
			lon, err = strconv.ParseFloat(line[8], 64)
			if err != nil {
				_ = catcher.Error("could not parse longitude", err, map[string]any{
					"process":   geoProcessID,
					"longitude": line[8],
				})
				return
			}
		}

		if line[9] != "" {
			if _, err = strconv.Atoi(line[9]); err != nil {
				_ = catcher.Error("could not parse accuracyRadius", err, map[string]any{
					"process":        geoProcessID,
					"accuracyRadius": line[9],
				})
				return
			}
		}

		locIdx := geoIDToIdx[geonameID]
		latE6 := int32(math.Round(lat * coordScale))
		lonE6 := int32(math.Round(lon * coordScale))

		prefix = prefix.Masked()
		addr := prefix.Addr()
		if addr.Is4() {
			start, end := prefixRangeV4(prefix)
			*outV4 = append(*outV4, cityBlockV4{
				start:  start,
				end:    end,
				locIdx: locIdx,
				latE6:  latE6,
				lonE6:  lonE6,
			})
		} else {
			startHi, startLo, endHi, endLo := prefixRangeV6(prefix)
			*outV6 = append(*outV6, cityBlockV6{
				startHi: startHi,
				startLo: startLo,
				endHi:   endHi,
				endLo:   endLo,
				locIdx:  locIdx,
				latE6:   latE6,
				lonE6:   lonE6,
			})
		}
	})
}

func prefixRangeV4(p netip.Prefix) (uint32, uint32) {
	b := p.Addr().As4()
	start := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	bits := p.Bits()
	if bits == 0 {
		return 0, ^uint32(0)
	}
	hostMask := uint32(1)<<(32-bits) - 1
	return start &^ hostMask, start | hostMask
}

func prefixRangeV6(p netip.Prefix) (startHi, startLo, endHi, endLo uint64) {
	b := p.Addr().As16()
	hi := uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
	lo := uint64(b[8])<<56 | uint64(b[9])<<48 | uint64(b[10])<<40 | uint64(b[11])<<32 |
		uint64(b[12])<<24 | uint64(b[13])<<16 | uint64(b[14])<<8 | uint64(b[15])

	bits := p.Bits()
	switch {
	case bits == 0:
		return 0, 0, ^uint64(0), ^uint64(0)
	case bits < 64:
		hostHi := uint64(1)<<(64-bits) - 1
		return hi &^ hostHi, 0, hi | hostHi, ^uint64(0)
	case bits == 64:
		return hi, 0, hi, ^uint64(0)
	case bits < 128:
		hostLo := uint64(1)<<(128-bits) - 1
		return hi, lo &^ hostLo, hi, lo | hostLo
	default:
		return hi, lo, hi, lo
	}
}

func finalizeASNV4(s []asnBlockV4) bool {
	if !slices.IsSortedFunc(s, cmpASNV4) {
		slices.SortFunc(s, cmpASNV4)
	}
	for i := 1; i < len(s); i++ {
		if s[i-1].end >= s[i].start {
			return true
		}
	}
	return false
}

func cmpASNV4(a, b asnBlockV4) int {
	if a.start != b.start {
		if a.start < b.start {
			return -1
		}
		return 1
	}
	// Broader prefix (larger end) first so narrower nested CIDRs appear later.
	if a.end > b.end {
		return -1
	}
	if a.end < b.end {
		return 1
	}
	return 0
}

func finalizeCityV4(s []cityBlockV4) bool {
	if !slices.IsSortedFunc(s, cmpCityV4) {
		slices.SortFunc(s, cmpCityV4)
	}
	for i := 1; i < len(s); i++ {
		if s[i-1].end >= s[i].start {
			return true
		}
	}
	return false
}

func cmpCityV4(a, b cityBlockV4) int {
	if a.start != b.start {
		if a.start < b.start {
			return -1
		}
		return 1
	}
	if a.end > b.end {
		return -1
	}
	if a.end < b.end {
		return 1
	}
	return 0
}

func finalizeASNV6(s []asnBlockV6) bool {
	if !slices.IsSortedFunc(s, cmpASNV6) {
		slices.SortFunc(s, cmpASNV6)
	}
	for i := 1; i < len(s); i++ {
		if cmpU128(s[i-1].endHi, s[i-1].endLo, s[i].startHi, s[i].startLo) >= 0 {
			return true
		}
	}
	return false
}

func cmpASNV6(a, b asnBlockV6) int {
	if c := cmpU128(a.startHi, a.startLo, b.startHi, b.startLo); c != 0 {
		return c
	}
	return cmpU128(b.endHi, b.endLo, a.endHi, a.endLo)
}

func finalizeCityV6(s []cityBlockV6) bool {
	if !slices.IsSortedFunc(s, cmpCityV6) {
		slices.SortFunc(s, cmpCityV6)
	}
	for i := 1; i < len(s); i++ {
		if cmpU128(s[i-1].endHi, s[i-1].endLo, s[i].startHi, s[i].startLo) >= 0 {
			return true
		}
	}
	return false
}

func cmpCityV6(a, b cityBlockV6) int {
	if c := cmpU128(a.startHi, a.startLo, b.startHi, b.startLo); c != 0 {
		return c
	}
	return cmpU128(b.endHi, b.endLo, a.endHi, a.endLo)
}
