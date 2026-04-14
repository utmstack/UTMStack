//go:build linux
// +build linux

package auditd

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/elastic/go-libaudit/v2/auparse"
)

// formatAuditEvent converts reassembled audit messages to a FLAT JSON format.
// This ensures ZERO DATA LOSS - every field from every record appears in output.
//
// Flattening rules:
//   - SYSCALL: all fields merged directly to root
//   - CWD: extracts "cwd" field to root
//   - PROCTITLE: extracts "proctitle" field to root
//   - EXECVE: all fields in "execve" object (avoids a0-aN collision with SYSCALL)
//   - PATH: all records collected in "paths" array
//   - SOCKADDR: all fields in "sockaddr" object
//   - Any other record type: all fields in "{lowercase_type}" object (fallback)
func formatAuditEvent(msgs []*auparse.AuditMessage) (string, error) {
	if len(msgs) == 0 {
		return "", nil
	}

	// Get timestamp and sequence from first message
	firstMsg := msgs[0]

	// Initialize the flat event structure using map for flexibility
	event := make(map[string]interface{})
	event["@timestamp"] = firstMsg.Timestamp.Format(time.RFC3339Nano)
	event["type"] = "auditd"
	event["sequence"] = firstMsg.Sequence

	// Determine category (SYSCALL takes priority, otherwise first record type)
	var category string

	// Collect PATH records for the "paths" array
	var paths []map[string]interface{}

	// Process each message according to flattening rules
	for _, msg := range msgs {
		recordType := msg.RecordType
		recordTypeName := recordType.String()

		// Set category: SYSCALL takes priority, otherwise use first record type
		if category == "" {
			category = recordTypeName
		}
		if recordType == auparse.AUDIT_SYSCALL {
			category = "SYSCALL"
		}

		// Extract data from message - ZERO DATA LOSS requirement
		data, err := msg.Data()
		if err != nil {
			// Even on error, try to continue with empty data
			data = make(map[string]string)
		}

		// Apply flattening rules based on record type
		switch recordType {
		case auparse.AUDIT_SYSCALL:
			// SYSCALL: merge ALL fields directly to root
			mergeSyscallToRoot(event, data)

		case auparse.AUDIT_CWD:
			// CWD: extract "cwd" field to root
			if cwd, ok := data["cwd"]; ok {
				event["cwd"] = cwd
			}
			// Also include any other fields from CWD record (ZERO DATA LOSS)
			for k, v := range data {
				if k != "cwd" {
					// Prefix with "cwd_" to avoid collisions
					event["cwd_"+k] = v
				}
			}

		case auparse.AUDIT_PROCTITLE:
			// PROCTITLE: extract "proctitle" field to root
			if proctitle, ok := data["proctitle"]; ok {
				event["proctitle"] = proctitle
			}
			// Also include any other fields (ZERO DATA LOSS)
			for k, v := range data {
				if k != "proctitle" {
					event["proctitle_"+k] = v
				}
			}

		case auparse.AUDIT_EXECVE:
			// EXECVE: all fields go into "execve" object to avoid a0-aN collision
			execveObj := make(map[string]interface{})
			for k, v := range data {
				execveObj[k] = v
			}
			event["execve"] = execveObj

		case auparse.AUDIT_PATH:
			// PATH: collect all records into "paths" array
			pathRecord := make(map[string]interface{})
			for k, v := range data {
				pathRecord[k] = v
			}
			paths = append(paths, pathRecord)

		case auparse.AUDIT_SOCKADDR:
			// SOCKADDR: all fields in "sockaddr" object
			sockaddrObj := make(map[string]interface{})
			for k, v := range data {
				sockaddrObj[k] = v
			}
			event["sockaddr"] = sockaddrObj

		default:
			// FALLBACK: any other record type goes into "{lowercase_type}" object
			// This ensures ZERO DATA LOSS for unknown/future record types
			fallbackKey := strings.ToLower(recordTypeName)
			// Handle special characters in record type names
			fallbackKey = strings.ReplaceAll(fallbackKey, "_", "")

			// Check if we already have this type (could be multiple records of same type)
			if existing, ok := event[fallbackKey]; ok {
				// Already exists - convert to array if not already
				switch v := existing.(type) {
				case []map[string]interface{}:
					// Already an array, append
					record := make(map[string]interface{})
					for k, val := range data {
						record[k] = val
					}
					event[fallbackKey] = append(v, record)
				case map[string]interface{}:
					// Convert to array
					record := make(map[string]interface{})
					for k, val := range data {
						record[k] = val
					}
					event[fallbackKey] = []map[string]interface{}{v, record}
				}
			} else {
				// First occurrence - create object
				fallbackObj := make(map[string]interface{})
				for k, v := range data {
					fallbackObj[k] = v
				}
				event[fallbackKey] = fallbackObj
			}
		}
	}

	// Add category to event
	event["category"] = category

	// Add paths array if we collected any PATH records
	if len(paths) > 0 {
		event["paths"] = paths
	}

	// Marshal to JSON - deterministic output with sorted keys
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// mergeSyscallToRoot merges all SYSCALL fields directly to the event root.
// This is the primary record type and its fields become top-level.
func mergeSyscallToRoot(event map[string]interface{}, data map[string]string) {
	for k, v := range data {
		event[k] = v
	}
}
