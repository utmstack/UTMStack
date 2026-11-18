package bytecode

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// ClamAV's JSON API assigns a new positive handle for each successful lookup;
// handle zero is the metadata root. Bound that per-run table independently of
// the VM instruction budget so repeated queries cannot exhaust host memory.
const maxJSONHandles = 1 << 16

func jsonSigned(v int32) uint64 { return uint64(uint32(v)) }

func (s *apiState) jsonNode(id uint64) (any, bool) {
	if s.jsonObjs == nil {
		s.jsonObjs = []any{s.jsonRoot}
	}
	if int32(id) < 0 || uint64(uint32(id)) >= uint64(len(s.jsonObjs)) {
		return nil, false
	}
	return s.jsonObjs[uint32(id)], true
}

func (s *apiState) jsonHandle(value any) uint64 {
	if len(s.jsonObjs) >= maxJSONHandles {
		return apiFailure
	}
	s.jsonObjs = append(s.jsonObjs, value)
	return uint64(len(s.jsonObjs) - 1)
}

func jsonKind(value any) (uint64, bool) {
	switch value.(type) {
	case nil:
		return 0, true
	case bool:
		return 1, true
	case float32, float64:
		return 2, true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return 3, true
	case json.Number:
		if strings.ContainsAny(string(value.(json.Number)), ".eE") {
			return 2, true
		}
		return 3, true
	case map[string]any:
		return 4, true
	case []any:
		return 5, true
	case string:
		return 6, true
	}
	return 0, false
}

// invokeJSON follows the pinned ClamAV 1.5.4 JSON API's wire IDs and handle
// behavior. With metadata collection disabled, ClamAV reports inactive (0)
// and every other JSON query returns -1.
func (s *apiState) invokeJSON(id int, args []uint64, mem *memory, n int, bytesWork func(uint64) error) (uint64, error) {
	if s.jsonRoot == nil {
		// ClamAV answers json_is_active from the scan-wide metadata
		// collection option. A scanner without that option has no JSON
		// tree, so the answer is false even during a scan.
		if id == 92 {
			return 0, nil
		}
		return apiFailure, nil
	}
	if id == 92 {
		return 1, nil
	}
	if id == 93 {
		if err := bytesWork(uint64(n)); err != nil {
			return 0, err
		}
		name, err := mem.read(args[0], n)
		if err != nil {
			return 0, err
		}
		name, _, _ = bytes.Cut(name, []byte{0})
		parent, ok := s.jsonNode(args[2])
		if !ok {
			return apiFailure, nil
		}
		if parent == nil {
			return apiFailure, nil
		}
		object, ok := parent.(map[string]any)
		if !ok {
			return 0, nil
		}
		value, found := object[string(name)]
		if !found {
			if s.jsonPartial {
				return 0, ErrUnsupported
			}
			return 0, nil
		}
		return s.jsonHandle(value), nil
	}
	object, ok := s.jsonNode(args[len(args)-1])
	if !ok {
		return apiFailure, nil
	}
	switch id {
	case 94:
		kind, known := jsonKind(object)
		if !known {
			return 0, ErrUnsupported
		}
		return kind, nil
	case 95:
		array, ok := object.([]any)
		if !ok {
			return jsonSigned(-2), nil
		}
		return uint64(len(array)), nil
	case 96:
		array, ok := object.([]any)
		if !ok {
			return jsonSigned(-2), nil
		}
		index := int32(args[0])
		if index < 0 || int(index) >= len(array) || array[index] == nil {
			return 0, nil
		}
		return s.jsonHandle(array[index]), nil
	case 97, 98:
		value, ok := object.(string)
		if !ok {
			return jsonSigned(-2), nil
		}
		if before, _, found := strings.Cut(value, "\x00"); found {
			value = before
		}
		if id == 97 {
			return uint64(len(value)), nil
		}
		if n == 0 {
			return 0, ErrRuntime
		}
		count := min(len(value), n-1)
		if err := bytesWork(uint64(count + 1)); err != nil {
			return 0, err
		}
		out := make([]byte, count+1)
		copy(out, value[:count])
		if err := mem.write(args[0], out); err != nil {
			return 0, err
		}
		return uint64(count + 1), nil
	case 99:
		value, known := jsonBoolean(object)
		if !known {
			return 0, ErrUnsupported
		}
		if value {
			return 1, nil
		}
		return 0, nil
	case 100:
		value, known := jsonInteger(object)
		if !known {
			return 0, ErrUnsupported
		}
		return jsonSigned(value), nil
	}
	return 0, ErrUnsupported
}

func jsonBoolean(value any) (bool, bool) {
	switch v := value.(type) {
	case nil:
		return false, true
	case bool:
		return v, true
	case string:
		return v != "", true
	case map[string]any:
		return false, true
	case []any:
		return false, true
	case float32:
		return v != 0, true
	case float64:
		return v != 0, true
	case json.Number:
		if strings.ContainsAny(string(v), ".eE") {
			f, err := strconv.ParseFloat(string(v), 64)
			return f != 0, err == nil
		}
		i, err := strconv.ParseInt(string(v), 10, 64)
		return i != 0, err == nil
	default:
		v, ok := jsonInteger(value)
		return v != 0, ok
	}
}

func jsonInteger(value any) (int32, bool) {
	clamp := func(v int64) int32 {
		if v > math.MaxInt32 {
			return math.MaxInt32
		}
		if v < math.MinInt32 {
			return math.MinInt32
		}
		return int32(v)
	}
	switch v := value.(type) {
	case nil:
		return 0, true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	case int:
		return clamp(int64(v)), true
	case int8:
		return int32(v), true
	case int16:
		return int32(v), true
	case int32:
		return v, true
	case int64:
		return clamp(v), true
	case uint, uint8, uint16, uint32, uint64:
		var x uint64
		switch t := v.(type) {
		case uint:
			x = uint64(t)
		case uint8:
			x = uint64(t)
		case uint16:
			x = uint64(t)
		case uint32:
			x = uint64(t)
		case uint64:
			x = t
		}
		if x > math.MaxInt32 {
			return math.MaxInt32, true
		}
		return int32(x), true
	case float32:
		return jsonInteger(float64(v))
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, false
		}
		if v > math.MaxInt32 {
			return math.MaxInt32, true
		}
		if v < math.MinInt32 {
			return math.MinInt32, true
		}
		return int32(v), true
	case json.Number:
		if strings.ContainsAny(string(v), ".eE") {
			f, err := strconv.ParseFloat(string(v), 64)
			if err != nil {
				return 0, false
			}
			return jsonInteger(f)
		}
		return jsonInteger(string(v))
	case string:
		// json-c's integer coercion accepts a decimal prefix (and leading
		// whitespace), stopping at the first non-digit.
		prefix := strings.TrimLeft(v, " \t\n\r\v\f")
		end := 0
		if len(prefix) > 0 && (prefix[0] == '+' || prefix[0] == '-') {
			end++
		}
		startDigits := end
		for end < len(prefix) && prefix[end] >= '0' && prefix[end] <= '9' {
			end++
		}
		if end == startDigits {
			return 0, true
		}
		i, err := strconv.ParseInt(prefix[:end], 10, 64)
		if err != nil {
			if strings.HasPrefix(prefix, "-") {
				return math.MinInt32, true
			}
			return math.MaxInt32, true
		}
		return clamp(i), true
	case map[string]any, []any:
		return 0, true
	}
	return 0, false
}
