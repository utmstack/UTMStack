//go:build linux || edr_nft_readback

package netblock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

// The Linux command boundary translates schema-1 wire values into the
// controller's representation. Upstream 1.0.9 and 1.1.6 json.c serialize set
// and element durations as integer seconds; the controller uses milliseconds.
// Under -n, nfproto uses AF_INET/AF_INET6. Full ordered verification follows.
func normalizeNFTReadback(data []byte) ([]byte, error) {
	if len(data) > nftMaxOutput {
		return nil, ErrNetworkBudget
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var doc map[string]any
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("invalid nftables wire readback: %w", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return nil, ErrNetworkUnverified
	}
	objects, ok := doc["nftables"].([]any)
	if !ok {
		return nil, ErrNetworkUnverified
	}
	metaCount := 0
	for _, value := range objects {
		object, ok := value.(map[string]any)
		if !ok {
			return nil, ErrNetworkUnverified
		}
		if meta, present := object["metainfo"]; present {
			m, ok := meta.(map[string]any)
			if !ok || m["json_schema_version"] != json.Number("1") {
				return nil, ErrNetworkUnverified
			}
			metaCount++
		}
	}
	if metaCount != 1 {
		return nil, ErrNetworkUnverified
	}
	changed := false
	for _, value := range objects {
		object := value.(map[string]any)
		for _, kind := range []string{"set", "element"} {
			if s, ok := object[kind].(map[string]any); ok {
				if kind == "set" {
					if err := nftWireDuration(s, "timeout", &changed); err != nil {
						return nil, err
					}
				}
				if elements, ok := s["elem"].([]any); ok {
					for _, value := range elements {
						if wrapper, ok := value.(map[string]any); ok {
							if elem, ok := wrapper["elem"].(map[string]any); ok {
								for _, field := range []string{"timeout", "expires"} {
									if err := nftWireDuration(elem, field, &changed); err != nil {
										return nil, err
									}
								}
							}
						}
					}
				}
			}
		}
		if rule, ok := object["rule"].(map[string]any); ok {
			if expressions, ok := rule["expr"].([]any); ok {
				for _, value := range expressions {
					expr, _ := value.(map[string]any)
					match, _ := expr["match"].(map[string]any)
					left, _ := match["left"].(map[string]any)
					meta, _ := left["meta"].(map[string]any)
					if meta["key"] != "nfproto" {
						continue
					}
					switch match["right"] {
					case json.Number("2"):
						match["right"], changed = "ipv4", true
					case json.Number("10"):
						match["right"], changed = "ipv6", true
					}
				}
			}
		}
	}
	if !changed {
		return data, nil
	}
	result, err := json.Marshal(doc)
	if len(result) > nftMaxOutput {
		return nil, ErrNetworkBudget
	}
	return result, err
}

func nftWireDuration(object map[string]any, field string, changed *bool) error {
	value, present := object[field]
	if !present {
		return nil
	}
	n, ok := value.(json.Number)
	if !ok {
		return ErrNetworkUnverified
	}
	seconds, err := strconv.ParseInt(string(n), 10, 64)
	if err != nil || seconds < 0 || seconds > int64((MaxIsolationDuration+nftLease)/time.Second) {
		return ErrNetworkUnverified
	}
	// Zero stays zero: expired/subsecond leases remain unverified.
	object[field] = json.Number(strconv.FormatInt(seconds*1000, 10))
	*changed = true
	return nil
}
