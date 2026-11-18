package cache

import (
	"encoding/json"
	"errors"
	"time"
)

// This versioned approval identity deliberately lists the immutable fields.
// []byte encodes strings losslessly, including non-UTF-8 Unix path bytes. Adding
// fields to QuarantineRecord must not silently change existing receipt hashes.
type purgeRecordWire struct {
	Version                                     int
	ID, OriginalPath, SHA256, Detection, Engine []byte
	QuarantinedSeconds                          int64
	QuarantinedNanoseconds                      int
	Restorable, Restored                        bool
	CaptureDir, CaptureDirID, FileID, Security  []byte
}

// EncodeQuarantinePurgeRecord preserves an approved record's identity. Purged
// and PurgedAt are excluded because completion changes precisely those fields.
func EncodeQuarantinePurgeRecord(rec QuarantineRecord) ([]byte, error) {
	return json.Marshal(purgeRecordWire{
		Version: 1, ID: []byte(rec.QuarantineID), OriginalPath: []byte(rec.OriginalPath),
		SHA256: []byte(rec.SHA256), Detection: []byte(rec.Detection), Engine: []byte(rec.Engine),
		QuarantinedSeconds: rec.QuarantinedAt.Unix(), QuarantinedNanoseconds: rec.QuarantinedAt.Nanosecond(),
		Restorable: rec.Restorable, Restored: rec.Restored,
		CaptureDir: []byte(rec.SourceCaptureDir), CaptureDirID: []byte(rec.SourceCaptureDirID),
		FileID: []byte(rec.SourceFileID), Security: []byte(rec.OriginalSecurity),
	})
}

func DecodeQuarantinePurgeRecord(data []byte) (QuarantineRecord, error) {
	var rec QuarantineRecord
	if len(data) > 1<<20 {
		return rec, errors.New("purge approval record exceeds its limit")
	}
	var w purgeRecordWire
	if err := json.Unmarshal(data, &w); err != nil {
		return rec, err
	}
	if w.Version != 1 || w.QuarantinedNanoseconds < 0 || w.QuarantinedNanoseconds >= 1e9 {
		return rec, errors.New("invalid purge approval encoding")
	}
	rec = QuarantineRecord{
		QuarantineID: string(w.ID), OriginalPath: string(w.OriginalPath), SHA256: string(w.SHA256),
		Detection: string(w.Detection), Engine: string(w.Engine),
		QuarantinedAt: time.Unix(w.QuarantinedSeconds, int64(w.QuarantinedNanoseconds)).UTC(),
		Restorable:    w.Restorable, Restored: w.Restored,
		SourceCaptureDir: string(w.CaptureDir), SourceCaptureDirID: string(w.CaptureDirID),
		SourceFileID: string(w.FileID), OriginalSecurity: string(w.Security),
	}
	return rec, nil
}
