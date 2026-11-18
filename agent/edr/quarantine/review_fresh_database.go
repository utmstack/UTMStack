package quarantine

import (
	"errors"
	"path/filepath"
)

// SQLite must not interpret a retained payload, mutable journal, owner ticket
// or published output as its database. Keep the ordinary configured database
// distinct by actual file/directory identities, including hard-link aliases.
func (p *freshRestorePermit) pinDatabase(path string, m restoreManifest) (string, func() error, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}
	if err = freshLocalDestination(absolute); err != nil {
		return "", nil, err
	}
	parent, err := platformRestoreDirectory(filepath.Dir(absolute))
	if err != nil {
		return "", nil, err
	}
	f, err := pinFreshDatabaseFile(parent, filepath.Base(absolute))
	if err != nil {
		parent.Close()
		return "", nil, err
	}
	release := func() error { return errors.Join(f.Close(), parent.Close()) }
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		release()
		return "", nil, errors.Join(err, errors.New("fresh metadata database must already be a regular file"))
	}
	id, err := admissionFileID(f)
	if err != nil {
		release()
		return "", nil, err
	}
	a := p.approval
	old, err := decodeRestoreManifest(a.Inspection.Original.Manifest, a.QuarantineID)
	if err != nil {
		release()
		return "", nil, err
	}
	// The mutable v2 journal is replaced as phases advance, so its current
	// file identity must be read from the exact bound operation. Its parent
	// check alone cannot detect a hard-link alias outside private storage.
	var journalID string
	if p.owner.OperationDirectoryID != "" {
		op, _, e := openFreshOwnedOperation(p.root, p.decision, p.owner)
		if e != nil {
			release()
			return "", nil, e
		}
		entry, e := readFreshControl(op, "manifest.json", maxAdmissionJournal)
		e = errors.Join(e, op.Close())
		if e != nil {
			release()
			return "", nil, e
		}
		journalID = entry.ManifestID
	}
	for _, protected := range []string{journalID, a.FileID, a.Inspection.Snapshot.FileID, a.Inspection.Hold.ManifestID, a.Inspection.Original.ManifestID, a.Inspection.Cancellation.ManifestID, old.Record.SourceFileID, old.OutputID, p.owner.FileID, p.owner.TicketID, m.OutputID} {
		if protected != "" && id == protected {
			release()
			return "", nil, errors.New("fresh metadata database aliases protected recovery evidence")
		}
	}
	directories := []string{a.QueueID, a.ItemID, a.DirectoryID, a.Inspection.Hold.DirectoryID, a.Inspection.Cancellation.DirectoryID, a.Inspection.Original.DirectoryID, old.Record.SourceCaptureDirID, old.StageID, p.owner.OperationDirectoryID, p.owner.StageID}
	for current := parent.path; ; current = filepath.Dir(current) {
		d, e := platformRestoreDirectory(current)
		if e != nil {
			release()
			return "", nil, e
		}
		dirID, e := admissionFileID(d.file)
		e = errors.Join(e, d.Close())
		if e != nil {
			release()
			return "", nil, e
		}
		for _, protected := range directories {
			if protected != "" && dirID == protected {
				release()
				return "", nil, errors.New("fresh metadata database is inside a protected recovery directory")
			}
		}
		if filepath.Dir(current) == current {
			break
		}
	}
	return absolute, release, nil
}
